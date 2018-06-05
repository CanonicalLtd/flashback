// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package core

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

func sysBlockFromDevice(devName string) string {
	return filepath.Join(base, devName)
}

func sysfsPartitions(devPath string) ([]string, error) {
	// List the directories in the sys device path
	dirs, err := ioutil.ReadDir(devPath)
	if err != nil {
		return []string{}, err
	}

	// Check if the directory is a partition directory
	partitions := []string{}
	for _, d := range dirs {
		p := filepath.Join(base, d.Name(), "partition")

		if _, err := os.Stat(p); err == nil {
			partitions = append(partitions, d.Name())
		}
	}
	return partitions, nil
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
