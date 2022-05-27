package core

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"github.com/moeshin/go-errs"
	"io"
	"log"
	"net/http"
	"time"
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
	Log *LogE
}

func (l *logWriter) Write(p []byte) (int, error) {
	n, err := l.Log.Writer().Write(p)
	_, _ = l.Buffer.Write(p)
	return n, err
}

type LogN struct {
	*LogE
	Writer   *logWriter
	Canceled bool
}

func (l *LogE) NewLogN() *LogN {
	var tag string
	var buf [4]byte
	_, err := io.ReadFull(rand.Reader, buf[:])
	if l.ErrPrint(err) {
		tag = fmt.Sprintf("[%d] ", time.Now().UnixMilli())
	} else {
		tag = fmt.Sprintf("[%X] ", buf[:])
	}
	writer := &logWriter{
		Buffer: &bytes.Buffer{},
		Log:    l,
	}
	return &LogN{
		LogE: &LogE{
			Logger: NewLogger(writer, tag),
		},
		Writer: writer,
	}
}

func (l *LogN) Notify() error {
	//data, err := ioutil.ReadAll(l.Writer.Buffer)
	//if err != nil {
	//	return err
	//}
	//log.Println("==== Notify ====")
	//log.Println(string(data))
	return nil
}

func (l *LogN) Close() error {
	return l.Notify()
}
