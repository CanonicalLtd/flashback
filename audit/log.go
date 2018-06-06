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
	// DefaultLogFile the log file is initially written to initramfs
	DefaultLogFile = "/run/initramfs/flashback.log"
)

func logFile() (*os.File, error) {
	return os.OpenFile(DefaultLogFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
}

// Printf records a response
func Printf(message string, a ...interface{}) {
	l, _ := logFile()
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
