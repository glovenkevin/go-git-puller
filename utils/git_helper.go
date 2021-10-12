package utils

import (
	"github.com/go-git/go-git/v5"
	"go.uber.org/zap"
)

// Checking current directory path given is a
// git repository
func CheckDirIsGitRepo(path string, logs *zap.Logger) bool {
	_, err := git.PlainOpen(path)
	return err == nil
}
