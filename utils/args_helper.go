package utils

import (
	"flag"
	"fmt"
	"os"
)

// Add custom the -h / help flags message
func SetUsageFlag() {
	var usage = func() {
		fmt.Fprintln(os.Stderr, "Usage Command go-git-pull [-c/-action=action] [-option=option ...]")
		fmt.Fprintln(os.Stderr, "\tAction: update, *update-gitlab, *pull-gitlab")

		fmt.Fprintln(os.Stderr, "\nList option available: ")
		flag.CommandLine.PrintDefaults()
	}
	flag.Usage = usage
}

// Define all args flags and then parse it
func ParseArgs() {
	flag.StringVar(&Action, "c", ".", "Set action")
	flag.StringVar(&Action, "action", ".", "Set action")

	flag.StringVar(&Auth, "m", "http", "Set auth method, use token or http basic (default http)")
	flag.StringVar(&Auth, "auth", "http", "Set auth method, use token or http basic (default http)")

	flag.StringVar(&Token, "t", "", "Token for authentication")
	flag.StringVar(&Token, "token", "", "Token for authentication")

	flag.StringVar(&Baseurl, "u", "", "Url gitlab repository")
	flag.StringVar(&Baseurl, "url", "", "Url gitlab repository")

	flag.StringVar(&Username, "U", "", "Put username for git")
	flag.StringVar(&Username, "username", "", "Put username for git")

	flag.StringVar(&Password, "P", "", "Put password for git")
	flag.StringVar(&Password, "password", "", "Put password for git")

	flag.StringVar(&RootDir, "path", ".", "Set Working directory root path")
	flag.BoolVar(&Verbose, "verbose", false, "Activate verbose/debug print")
	flag.BoolVar(&HardReset, "hard-reset", false, "Set false to use softreset or otherwise")

	flag.Parse()
}

// Validation for checking if action
// being used is exist
func ValidateEnvirontment() bool {
	validator := setArgsValidator()

	if Action == "" {
		Error("Action can't be empty")
		return false
	}

	if Auth == "" {
		Error("Auth can't be empty")
		return false
	}

	if validator[Action] == nil && Action != "" {
		Errorf("Action is unknown [ %v ]", Action)
		return false
	}

	return validator[Action]()
}

// Put the validation for every action.
// Being separated for better maintenance
func setArgsValidator() map[string]func() bool {
	rtn := make(map[string]func() bool)

	rtn["update"] = validateUpdateAction
	rtn["update-gitlab"] = validateUpdateGitlab

	return rtn
}

// Validate for update action needs
func validateUpdateAction() bool {
	if Auth == "http" {
		if Username == "" {
			Error("Username can't be blank")
			return false
		}

		if Password == "" {
			Error("Password can't be blank")
			return false
		}
	}

	if Auth == "token" && Token == "" {
		Error("Token can't be blank")
		return false
	}

	return ValidateFolder(RootDir)
}

func validateUpdateGitlab() bool {
	if Token == "" {
		Error("Token can't be blank")
		return false
	}

	if Baseurl == "" {
		Warn("Default https://gitlab.com is being used")
	}

	return ValidateFolder(RootDir)
}

func ValidateFolder(path string) bool {

	folder, err := os.Stat(path)
	if err != nil || !folder.IsDir() {
		Debugf("Directory not exist or isn't a directory")
		return false
	}

	return true
}
