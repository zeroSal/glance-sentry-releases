package logger

type LoggerInterface interface {
	GetIdentifier() string
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	Success(msg string)
	List([]string)
}
