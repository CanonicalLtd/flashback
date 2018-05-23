// -*- Mode: Go; indent-tabs-mode: t -*-
// Curtin Core
// Copyright 2018 Canonical Ltd.  All rights reserved.

package curtin

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

const (
	base             = "/sys/class/block"
	dev              = "/dev"
	defaultBlockSize = 512
)

// DiskPath defines a disk path in its different formats
type DiskPath struct {
	Device       string // sdd
	DevicePath   string // /dev/sdd
	SysBlockPath string // /sys/class/block/sdd
}

func devicePathFromDevice(devName string) string {
	return filepath.Join(dev, devName)
}

func deviceNameFromPath(path string) string {
	return filepath.Base(path)
}

func sysBlockFromDevice(devName string) string {
	return filepath.Join(base, devName)
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
			partitions = append(partitions, d.Name())
		}
	}
	return partitions, nil
}

func logicalBlockSize(sysfsPath string) int {
	lbsPath := filepath.Join(sysfsPath, "queue", "logical_block_size")
	i, err := readSizeFromContent(lbsPath)
	if err != nil {
		return defaultBlockSize
	}
	return i
}

func partitionSize(sysfsPath string) int {
	lbsPath := filepath.Join(sysfsPath, "size")
	i, err := readSizeFromContent(lbsPath)
	if err != nil {
		return 0
	}
	return i
}

func partitionStart(sysfsPath string) int {
	lbsPath := filepath.Join(sysfsPath, "start")
	i, err := readSizeFromContent(lbsPath)
	if err != nil {
		return 0
	}
	return i
}

func readSizeFromContent(path string) (int, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return 0, err
	}

	// Remove any control characters e.g. LF
	reg, err := regexp.Compile("[^0-9]+")
	if err != nil {
		return 0, err
	}
	cleaned := reg.ReplaceAllString(string(b), "")

	return strconv.Atoi(cleaned)
}

func sgDiskFlag(flag string) string {
	sgdiskFlags := map[string]string{
		"boot":      "ef00",
		"lvm":       "8e00",
		"raid":      "fd00",
		"bios_grub": "ef02",
		"prep":      "4100",
		"swap":      "8200",
		"home":      "8302",
		"linux":     "8300"}

	if val, ok := sgdiskFlags[flag]; ok {
		return val
	}
	return sgdiskFlags["linux"]
}
