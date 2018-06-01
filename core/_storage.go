// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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

	return stringToInt(string(b))
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

// mkfs formats a filesystem on block device with given path using given fstype
func mkfs(path, fstype, label string) error {
	family := fsFamily(fstype)
	mkfsCmd := mkfsCommand(fstype)

	cmd := []string{}

	// Add options for the sector size if it's not the default size
	_, logSec := sectorSize(path)
	if logSec > defaultBlockSize {
		optSector, err := familyFlag("sectorsize", family)
		if err != nil {
			fmt.Println(err)
		} else {
			cmd = append(cmd, optSector)
			cmd = append(cmd, string(logSec))
		}
	}

	// Always set the force option
	optForce, err := familyFlag("force", family)
	if err != nil {
		fmt.Println(err)
	} else {
		cmd = append(cmd, optForce)
	}

	if len(label) > 0 {
		// Always set the force option
		optLabel, err := familyFlag("label", family)
		if err != nil {
			fmt.Println(err)
		} else {
			cmd = append(cmd, optLabel)
			cmd = append(cmd, label)
		}
	}

	// Add the path to the command
	cmd = append(cmd, path)

	// Run the mkfs.<fstype> command
	out, err := exec.Command(mkfsCmd, cmd...).Output()
	if err != nil {
		fmt.Println(string(out))
		return err
	}
	fmt.Println(string(out))
	return nil
}

func mkfsCommand(fstype string) string {
	mkfsCommands := map[string]string{
		"btrfs":    "mkfs.btrfs",
		"ext2":     "mkfs.ext2",
		"ext3":     "mkfs.ext3",
		"ext4":     "mkfs.ext4",
		"fat":      "mkfs.vfat",
		"fat12":    "mkfs.vfat",
		"fat16":    "mkfs.vfat",
		"fat32":    "mkfs.vfat",
		"vfat":     "mkfs.vfat",
		"jfs":      "jfs_mkfs",
		"ntfs":     "mkntfs",
		"reiserfs": "mkfs.reiserfs",
		"swap":     "mkswap",
		"xfs":      "mkfs.xfs",
	}

	if val, ok := mkfsCommands[fstype]; ok {
		return val
	}
	return "mkfs.ext4"
}

func fsFamily(fstype string) string {
	family := map[string]string{
		"ext2":  "ext",
		"ext3":  "ext",
		"ext4":  "ext",
		"fat12": "fat",
		"fat16": "fat",
		"fat32": "fat",
		"vfat":  "fat",
	}

	if val, ok := family[fstype]; ok {
		return val
	}
	return "ext"
}

func sectorSize(path string) (int, int) {
	out, err := exec.Command(
		"lsblk", "--noheadings", "--bytes", "--output=PHY-SEC,LOG-SEC").Output()
	if err != nil {
		fmt.Printf("  Error fetching sector size for `%s`: %v", path, err)
		return defaultBlockSize, defaultBlockSize
	}

	// Output will be in the format: `    512     512`
	vals := strings.Split(strings.TrimSpace(string(out)), " ")

	if len(vals) < 2 {
		fmt.Printf("  Error fetching sector size for `%s`: %s", path, string(out))
		return defaultBlockSize, defaultBlockSize
	}

	// Physical sector is the first parameter
	phySec, err := stringToInt(vals[0])
	if err != nil {
		fmt.Printf("  Error fetching sector size for `%s`: %v", path, err)
		return defaultBlockSize, defaultBlockSize
	}

	for _, v := range vals[1:] {
		if len(v) > 0 {
			logSec, err := stringToInt(v)
			if err != nil {
				fmt.Printf("  Error fetching sector size for `%s`: %v", path, err)
				return defaultBlockSize, defaultBlockSize
			}
			return phySec, logSec
		}
	}

	fmt.Printf("  Error fetching sector size for `%s`: %s", path, string(out))
	return defaultBlockSize, defaultBlockSize
}

func familyFlag(flag, family string) (string, error) {
	switch flag {
	case "force":
		switch family {
		case "ext":
			return "-F", nil
		case "fat":
			return "-I", nil
		case "swap":
			return "--force", nil
		default:
			return "", fmt.Errorf("`force` for family `%s` is not implemented", family)
		}

	case "sectorsize":
		switch family {
		case "ext":
			return "-b", nil
		case "fat":
			return "-S", nil
		default:
			return "", fmt.Errorf("`sectorsize` for family `%s` is not implemented", family)
		}

	case "label":
		switch family {
		case "ext":
			return "-L", nil
		case "fat":
			return "-n", nil
		case "swap":
			return "--label", nil
		default:
			return "", fmt.Errorf("`label` for family `%s` is not implemented", family)
		}

	default:
		return "", fmt.Errorf("flag `%s` is not implemented", flag)
	}
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
