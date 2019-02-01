package shell

const (
	shellPath            = "cmd.exe"
	shellCmdPrefix       = "/C"
	winCodePageFixPrefix = "chcp 65001 > nil" // Force use UTF-8 to provide correct output to stdout
)

func wrapCommand(cmd string) string {
	return winCodePageFixPrefix + " && " + cmd
}
