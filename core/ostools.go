// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package core

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/CanonicalLtd/flashback/audit"
)

// Constants for saving the system image
const (
	BackupImagePath = "/restore/system-boot.img.gz"
	RestorePath     = "/restore"
	TargetPath      = "/target"
	SystemDataPath  = "/restore/system-data"
	SystemData      = "system-data"
	TempBackupPath  = "/tmp/flashbackup"
	MMCPrefix       = "mmcblk"
)

// FindFS locates a filesystem by label
func FindFS(label string) (string, error) {
	out, err := exec.Command("findfs", fmt.Sprintf("LABEL=%s", label)).Output()

	// Remove non-printable characters from the response
	cleaned := cleanOutput(string(out))
	return cleaned, err
}

// FreePartitionSpace finds the free space on a device
func FreePartitionSpace(device string) (int, error) {
	out, err := exec.Command("parted", "-ms", device, "unit B print").Output()
	if err != nil {
		return 0, err
	}

	// Parse the output to get the unit number
	s := string(out)
	last := strings.Split(s, ":")

	return stringToInt(last[7])
}

// RelabelDisk changes the label of a disk
func RelabelDisk(device, label string) (string, error) {
	out, err := exec.Command("tune2fs", "-L", label, device).Output()

	// Remove non-printable characters from the response
	cleaned := cleanOutput(string(out))
	return cleaned, err
}

// // CreatePartitionGPT creates a new partition that occupies the rest of the disk
// func CreatePartitionGPT(tableType, devicePath, label string) error {
// 	var cmd *exec.Cmd
// 	switch tableType {
// 	case "gpt":
// 		cmd = exec.Command("sgdisk",
// 			fmt.Sprintf("--largest-new=%s", devicePath),
// 			fmt.Sprintf("--change-name=%s", label),
// 		)
// 	default:
// 		return fmt.Errorf("unknown partition table type: %s", tableType)
// 	}

// 	out, err := cmd.Output()
// 	if len(out) > 0 {
// 		audit.Println(string(out))
// 	}
// 	return err
// }

// CreateNextPartition creates a new partition that occupies the rest of the disk.
// Creates /dev/sdd3 that occupies the rest of the disk when supplied with /dev/sdd2
func CreateNextPartition(currentDevice string) (string, error) {
	// Get the previous partition details
	var (
		currSizeSectors  int
		currStartSectors int
		offsetSectors    int
	)

	// Get the current partition number from the device e.g. 2 from /dev/sdd2
	number, err := DeviceNumberFromPath(currentDevice)
	if err != nil {
		return "", fmt.Errorf("The device name does not include the partition number: %v", err)
	}

	// Get the details of this partition
	currPtnName := DeviceNameFromPath(currentDevice)
	currPtn := SysBlockFromDevice(currPtnName) // /sys/class/block/sdd1
	lbs := logicalBlockSize(currPtn)
	currSize := partitionSize(currPtn)
	currStart := partitionStart(currPtn)
	currSizeSectors = int(currSize * 512 / lbs)
	currStartSectors = int(currStart * 512 / lbs)

	alignmentOffset := int((1 << 20)) / lbs
	if number == 1 {
		offsetSectors = alignmentOffset
	} else {
		// We're only handling GPT and physical partitions (not extended/logical ones)
		offsetSectors = currStartSectors + currSizeSectors
	}

	// Format the name of the device
	rootDeviceName := RootDeviceNameFromPath(currentDevice) // e.g. sdd
	rootDevicePath := DevicePathFromDevice(rootDeviceName)  // e.g. /dev/sdd

	lastDevicePath := fmt.Sprintf("%s%d", rootDevicePath, number+1)

	// Run `parted` to create the new partition
	audit.Printf("Create partition `%s` starting at sector %d\n", lastDevicePath, offsetSectors)
	out, err := exec.Command(
		"parted",
		"-ms", rootDevicePath,
		"mkpart", "primary", fmt.Sprintf("%ds", offsetSectors), "100%").Output()
	if err != nil {
		return lastDevicePath, fmt.Errorf("Error creating partition `%s`: %v", lastDevicePath, err)
	}
	audit.Printf("Partition `%s` created: %s\n", lastDevicePath, out)
	return lastDevicePath, nil
}

// FormatDisk formats and labels a disk
func FormatDisk(path, fstype, label string) error {
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

// RefreshPartitionTable refreshes the partition table by re-reading it
func RefreshPartitionTable(device string) error {
	rootDevName := RootDeviceNameFromPath(device)
	out, err := exec.Command("blockdev", "--rereadpt", DevicePathFromDevice(rootDevName)).Output()
	if len(out) > 0 {
		audit.Println(string(out))
	}
	return err
}

// ReadAndGzipToFile reads a file/device, zips it and writes it to a file
func ReadAndGzipToFile(inFile, outFile string) error {
	// Open the input file
	fIn, err := os.Open(inFile)
	if err != nil {
		audit.Println("Error backing up system-boot (open input):", err)
		return err
	}
	defer fIn.Close()

	// Create the output file
	fOut, err := os.Create(outFile)
	if err != nil {
		audit.Println("Error backing up system-boot (open output):", err)
		return err
	}
	defer fOut.Close()

	// Read from the input and gzip it
	buffer := bufio.NewReader(fIn)
	gw := gzip.NewWriter(fOut)

	// Take the buffered input and write it to the output file via gzip
	n, err := io.Copy(gw, buffer)
	gw.Close()
	audit.Printf("%d bytes read, compressed and written to file", n)
	return err
}

// UnzipToDevice reads gzip file and decompresses it to a device
func UnzipToDevice(inFile, device string) error {
	// Open the input file
	fIn, err := os.Open(inFile)
	if err != nil {
		audit.Println("Error restoring system-boot (open input):", err)
		return err
	}
	defer fIn.Close()

	// Create the output file
	fOut, err := os.Create(device)
	if err != nil {
		audit.Println("Error restoring system-boot (open output):", err)
		return err
	}
	defer fOut.Close()

	// Read from the gzip reader and output to the writer
	gr, err := gzip.NewReader(fIn)
	if err != nil {
		audit.Println("Error restoring system-boot (gzip reader):", err)
		return err
	}
	buffer := bufio.NewWriter(fOut)

	// Take the gzipped input and write it to the output device
	n, err := io.Copy(buffer, gr)
	gr.Close()
	audit.Printf("%d bytes read, uncompressed and written to device", n)
	return err
}

// Mount mounts the device at a path
func Mount(device, target string) error {
	_ = os.MkdirAll(target, os.ModePerm)

	// Unmount the device, just in case
	_ = Unmount(device)

	audit.Println("Mount the device as", target)
	out, err := exec.Command("mount", device, target).Output()
	if len(out) > 0 {
		audit.Println(string(out))
	}

	return err
}

// Unmount unmounts the device
func Unmount(device string) error {
	audit.Println("Unmount the device", device)
	out, err := exec.Command("umount", device).Output()
	if len(out) > 0 {
		audit.Println(string(out))
	}

	return err
}

// CopyDirectory from one location to another
func CopyDirectory(source, target string) error {
	out, err := exec.Command("rsync", "-a", source, target).Output()
	if len(out) > 0 {
		audit.Println(string(out))
	}
	return err
}

// CopyFile from one location to another
func CopyFile(source, target string) error {
	_ = os.MkdirAll(filepath.Dir(target), os.ModePerm)
	out, err := exec.Command("cp", "-a", source, target).Output()
	if len(out) > 0 {
		audit.Println(string(out))
	}

	return err
}
