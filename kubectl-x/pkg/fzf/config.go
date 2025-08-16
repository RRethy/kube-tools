package fzf

type Config struct {
	ExactMatch bool
	Sorted     bool
	Multi      bool
	Prompt     string
	Query      string
}

func (c Config) buildArgs() []string {
	args := []string{
		"--height",
		"30%",
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
	return args
}
