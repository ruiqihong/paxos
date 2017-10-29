package log

type F map[string]interface{}

type Logger interface {
	Debug(string)
	Info(string)
	Warn(string)
	Err(string)
}

func With(fields F) Logger {
	return nil
}
