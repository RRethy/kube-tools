package fzf

const defaultHeight = "30%"

// Config holds configuration options for fzf execution
type Config struct {
	ExactMatch bool   // ExactMatch enables exact matching instead of fuzzy
	Sorted     bool   // Sorted preserves the original order of items
	Multi      bool   // Multi allows selecting multiple items
	Prompt     string // Prompt sets the prompt text
	Query      string // Query sets the initial search query
	Preview    string // Preview command to show additional information about selected items
	Height     string // Height of the fzf window, e.g., "30%" or "50%"
}

func (c Config) buildArgs() []string {
	height := c.Height
	if height == "" {
		height = defaultHeight
	}

	args := []string{
		"--height",
		height,
		"--ansi",
		"--select-1",
		"--exit-0",
		"--color=dark",
		"--layout=reverse",
	}
	if c.ExactMatch {
		args = append(args, "--exact")
	}
	if c.Multi {
		args = append(args, "--multi")
	}
	if c.Prompt != "" {
		args = append(args, "--prompt", c.Prompt+"> ")
	}
	if c.Query != "" {
		args = append(args, "--query", c.Query)
	}
	if c.Preview != "" {
		args = append(args, "--preview", c.Preview)
	}
	return args
}
