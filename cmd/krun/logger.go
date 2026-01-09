package main

import (
	"fmt"
	"os"
)

type Logger interface {
	Debugf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Infof(format string, args ...interface{})
}

type ConsoleLogger struct {
	verbose bool
}

func (l *ConsoleLogger) Debugf(format string, args ...interface{}) {
	if l.verbose {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}

func (l *ConsoleLogger) Errorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func (l *ConsoleLogger) Infof(format string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, format+"\n", args...)
}
