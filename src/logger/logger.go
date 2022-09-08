package logger

type Logger interface {
	Debug(args ...interface{})
	Debugf(template string, fields ...interface{})
	Info(args ...interface{})
	Infof(template string, fields ...interface{})
	Warn(args ...interface{})
	Warnf(template string, fields ...interface{})
	Error(args ...interface{})
	Errorf(template string, fields ...interface{})
	Fatal(args ...interface{})
	Fatalf(template string, fields ...interface{})
}

var L Logger

func InitLogger(logger Logger) {
	L = logger
}
