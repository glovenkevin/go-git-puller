package utils

func ErrorHandler() bool {
	if err := recover(); err != nil {
		Logs.Sugar().Error(err)
		return false
	}
	return true
}
