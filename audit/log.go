// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package audit

import (
	"io"
	"log"
	"os"
)

const (
	defaultLogFile = "flashback.log"
)

// LogFile is the path to the log file
var LogFile = defaultLogFile

func logFile() (*os.File, error) {
	f, err := os.OpenFile(LogFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err == nil {
		return f, err
	}
	return os.OpenFile(defaultLogFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
}

// Printf records a response
func Printf(message string, a ...interface{}) {
	l, _ := logFile()
	//fmt.Printf(message, a)
	mw := io.MultiWriter(os.Stdout, l)
	log.SetOutput(mw)
	log.Printf(message, a...)
}

// Println records a response
func Println(v ...interface{}) {
	l, _ := logFile()
	mw := io.MultiWriter(os.Stdout, l)
	log.SetOutput(mw)
	log.Println(v...)
}
