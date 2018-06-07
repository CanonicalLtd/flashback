// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package core

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/CanonicalLtd/flashback/audit"
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

// RootDeviceNameFromPath converts a path name from:
//   /dev/sdd1 to sdd
//   /dev/mmcblk1p2 to mmcblk1
func RootDeviceNameFromPath(path string) string {
	var re *regexp.Regexp
	p := filepath.Base(path)

	if strings.HasPrefix(p, MMCPrefix) {
		// mmcblk1p2 format
		re = regexp.MustCompile("p[0-9]+$")
	} else {
		// sdd format
		re = regexp.MustCompile("[^/a-zA-Z]")
	}
	return re.ReplaceAllString(p, "")
}

// DevicePathFromDevice converts sdd1 to /dev/sdd1
func DevicePathFromDevice(devName string) string {
	return filepath.Join(dev, devName)
}

// DevicePathFromNumber gets the device path using a number
// e.g. /dev/sdd1 + 3 = /dev/sdd3
// e.g. /dev/mmcblk1p2 + 3 = /dev/mmcblk1p3
func DevicePathFromNumber(device string, number int) string {
	var name string
	d := RootDeviceNameFromPath(device)
	if strings.HasPrefix(d, MMCPrefix) {
		name = fmt.Sprintf("%sp%d", d, number)
	} else {
		name = fmt.Sprintf("%s%d", d, number)
	}
	return filepath.Join(dev, name)
}

// DeviceNameFromPath converts the path /dev/sdd1 to sdd1
func DeviceNameFromPath(path string) string {
	return filepath.Base(path)
}

// DeviceNumberFromPath converts the path /dev/sdd1 to 1
func DeviceNumberFromPath(path string) (int, error) {
	d := filepath.Base(path)

	if strings.HasPrefix(d, MMCPrefix) {
		re := regexp.MustCompile("^mmcblk[0-9]+")
		d = re.ReplaceAllString(d, "")
	}

	return stringToInt(d)
}

// SysBlockFromDevice returns the /sys/class/block/<dev> from <dev>
func SysBlockFromDevice(devName string) string {
	return filepath.Join(base, devName)
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
