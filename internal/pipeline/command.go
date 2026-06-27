package pipeline

import (
	"sort"
	"strings"
)

// CommandLine returns a shell pipeline equivalent to the enabled transform
// chain. Input is intentionally omitted so callers can pipe files, stdin, or
// other command output into the recipe.
func (p *Pipeline) CommandLine() string {
	var commands []string
	for _, step := range p.steps {
		if step.Disabled || step.Plugin == "" {
			continue
		}
		cmd := "deen "
		if step.Unprocess {
			cmd += "."
		}
		cmd += shellQuote(step.Plugin)
		for _, name := range sortedOptionNames(step.Options) {
			value := step.Options[name]
			if value == "" || value == "false" {
				continue
			}
			cmd += " -" + shellQuote(name)
			if value != "true" {
				cmd += " " + shellQuote(value)
			}
		}
		commands = append(commands, cmd)
	}
	return strings.Join(commands, " | ")
}

func sortedOptionNames(options map[string]string) []string {
	names := make([]string, 0, len(options))
	for name := range options {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	if isShellSafe(s) {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}

func isShellSafe(s string) bool {
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			continue
		}
		switch r {
		case '_', '-', '.', '/', ':', '=', ',':
			continue
		default:
			return false
		}
	}
	return true
}
