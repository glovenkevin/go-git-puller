package utils

func CheckIsError(err error) bool {
	if err != nil && Verbose {
		logs.Sugar().Warnf(err.Error())
		return true
	}
	return false
}

// Check if given string not contain any special character or number
func CheckIsStringOnly(value string) bool {
	return StringOnlyPattern.Match([]byte(value))
}

// Print debug message if the verbose flag being set true
func Debug(input ...interface{}) {
	if Verbose {
		logs.Sugar().Info(input...)
	}
}

// Print debug message with format and if the verbose flag being set true
func Debugf(str string, param ...interface{}) {
	if Verbose {
		logs.Sugar().Infof(str, param...)
	}
}

func Warn(str string) {
	logs.Warn(str)
}

func Warnf(str string, args ...interface{}) {
	logs.Sugar().Warnf(str, args...)
}

func Error(str string) {
	logs.Error(str)
}

func Errorf(str string, args ...interface{}) {
	logs.Sugar().Errorf(str, args...)
}

func Panic(str string) {
	logs.Panic(str)
}

func Panicf(str string, args ...interface{}) {
	logs.Sugar().Panicf(str, args...)
}
