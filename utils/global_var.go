package utils

import (
	"go.uber.org/zap"
)

// Flag for verbosing the action
// produced by application
var Verbose bool

// Define root directory of the action
var RootDir string

// Set repository for using
// hard or softreset
var HardReset bool

// Define action for command line
// Current:
//		update
//		update-gitlab (To-Be)
var Action string

// Define auth AuthMethod
// Supported:
//		http (basic auth http using username and password)
//		token (using token from your git repository)
var Auth string

// Define base url for gitlab or github repository
var Baseurl string

// Username used for login
// in gitlab or other git repository
var Username string

// Password used for login
// in gitlab or other git repository
var Password string

// Token from git service for authentication
var Token string

// Define const pattern for regex
// String only validation
const PATERN_STRING_ONLY string = "^[a-zA-Z]+$"

// Init Global Logger for the application
var logs *zap.Logger

func init() {
	Verbose = false
	RootDir = "."
	HardReset = false

	logs, _ = zap.NewProduction()
}
