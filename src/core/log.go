package core

import (
	"github.com/moeshin/go-errs"
	"io"
	"log"
	"net/http"
)

func NewLogger(out io.Writer) *log.Logger {
	logger := log.Default()
	return log.New(out, logger.Prefix(), logger.Flags())
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
