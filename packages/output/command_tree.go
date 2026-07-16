package output

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CommandFlag represents a single flag/option for a command.
type CommandFlag struct {
	Name           string   `json:"name"`
	Shorthand      string   `json:"shorthand,omitempty"`
	Description    string   `json:"description"`
	DefaultValue   string   `json:"default_value,omitempty"`
	Required       bool     `json:"required"`
	Type           string   `json:"type"`
	AcceptedValues []string `json:"accepted_values,omitempty"`
}

// CommandNode represents a single command in the command tree.
type CommandNode struct {
	Use         string         `json:"use"`
	Short       string         `json:"short"`
	Long        string         `json:"long,omitempty"`
	Aliases     []string       `json:"aliases,omitempty"`
	Example     string         `json:"example,omitempty"`
	Subcommands []*CommandNode `json:"subcommands,omitempty"`
	Flags       []CommandFlag  `json:"flags,omitempty"`
	Args        string         `json:"args,omitempty"`
	Deprecated  string         `json:"deprecated,omitempty"`
}

// CommandTree is the top-level wrapper for the CLI command tree structure.
type CommandTree struct {
	CLIName string       `json:"cli_name"`
	Version string       `json:"version,omitempty"`
	Root    *CommandNode `json:"root_command"`
}

// BuildCommandTree recursively builds a command tree from a cobra root command.
func BuildCommandTree(root *cobra.Command) *CommandTree {
	tree := &CommandTree{
		CLIName: root.Name(),
		Root:    buildNode(root),
	}

	// Extract version if available
	if root.Version != "" {
		tree.Version = root.Version
	}

	return tree
}

// buildNode recursively converts a cobra.Command into a CommandNode.
func buildNode(cmd *cobra.Command) *CommandNode {
	node := &CommandNode{
		Use:        cmd.Use,
		Short:      cmd.Short,
		Long:       cmd.Long,
		Aliases:    cmd.Aliases,
		Example:    cmd.Example,
		Args:       describeArgs(cmd),
		Deprecated: cmd.Deprecated,
	}

	// Build flags
	node.Flags = buildFlags(cmd)

	// Build subcommands
	if cmd.HasSubCommands() {
		for _, sub := range cmd.Commands() {
			// Skip the built-in help command if it has no subcommands
			if sub.Name() == "help" && !sub.HasSubCommands() {
				continue
			}
			node.Subcommands = append(node.Subcommands, buildNode(sub))
		}
		// Sort subcommands alphabetically for consistent output
		sort.Slice(node.Subcommands, func(i, j int) bool {
			return node.Subcommands[i].Use < node.Subcommands[j].Use
		})
	}

	return node
}

// buildFlags extracts all flags from a cobra command, including persistent flags.
func buildFlags(cmd *cobra.Command) []CommandFlag {
	seen := make(map[string]bool)
	var flags []CommandFlag

	addFlags := func(flagSet *pflag.FlagSet) {
		if flagSet == nil {
			return
		}
		flagSet.VisitAll(func(f *pflag.Flag) {
			if f.Hidden {
				return
			}
			// Deduplicate by flag name (persistent flags may overlap with local flags)
			if seen[f.Name] {
				return
			}
			seen[f.Name] = true

			cf := CommandFlag{
				Name:         f.Name,
				Shorthand:    f.Shorthand,
				Description:  f.Usage,
				DefaultValue: f.DefValue,
				Type:         f.Value.Type(),
			}

			// Check if flag is required via annotation
			if f.Annotations != nil {
				if vals, ok := f.Annotations[cobra.BashCompOneRequiredFlag]; ok {
					for _, v := range vals {
						if v == "true" {
							cf.Required = true
							break
						}
					}
				}
				// Extract accepted values from custom bash completion
				if vals, ok := f.Annotations[cobra.BashCompCustom]; ok {
					cf.AcceptedValues = vals
				}
			}

			flags = append(flags, cf)
		})
	}

	// Local flags
	addFlags(cmd.Flags())

	// Persistent flags
	addFlags(cmd.PersistentFlags())

	// Sort flags by name for consistent output
	sort.Slice(flags, func(i, j int) bool {
		return flags[i].Name < flags[j].Name
	})

	return flags
}

// describeArgs returns a human-readable description of the command's expected arguments.
func describeArgs(cmd *cobra.Command) string {
	if len(cmd.ValidArgs) > 0 {
		return fmt.Sprintf("[%s]", strings.Join(cmd.ValidArgs, "|"))
	}

	// Use the positional argument information
	if cmd.Args != nil {
		// We can't compare function values, so we describe based on Use string
		if cmd.Use == cmd.Name() {
			return "none"
		}
	}

	return ""
}

// MarshalCommandTreeJSON returns the command tree as a JSON byte slice.
func MarshalCommandTreeJSON(tree *CommandTree) ([]byte, error) {
	return json.MarshalIndent(tree, "", "  ")
}

// WriteCommandTreeJSON writes the command tree as pretty-printed JSON to the given writer.
func WriteCommandTreeJSON(tree *CommandTree, w interface{ Write([]byte) (int, error) }) error {
	data, err := MarshalCommandTreeJSON(tree)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}
