package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-gilbert/gilbert/support/fs"
	"github.com/go-gilbert/gilbert/v2/manifest"
	"github.com/spf13/cobra"
	"github.com/zclconf/go-cty/cty"
)

type CommandHandler = func(c *cobra.Command, args []string)

func FindManifest() (string, bool, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", false, fmt.Errorf("failed to get working directory: %s", err)
	}

	manPath, found, err := fs.Lookup(manifest.DefaultFileName, wd, 3)
	if err != nil {
		return "", false, fmt.Errorf("failed to find file %q: %s", manifest.DefaultFileName, err.Error())
	}

	return manPath, found, nil
}

func ExitWithError(msg string, args ...interface{}) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	_, _ = fmt.Fprintln(os.Stderr, "error: ", msg)
	os.Exit(1)
}

func LoadManifest(parentCmd *cobra.Command, manPath string) (*manifest.Manifest, error) {
	// TODO: add default context
	m, err := manifest.FromFile(manPath, nil)
	if err != nil {
		return nil, err
	}

	// populate manifest tasks into command
	for _, task := range m.Tasks {
		cmd, err := taskToCommand(task)
		if err != nil {
			return nil, fmt.Errorf("loadManifest(): %s", err)
		}

		parentCmd.AddCommand(cmd)
	}

	return m, nil
}

func PrintManifestCommandHandler(m *manifest.Manifest, asJson bool) CommandHandler {
	return func(c *cobra.Command, args []string) {
		PrintManifestSummary(m, asJson)
	}
}

func PrintManifestSummary(m *manifest.Manifest, asJson bool) {
	if asJson {
		// TODO: add JSON output
		panic("not implemented")
	}

	fmt.Printf("List of tasks defined in %q\n\n", m.FilePath())
	for _, task := range m.Tasks {
		fmt.Printf("- %q\n", task.Name)

		if task.HasDescription() {
			fmt.Println("  Description:", task.Description)
		}

		if task.HasParameters() {
			fmt.Println("  Parameters:")
			for _, param := range task.Parameters {
				fmt.Printf("  * %q\t(%s)\t%s\n", param.Name, param.Type.FriendlyName(), param.Description)
			}
		}
		fmt.Printf("\n")
	}
}

func taskToCommand(task manifest.Task) (*cobra.Command, error) {
	c := &cobra.Command{
		Use:   task.Name,
		Short: task.Description,
		Long:  getTaskSummary(task),
	}

	if len(task.Parameters) == 0 {
		return c, nil
	}

	args := c.PersistentFlags()
	for name, param := range task.Parameters {
		switch param.Type {
		case cty.String:
			defVal := ""
			if param.HasDefaultValue() {
				defVal = param.DefaultValue.AsString()
			}

			args.String(name, defVal, param.Description)
		case cty.Number:
			var defVal float32
			if param.HasDefaultValue() {
				defVal, _ = param.DefaultValue.AsBigFloat().Float32()
			}

			args.Float32(name, defVal, param.Description)
		case cty.Bool:
			var defVal bool
			if param.HasDefaultValue() {
				defVal = param.DefaultValue.True()
			}

			args.Bool(name, defVal, param.Description)
		default:
			return nil, fmt.Errorf(
				"unsupported parameter %q type %q in task %q",
				name, param.Type.FriendlyName(), task.Name,
			)
		}
	}

	return c, nil
}

func getTaskSummary(t manifest.Task) string {
	str := fmt.Sprintf("Run %q task\n", t.Name)

	if t.HasDescription() {
		str += fmt.Sprintln("\nDescription:\n", t.Description)
	}

	if t.HasParameters() {
		sb := strings.Builder{}
		sb.WriteString(str)
		sb.WriteString("\nParameters:\n")
		for _, param := range t.Parameters {
			sb.WriteString("* ")
			sb.WriteString(param.Name)
			sb.WriteRune('\t')
			sb.WriteString(param.Type.FriendlyName())
			sb.WriteRune('\t')
			sb.WriteString(param.Description)
			if !param.HasDefaultValue() {
				sb.WriteString(" (required)")
			}
			sb.WriteRune('\n')
		}

		return sb.String()
	}

	return str
}
