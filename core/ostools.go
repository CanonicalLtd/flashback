// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package core

import (
	"archive/tar"
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
	PartitionSystemBoot   = "system-boot"
	PartitionRestore      = "restore"
	PartitionWritable     = "writable"
	BackupImageWritable   = "/restore/writable.tar.gz"
	BackupImageSystemBoot = "/restore/system-boot.img.gz"
	RestorePath           = "/restore"
	WritablePath          = "/writable"
	SystemDataPath        = "/restore/system-data"
	SystemData            = "system-data"
	TempBackupPath        = "/tmp/flashbackup"
	TempFSMount           = "/mnt/tmprestore"
	MMCPrefix             = "mmcblk"
)

// FindFS locates a filesystem by label
func FindFS(label string) (string, error) {
	out, err := exec.Command("findfs", fmt.Sprintf("LABEL=%s", label)).CombinedOutput()

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
	logSec := sectorSize(path)
	if logSec > defaultBlockSize {
		optSector, err := familyFlag("sectorsize", family)
		if err != nil {
			audit.Println(err)
		} else {
			cmd = append(cmd, optSector)
			cmd = append(cmd, string(logSec))
		}
	}

	// Always set the force option
	optForce, err := familyFlag("force", family)
	if err != nil {
		audit.Println(err)
	} else {
		cmd = append(cmd, optForce)
	}

	if len(label) > 0 {
		// Set the label on the disk
		optLabel, err := familyFlag("label", family)
		if err != nil {
			audit.Println(err)
		} else {
			cmd = append(cmd, optLabel)
			cmd = append(cmd, label)
		}
	}

	// Add the path to the command
	cmd = append(cmd, path)

	// Run the mkfs.<fstype> command
	out, err := exec.Command(mkfsCmd, cmd...).CombinedOutput()
	if err != nil {
		audit.Println(string(out))
		return err
	}
	audit.Println(string(out))
	return nil
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

	out, err := exec.Command("mount", device, target).CombinedOutput()
	if len(out) > 0 {
		audit.Println("Mount the device as", target)
		audit.Println(string(out))
	}

	return err
}

// Unmount unmounts the device
func Unmount(device string) error {
	out, err := exec.Command("umount", device).Output()
	if len(out) > 0 {
		audit.Println("Unmount the device", device)
		audit.Println(string(out))
	}

	return err
}

// CopyDirectory from one location to another
func CopyDirectory(sourceDir, destDir string) (err error) {
	// Make sure the target path exists
	_ = os.MkdirAll(destDir, os.ModePerm)

	out, err := exec.Command("cp", "-arv", sourceDir, destDir).Output()
	audit.Println(string(out))
	return err
}

// CopyFile from one location to another
func CopyFile(source, target string) error {
	// Make sure the target path exists
	_ = os.MkdirAll(filepath.Dir(target), os.ModePerm)

	out, err := exec.Command("cp", "-av", source, target).Output()
	audit.Println(string(out))
	return err
}

// CreateTmpfsDisk creates a RAM disk of a fixed size
func CreateTmpfsDisk(mount string, size int) error {
	audit.Println("Create a RAM disk of size", size, "bytes")
	_ = os.MkdirAll(mount, os.ModePerm)
	out, err := exec.Command("mount", "-t", "tmpfs", "-o", fmt.Sprintf("size=%dM", size), "tmpfs", mount).CombinedOutput()
	if len(out) > 0 {
		audit.Println(string(out))
	}
	return err
}

// FSType retrieves the file-system type of a partition
func FSType(device string) (string, error) {
	out, err := exec.Command("blkid", "-o", "value", "-s", "TYPE", device).Output()
	if err != nil {
		return "", err
	}
	return cleanOutput(string(out)), nil
}

// Tar creates a tarball from a file or directory structure
func Tar(source string, tarball *tar.Writer) error {
	// Check that the source exists
	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	return filepath.Walk(source,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return err
			}

			if baseDir != "" {
				header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
			}

			if err := tarball.WriteHeader(header); err != nil {
				return err
			}

			// Skip copying contents for directories and non-regular files e.g. symlinks
			if info.IsDir() {
				return nil
			}
			if !info.Mode().IsRegular() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(tarball, file)
			return err
		})
}

func sectorSize(path string) int {
	out, err := exec.Command(
		"blkid", "-i", "-o", "value", "-s", "LOGICAL_SECTOR_SIZE", path).Output()
	if err != nil {
		fmt.Printf("Error fetching sector size for `%s`: %v", path, err)
		return defaultBlockSize
	}

	logSec, err := stringToInt(string(out))
	if err == nil {
		return logSec
	}

	fmt.Printf("  Error fetching sector size for `%s`: %s", path, string(out))
	return defaultBlockSize
}
