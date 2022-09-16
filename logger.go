package closer

import (
	"fmt"
	"io"
	"os"
	"sync"
)

type Logger interface {
	Errorf(format string, args ...interface{})
}

func stdErrLogger() Logger {
	return &logger{dest: os.Stderr}
}

type logger struct {
	mu   sync.Mutex
	dest io.Writer
}

func (l *logger) Errorf(format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	fmt.Fprintf(l.dest, format, args...)
	fmt.Fprintln(l.dest)
}
