// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package core

import (
	"regexp"
	"strconv"

	"github.com/CanonicalLtd/flashback/audit"
)

func cleanOutput(s string) string {
	// Remove any control characters e.g. LF
	reg, err := regexp.Compile("[^a-zA-Z0-9/]+")
	if err != nil {
		audit.Println("Error cleaning string:", err)
		return ""
	}
	return reg.ReplaceAllString(s, "")
}

func stringToInt(s string) (int, error) {
	// Remove any control characters e.g. LF
	reg, err := regexp.Compile("[^0-9]+")
	if err != nil {
		return 0, err
	}
	cleaned := reg.ReplaceAllString(s, "")

	return strconv.Atoi(cleaned)
}
