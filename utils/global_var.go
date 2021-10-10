package utils

import "go.uber.org/zap"

var Verbose bool
var RootDir string
var Environtment string
var HardReset bool
var Action string

var Username string
var Password string

const PATERN_STRING_ONLY string = "^[a-zA-Z]+$"

var Logs *zap.Logger

func init() {
	Verbose = false
	RootDir = "."
	Environtment = "PROD"
	HardReset = false
}
