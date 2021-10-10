package utils

import "go.uber.org/zap"

func CheckIsError(err error, logs *zap.Logger) bool {
	if err != nil && Verbose {
		logs.Sugar().Warnf(err.Error())
		return true
	}
	return false
}

func CheckIsStringOnly(value string) bool {
	return StringOnlyPattern.Match([]byte(value))
}

func Debug(input ...interface{}) {
	if Verbose {
		Logs.Sugar().Info(input...)
	}
}

func Debugf(str string, param ...interface{}) {
	if Verbose {
		Logs.Sugar().Infof(str, param...)
	}
}
