package utils

import (
	"github.com/go-git/go-git/v5"
)

// Checking current directory path given is a
// git repository
func CheckDirIsGitRepo(path string) bool {
	_, err := git.PlainOpen(path)
	return err == nil
}
