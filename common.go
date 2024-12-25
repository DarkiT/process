package process

import (
	"runtime"

	"github.com/darkit/process/utils"
)

func getShell() string {
	switch runtime.GOOS {
	case "windows":
		return utils.SearchBinary("cmd.exe")
	default:
		// Check the default binary storage path.
		if utils.Exists("/bin/bash") {
			return "/bin/bash"
		}
		if utils.Exists("/bin/sh") {
			return "/bin/sh"
		}
		// Else search the env PATH.
		path := utils.SearchBinary("bash")
		if path == "" {
			path = utils.SearchBinary("sh")
		}
		return path
	}
}

func getShellOption() string {
	switch runtime.GOOS {
	case "windows":
		return "/c"
	default:
		return "-c"
	}
}

func parseCommand(cmd string) (args []string) {
	if runtime.GOOS != "windows" {
		return []string{cmd}
	}
	// Just for "cmd.exe" in windows.
	var argStr string
	var firstChar, prevChar, lastChar1, lastChar2 byte
	array := utils.SplitAndTrim(cmd, " ")
	for _, v := range array {
		if len(argStr) > 0 {
			argStr += " "
		}
		firstChar = v[0]
		lastChar1 = v[len(v)-1]
		lastChar2 = 0
		if len(v) > 1 {
			lastChar2 = v[len(v)-2]
		}
		if prevChar == 0 && (firstChar == '"' || firstChar == '\'') {
			// It should remove the first quote char.
			argStr += v[1:]
			prevChar = firstChar
		} else if prevChar != 0 && lastChar2 != '\\' && lastChar1 == prevChar {
			// It should remove the last quote char.
			argStr += v[:len(v)-1]
			args = append(args, argStr)
			argStr = ""
			prevChar = 0
		} else if len(argStr) > 0 {
			argStr += v
		} else {
			args = append(args, v)
		}
	}
	return
}
