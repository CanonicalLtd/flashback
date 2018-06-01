// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package core

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	snadpdPath = "/system-data/var/lib/snapd"
)

func simpleHandler(blockMeta PartitionCommand, config Config) error {
	// Get the first dd-image from the sources section
	ddType, ddImage, err := firstDDImage(config.Sources)
	if err != nil {
		return err
	}
	fmt.Printf("Found image to restore `%s`\n", ddImage)

	// Get the first block device
	device, err := blockDevice(config)
	if err != nil {
		return err
	}

	// Check that this is a valid block device
	isBlk, err := isBlockDevice(device)
	if err != nil {
		return err
	}
	if !isBlk {
		return fmt.Errorf("  %s: not an existing block device", device)
	}
	fmt.Printf("Found device for the restore `%s`\n", device)

	// Write the image
	err = writeImageToDisk(ddType, ddImage, device)
	if err != nil {
		return err
	}

	// Get the path to the root device (containing /system-data/var/lib/snapd) e.g. /dev/sdd2
	rootDev, err := rootDevice(device)
	if err != nil {
		return err
	}
	fmt.Printf("Found root device to mount `%s`\n", rootDev)

	// Mount the device to the target
	return mount(rootDev, blockMeta.Target)
}

func firstDDImage(sources []interface{}) (string, string, error) {
	var ddType, ddImage string
	var err error

	// Get the first dd image
	for _, source := range sources {
		ddType, ddImage, err = sourceImage(source)
		if err != nil {
			return "", "", err
		}
		if len(ddImage) > 0 {
			// We have the first ddImage
			return ddType, ddImage, nil
		}
	}
	return "", "", errors.New("Cannot find a dd-image in the `sources`")
}

func sourceImage(sourceInterface interface{}) (string, string, error) {
	source := ""
	sourceKey := ""
	sourceMap, ok := sourceInterface.(map[interface{}]interface{})
	if !ok {
		// Probably a string, so not a dd-img
		return "", "", nil
	}

	for k, v := range sourceMap {
		source = v.(string)
		sourceKey = k.(string)

		// Check that we have a dd-image
		if strings.HasPrefix(sourceKey, "dd-") {
			return sourceKey, source, nil
		}
	}

	// Not the image that we're looking for
	return "", "", nil
}

func blockDevice(config Config) (string, error) {
	if len(config.BlockMeta.Devices) == 0 {
		return "", errors.New("No valid target devices found to install on")
	}
	if len(config.BlockMeta.Devices) > 0 {
		fmt.Println("Multiple devices given. Using the first available")
	}
	return config.BlockMeta.Devices[0], nil
}

func writeImageToDisk(ddType, ddImage, device string) error {
	extractor := imageExtractor(ddType)

	c := fmt.Sprintf("cat \"%s\" %s | dd bs=4M of=%s", ddImage, extractor, device)

	out, err := exec.Command("sh", "-c", c).Output()
	if len(out) > 0 {
		fmt.Println(string(out))
	}
	if err != nil {
		return err
	}
	fmt.Printf("Started writing image `%s` to device `%s`\n", ddImage, device)

	out, err = exec.Command("partprobe", device).Output()
	if len(out) > 0 {
		fmt.Println(string(out))
	}
	if err != nil {
		return err
	}

	return nil
}

func imageExtractor(ddType string) string {
	switch ddType {
	case "dd-tgz":
		return "|tar -xOzf -"
	case "dd-txz":
		return "|tar -xOJf -"
	case "dd-tbz":
		return "|tar -xOjf -"
	case "dd-tar":
		return "|smtar -xOf -"
	case "dd-bz2":
		return "|bzcat"
	case "dd-gz":
		return "|zcat"
	case "dd-xz":
		return "|xzcat"
	case "dd-raw":
		return ""
	default:
		return ""
	}
}

// rootDevice mounts the partitions and identifies the root device
func rootDevice(device string) (string, error) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)

	// Get the list of partitions
	sysBlockPath := sysBlockFromPath(device)
	partitions, err := sysfsPartitions(sysBlockPath)
	if err != nil {
		return "", err
	}

	for _, ptn := range partitions {
		// Mount the partition to the temp directory
		d := devicePathFromDevice(ptn)
		out, err := exec.Command("mount", d, dir).Output()
		if err != nil {
			fmt.Println(string(out))
			return "", err
		}

		// Check if the snapd path is there
		tmpPoint := filepath.Join(dir, snadpdPath)
		if _, err = os.Stat(tmpPoint); err == nil {
			// Unmount the partition
			out, err = exec.Command("umount", d).Output()
			if err != nil {
				fmt.Println(string(out))
				return "", err
			}
			return d, nil
		}

		// Unmount the partition
		out, err = exec.Command("umount", d).Output()
		if err != nil {
			fmt.Println(string(out))
			return "", err
		}
	}

	return "", fmt.Errorf("Could not find the root device")
}

func mount(device, target string) error {
	_ = os.MkdirAll(target, os.ModePerm)

	fmt.Println("Unmount the device")
	out, _ := exec.Command("umount", device).Output()
	if len(out) > 0 {
		fmt.Println(string(out))
	}

	fmt.Println("Mount the device as", target)
	out, err := exec.Command("mount", device, target).Output()
	if len(out) > 0 {
		fmt.Println(string(out))
	}

	return err
}
