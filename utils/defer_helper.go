package utils

// Handle error on program
// and print it to the log. Put it on defer
func ErrorHandler() bool {
	if err := recover(); err != nil {
		Logs.Sugar().Error(err)
		return false
	}
	return true
}
