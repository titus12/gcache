package gcache

import (
	"fmt"
	"log"
	"os"
)

type Logger interface {
	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Fatalf(format string, v ...interface{})
}

var l Logger = &dl{log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)}

type dl struct {
	*log.Logger
}

const (
	callDepth = 3
)

func prefix(lvl, msg string) string {
	return fmt.Sprintf("%s: %s", lvl, msg)
}

func (d *dl) Debugf(f string, v ...interface{}) {
	d.Output(callDepth, prefix("DEBUG", fmt.Sprintf(f, v...)))
}
func (d *dl) Infof(format string, v ...interface{}) {
	d.Output(callDepth, prefix("INFO ", fmt.Sprintf(format, v...)))
}
func (d *dl) Warnf(format string, v ...interface{}) {
	d.Output(callDepth, prefix("WARN ", fmt.Sprintf(format, v...)))
}
func (d *dl) Errorf(format string, v ...interface{}) {
	d.Output(callDepth, prefix("ERROR", fmt.Sprintf(format, v...)))
}
func (d *dl) Fatalf(format string, v ...interface{}) {
	d.Output(callDepth, prefix("FATAL", fmt.Sprintf(format, v...)))
	os.Exit(1)
}

func setLogger(logger Logger) {
	if logger == nil {
		return
	}
	l = logger
}

func L() Logger {
	return l
}
