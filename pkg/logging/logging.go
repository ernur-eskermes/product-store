package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"runtime"

	"github.com/sirupsen/logrus"
)

type writerHook struct {
	Writer    []io.Writer
	LogLevels []logrus.Level
}

func (h writerHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}

	for _, w := range h.Writer {
		if _, err = w.Write([]byte(line)); err != nil {
			log.Fatal(err)
		}
	}

	return nil
}

func (h writerHook) Levels() []logrus.Level {
	return h.LogLevels
}

var e *logrus.Entry

type Logger struct {
	*logrus.Entry
}

func GetLogger() *Logger {
	return &Logger{e}
}

func (l *Logger) GetLoggerWithField(k string, v interface{}) Logger {
	return Logger{l.WithField(k, v)}
}

func init() {
	l := logrus.New()
	l.SetReportCaller(true)
	l.SetFormatter(&logrus.JSONFormatter{
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			filename := path.Base(frame.File)

			return fmt.Sprintf("%s()", frame.Function), fmt.Sprintf("%s:%d", filename, frame.Line)
		},
	})

	if err := os.MkdirAll("logs", 0o755); err != nil {
		log.Fatal(err)
	}

	allFile, err := os.OpenFile("logs/all.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o640)
	if err != nil {
		log.Fatal(err)
	}

	l.SetOutput(io.Discard)
	l.AddHook(&writerHook{
		Writer:    []io.Writer{allFile},
		LogLevels: []logrus.Level{logrus.ErrorLevel, logrus.WarnLevel, logrus.FatalLevel},
	})
	l.AddHook(&writerHook{
		Writer:    []io.Writer{os.Stdout},
		LogLevels: logrus.AllLevels,
	})
	l.SetLevel(logrus.TraceLevel)

	e = logrus.NewEntry(l)
}
