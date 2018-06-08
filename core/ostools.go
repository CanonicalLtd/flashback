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

	"github.com/CanonicalLtd/flashback/audit"
)

// Constants for saving the system image
const (
	BackupImageWritable   = "/restore/writable.img.gz"
	BackupImageSystemBoot = "/restore/system-boot.img.gz"
	RestorePath           = "/restore"
	TargetPath            = "/target"
	SystemDataPath        = "/restore/system-data"
	SystemData            = "system-data"
	TempBackupPath        = "/tmp/flashbackup"
	MMCPrefix             = "mmcblk"
)

// FindFS locates a filesystem by label
func FindFS(label string) (string, error) {
	out, err := exec.Command("findfs", fmt.Sprintf("LABEL=%s", label)).Output()

	// Remove non-printable characters from the response
	cleaned := cleanOutput(string(out))
	return cleaned, err
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
