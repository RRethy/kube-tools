package each

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/fatih/color"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"sigs.k8s.io/yaml"

	"github.com/RRethy/kubectl-x/pkg/fzf"
	"github.com/RRethy/kubectl-x/pkg/kubeconfig"
)

type Eacher struct {
	IOStreams  genericclioptions.IOStreams
	Kubeconfig kubeconfig.Interface
	Namespace  string
	Fzf        fzf.Interface
}

type clusterResult struct {
	Context string `json:"context" yaml:"context"`
	Args    string `json:"command" yaml:"command"`
	Output  any    `json:"output,omitempty" yaml:"output,omitempty"`
	Error   string `json:"error,omitempty" yaml:"error,omitempty"`
}

func (e *Eacher) Each(ctx context.Context, contextPattern, outputFormat string, interactive bool, args []string) error {
	if !slices.Contains([]string{"yaml", "json", "raw"}, outputFormat) {
		return fmt.Errorf("invalid output format: %s (must be yaml, json, or raw)", outputFormat)
	}

	contexts, err := e.selectContexts(ctx, contextPattern, interactive)
	if err != nil {
		return fmt.Errorf("selecting contexts: %w", err)
	}

	if len(contexts) == 0 {
		return fmt.Errorf("no contexts matched pattern: %s", contextPattern)
	}

	results := e.executeCommands(ctx, contexts, outputFormat, args)

	return e.outputResults(results, outputFormat)
}

func (e *Eacher) selectContexts(ctx context.Context, pattern string, interactive bool) ([]string, error) {
	allContexts := e.Kubeconfig.Contexts()
	filteredContexts, err := e.selectContextsByPattern(allContexts, pattern)
	if err != nil {
		return nil, fmt.Errorf("selecting contexts by pattern: %w", err)
	}

	if interactive {
		return e.selectContextsInteractively(ctx, filteredContexts, pattern)
	}
	return filteredContexts, nil
}

func (e *Eacher) selectContextsInteractively(ctx context.Context, allContexts []string, initialPattern string) ([]string, error) {
	fzfCfg := fzf.Config{ExactMatch: false, Sorted: true, Multi: true, Prompt: "Select context", Query: initialPattern}
	selectedContexts, err := e.Fzf.Run(ctx, allContexts, fzfCfg)
	if err != nil {
		return nil, fmt.Errorf("interactive selection: %w", err)
	}

	return selectedContexts, nil
}

func (e *Eacher) selectContextsByPattern(allContexts []string, pattern string) ([]string, error) {
	if pattern == "" {
		return allContexts, nil
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	var matched []string
	for _, ctx := range allContexts {
		if re.MatchString(ctx) {
			matched = append(matched, ctx)
		}
	}

	return matched, nil
}

func (e *Eacher) executeCommands(ctx context.Context, contexts []string, outputFormat string, args []string) []clusterResult {
	var wg sync.WaitGroup
	sem := make(chan struct{}, 32)
	results := make(chan clusterResult, len(contexts))

	for _, context := range contexts {
		sem <- struct{}{}
		wg.Add(1)
		go func(contextName string) {
			defer func() {
				wg.Done()
				<-sem
			}()

			result := e.executeCommand(ctx, contextName, outputFormat, args)
			results <- result
		}(context)
	}

	go func() {
		wg.Wait()
		close(sem)
		close(results)
	}()

	var items []clusterResult
	for result := range results {
		items = append(items, result)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Context < items[j].Context
	})

	return items
}

func (e *Eacher) executeCommand(ctx context.Context, contextName string, outputFormat string, args []string) clusterResult {
	cmdArgs := append(args, "--context", contextName)
	cmdArgs = append(cmdArgs, "--namespace", e.Namespace)

	if outputFormat != "raw" {
		cmdArgs = append(cmdArgs, fmt.Sprintf("-o%s", outputFormat))
	}

	result := clusterResult{
		Context: contextName,
		Args:    strings.Join(cmdArgs, " "),
	}

	cmd := exec.CommandContext(ctx, "kubectl", cmdArgs...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		result.Error = err.Error()
		result.Output = strings.TrimSpace(string(output))
	} else if outputFormat == "raw" {
		result.Output = strings.TrimSpace(string(output))
	} else {
		var yamlOutput map[string]any
		if err := yaml.Unmarshal(output, &yamlOutput); err == nil {
			result.Output = yamlOutput
		} else {
			result.Output = strings.TrimSpace(string(output))
		}
	}

	return result
}

func (e *Eacher) outputResults(items []clusterResult, outputFormat string) error {
	switch outputFormat {
	case "json":
		return e.outputJSON(items)
	case "yaml":
		return e.outputYAML(items)
	case "raw":
		return e.outputRaw(items)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}

func (e *Eacher) outputJSON(items []clusterResult) error {
	output := map[string]any{
		"apiVersion": "v1",
		"kind":       "List",
		"items":      items,
	}

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling JSON: %w", err)
	}

	_, err = fmt.Fprintln(e.IOStreams.Out, string(jsonBytes))
	return err
}

func (e *Eacher) outputYAML(items []clusterResult) error {
	output := map[string]any{
		"apiVersion": "v1",
		"kind":       "List",
		"items":      items,
	}

	yamlBytes, err := yaml.Marshal(output)
	if err != nil {
		return fmt.Errorf("marshalling YAML: %w", err)
	}

	_, err = fmt.Fprint(e.IOStreams.Out, string(yamlBytes))
	return err
}

func (e *Eacher) outputRaw(items []clusterResult) error {
	for _, result := range items {
		contextColor := color.New(color.FgCyan, color.Bold)
		contextColor.Fprintf(e.IOStreams.Out, "%s:\n", result.Context)

		if result.Error != "" {
			errorColor := color.New(color.FgRed)
			errorColor.Fprintf(e.IOStreams.Out, "  Error: %s\n", result.Error)
		}

		if outStr, ok := result.Output.(string); ok {
			for line := range strings.SplitSeq(outStr, "\n") {
				if line != "" {
					fmt.Fprintf(e.IOStreams.Out, "  %s\n", line)
				}
			}
		} else if result.Output != nil {
			fmt.Fprintf(e.IOStreams.Out, "  %v\n", result.Output)
		}
	}
	return nil
}
