// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package reset

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"

	"github.com/CanonicalLtd/flashback/audit"
	"github.com/CanonicalLtd/flashback/config"
	"github.com/CanonicalLtd/flashback/core"
)

// backupUserData backs up the requested data to the RAM disk
func backupUserData() error {
	audit.Println("Backup user data to the RAM disk")
	// Mount the writable path
	if err := core.Mount(core.PartitionTable.Writable, core.TargetPath); err != nil {
		return err
	}

	// Create the tar file
	t := filepath.Join(core.TempFSMount, "userdata.tar.gz")
	tarfile, err := os.Create(t)
	if err != nil {
		return err
	}
	defer tarfile.Close()

	// Open the gzip writer
	gw := gzip.NewWriter(tarfile)
	defer gw.Close()

	// Open the tar writer
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Backup the directories to the RAM disk (copy of the restore partition)
	for _, d := range config.Store.Backup.Directories {
		// Check if the directory exists
		dir := filepath.Join(core.TargetPath, core.SystemData, d)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			audit.Println("Directory not found:", d)
			continue
		}

		audit.Println("Backup directory:", d)
		if err := core.Tar(dir, tw); err != nil {
			return err
		}
	}

	// Backup the files
	for _, f := range config.Store.Backup.Files {
		// Check if the file exists
		file := filepath.Join(core.TargetPath, core.SystemData, f)
		if _, err := os.Stat(file); os.IsNotExist(err) {
			audit.Println("File not found:", f)
			continue
		}

		audit.Println("Backup file:", f)
		if err := core.Tar(file, tw); err != nil {
			return err
		}
	}

	// Unmount the writable partition
	_ = core.Unmount(core.TargetPath)

	return nil
}

// restoreUserData restores the requested data from the RAM disk
func restoreUserData(writable string) error {
	audit.Println("Restore user data from the RAM disk")
	// Mount the writable path
	if err := core.Mount(writable, core.TargetPath); err != nil {
		return err
	}

	// Restore the directories
	for _, d := range config.Store.Backup.Directories {
		// Check if the directory exists
		tempPath := filepath.Join(core.TempBackupPath, d)
		if _, err := os.Stat(tempPath); os.IsNotExist(err) {
			audit.Println("Directory not found:", d)
			continue
		}
		audit.Println("Restore directory:", d)
		targetPath := filepath.Join(core.TargetPath, core.SystemData, d)
		if err := core.CopyDirectory(tempPath, targetPath); err != nil {
			return err
		}
	}

	// Restore the files
	for _, f := range config.Store.Backup.Files {
		// Check if the file exists
		tempPath := filepath.Join(core.TempBackupPath, f)
		if _, err := os.Stat(tempPath); os.IsNotExist(err) {
			audit.Println("File not found:", f)
			continue
		}

		audit.Println("Restore file:", f)
		file := filepath.Join(core.TargetPath, core.SystemData, f)
		if err := core.CopyFile(tempPath, file); err != nil {
			return err
		}
	}

	// Unmount the writable partition
	_ = core.Unmount(core.TargetPath)

	return nil
}

// createRestoreRAMDisk creates a copy of the restore partition as a RAM disk
func createRestoreRAMDisk() error {
	// Get the size of the restore partition
	size, err := core.PartitionSize(core.PartitionTable.Restore)
	if err != nil {
		return err
	}

	// Mount the restore path
	if err := core.Mount(core.PartitionTable.Restore, core.RestorePath); err != nil {
		return err
	}

	// Create the RAM disk the same size
	if err = core.CreateRAMDisk(core.TempFSMount, size); err != nil {
		return err
	}

	// Copy the drive to the RAM disk
	audit.Println("Copy", core.RestorePath+"/", "to", core.TempFSMount)
	return core.CopyDirectory(core.RestorePath+"/", core.TempFSMount)
}

func copyRAMDiskToRestore() error {
	// The partitions are already mounted

	audit.Println("Copy the data from the RAM partition to the restore partition")
	return core.CopyDirectory(core.TempFSMount+"/", core.RestorePath)
}
