package shell

const (
	shellPath      = "/bin/sh"
	shellCmdPrefix = "-c"
)

func wrapCommand(cmd string) string {
	return cmd
}
