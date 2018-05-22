// -*- Mode: Go; indent-tabs-mode: t -*-
// Curtin Core
// Copyright 2018 Canonical Ltd.  All rights reserved.

package curtin

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	base = "/sys/class/block"
)

// DiskPath defines a disk path in its different formats
type DiskPath struct {
	Device       string // sdd
	DevicePath   string // /dev/sdd
	SysBlockPath string // /sys/class/block/sdd
}

func deviceNameFromPath(path string) string {
	return filepath.Base(path)
}

func sysBlockFromPath(path string) string {
	return filepath.Join(base, deviceNameFromPath(path))
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
			fmt.Println(d.Name())
			partitions = append(partitions, d.Name())
		}
	}
	return partitions, nil
}
