package closer

import (
	"context"
	"fmt"
	"os"
)

type Logger interface {
	Errorf(ctx context.Context, format string, args ...interface{})
}

type stdErrLogger struct{}

func (s stdErrLogger) Errorf(_ context.Context, format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
}
