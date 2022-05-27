package core

import (
	"bytes"
	"github.com/moeshin/go-errs"
	"io"
	"log"
	"net/http"
)

func NewLogger(out io.Writer, prefix string) *log.Logger {
	logger := log.Default()
	return log.New(out, prefix+logger.Prefix(), logger.Flags())
}

type LogE struct {
	*log.Logger
}

func (l *LogE) ErrPrint(err error) bool {
	return errs.PrintWithDepthToLogger(err, 1, l.Logger)
}

func (l *LogE) ErrClose(closer io.Closer) {
	if closer == nil {
		return
	}
	l.ErrPrint(closer.Close())
}

func (l *LogE) CloseResponse(resp *http.Response) {
	l.ErrClose(resp.Body)
}

type logWriter struct {
	*bytes.Buffer
	Skip bool
	Log  *LogE
}

func (l *logWriter) Write(p []byte) (int, error) {
	n, err := l.Log.Writer().Write(p)
	if !l.Skip {
		_, _ = l.Buffer.Write(p)
	}
	return n, err
}
