package logger

type NopLogger struct{}

func NewNopLogger() Logger {
	return &NopLogger{}
}

func (n *NopLogger) Debug(msg string, args ...Field) {}
func (n *NopLogger) Info(msg string, args ...Field)  {}
func (n *NopLogger) Warn(msg string, args ...Field)  {}
func (n *NopLogger) Error(msg string, args ...Field) {}
