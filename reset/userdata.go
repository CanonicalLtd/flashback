// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package reset

import (
	"os"
	"path/filepath"

	"github.com/CanonicalLtd/flashback/audit"
	"github.com/CanonicalLtd/flashback/config"
	"github.com/CanonicalLtd/flashback/core"
)

// backupUserData backs up the requested data to the RAM disk
func backupUserData(writable string) error {
	audit.Println("Backup user data to the RAM disk")
	// Mount the writable path
	if err := core.Mount(writable, core.TargetPath); err != nil {
		return err
	}

	// Make the backup directory on the RAM disk
	_ = os.MkdirAll(core.TempBackupPath, os.ModePerm)

	// Backup the directories
	for _, d := range config.Store.Backup.Directories {
		// Check if the directory exists
		dir := filepath.Join(core.TargetPath, core.SystemData, d)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			audit.Println("Directory not found:", d)
			continue
		}

		audit.Println("Backup directory:", d)
		tempPath := filepath.Join(core.TempBackupPath, d)
		if err := core.CopyDirectory(dir, tempPath); err != nil {
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
		tempPath := filepath.Join(core.TempBackupPath, f)
		if err := core.CopyFile(file, tempPath); err != nil {
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
