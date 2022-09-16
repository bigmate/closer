package closer

import (
	"fmt"
	"io"
	"os"
)

type Logger interface {
	Errorf(format string, args ...interface{})
}

func stdErrLogger() Logger {
	return logger{os.Stderr}
}

type logger struct {
	io.Writer
}

func (l logger) Errorf(format string, args ...interface{}) {
	fmt.Fprintf(l, format, args...)
	fmt.Fprintln(l)
}
