package utils

import (
	"os/exec"
	"strings"

	"go.uber.org/zap"
)

// Checking current directory path given is a
// git repository
func CheckDirIsGitRepo(path string, logs *zap.Logger) bool {

	cmdShell := exec.Command("git", "-C", path, "rev-parse", "--is-inside-work-tree")
	result, err := cmdShell.Output()

	if CheckIsError(err, logs) {
		return false
	}

	stdout := string(result)
	return strings.Contains(stdout, "true")
}
