package game

type Logger struct {
	rows *[]string
}

func NewLogger() Logger {
	return Logger{
		rows: new([]string),
	}
}

func (l Logger) Write(p []byte) (n int, err error) {
	*l.rows = append(*l.rows, string(p))
	return len(p), nil
}

func (l Logger) GetLogs() []string {
	return *l.rows
}
