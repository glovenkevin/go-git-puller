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

// Validation for checking if action
// being used is exist
func ValidateEnvirontment() bool {
	validator := setArgsValidator()

	if Action == "" {
		Logs.Sugar().Error("Action can't be empty")
		return false
	}

	if Auth == "" {
		Logs.Sugar().Error("Auth can't be empty")
		return false
	}

	if validator[Action] == nil && Action != "" {
		Logs.Sugar().Errorf("Action is unknown [ %v ]", Action)
		return false
	}

	return validator[Action]()
}

// Put the validation for every action.
// Being separated for better maintenance
func setArgsValidator() map[string]func() bool {
	rtn := make(map[string]func() bool)

	rtn["update"] = validateUpdateAction

	return rtn
}

// Validate for update action needs
func validateUpdateAction() bool {
	if Auth == "http" {
		if Username == "" {
			Logs.Error("Username can't be blank")
			return false
		}

		if Password == "" {
			Logs.Error("Password can't be blank")
			return false
		}
	}

	if Auth == "token" && Token == "" {
		Logs.Error("Token can't be blank")
		return false
	}

	folder, err := os.Stat(RootDir)
	if err != nil || !folder.IsDir() {
		Logs.Error("Directory not exist or isn't a directory")
		return false
	}

	return true
}
