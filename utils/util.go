package utils

import "go.uber.org/zap"

func CheckIsError(err error, logs *zap.Logger) bool {
	if err != nil && Verbose {
		logs.Sugar().Warnf(err.Error())
		return true
	}
	return false
}

// Chec if given string not contain any special character or number
func CheckIsStringOnly(value string) bool {
	return StringOnlyPattern.Match([]byte(value))
}

// Print debug message if the verbose flag being set true
func Debug(input ...interface{}) {
	if Verbose {
		Logs.Sugar().Info(input...)
	}
}

// Print debug message with format and if the verbose flag being set true
func Debugf(str string, param ...interface{}) {
	if Verbose {
		Logs.Sugar().Infof(str, param...)
	}
}
