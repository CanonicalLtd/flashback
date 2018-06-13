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
func backupUserData() error {
	audit.Println("Backup user data to the tmpfs store")
	// Mount the writable path
	if err := core.Mount(core.PartitionTable.Writable, core.WritablePath); err != nil {
		return err
	}

	// Backup the directories to tmpfs
	for _, d := range config.Store.Backup.Data {
		// Check if the file/directory exists
		path := filepath.Join(core.WritablePath, core.SystemData, d)
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			audit.Println("Path not found:", d)
			continue
		}

		target := filepath.Join(core.TempFSMount, d)
		if info.IsDir() {
			audit.Println("Backup directory:", d)
			// Move up a directory so we don't create nested directories
			target, err = filepath.Abs(filepath.Join(target, ".."))
			if err != nil {
				return err
			}
			err = core.CopyDirectory(path, target)
		} else {
			audit.Println("Backup file:", d)
			err = core.CopyFile(path, target)
		}
		if err != nil {
			_ = core.Unmount(core.WritablePath)
			return err
		}
	}

	// Unmount the writable partition
	_ = core.Unmount(core.WritablePath)

	return nil
}

// restoreUserData restores the requested data from the tmpfs store
func restoreUserData() error {
	audit.Println("Restore user data from the tmpfs store")
	// Mount the writable path
	if err := core.Mount(core.PartitionTable.Writable, core.WritablePath); err != nil {
		return err
	}

	// Restore the directories
	for _, d := range config.Store.Backup.Data {
		// Check if the directory exists
		tempPath := filepath.Join(core.TempFSMount, d)
		info, err := os.Stat(tempPath)
		if os.IsNotExist(err) {
			audit.Println("Path not found:", d)
			continue
		}

		targetPath := filepath.Join(core.WritablePath, core.SystemData, d)
		if info.IsDir() {
			audit.Println("Restore directory:", d)
			// Move up a directory so we don't create nested directories
			targetPath, err = filepath.Abs(filepath.Join(targetPath, ".."))
			if err != nil {
				return err
			}
			if err := core.CopyDirectory(tempPath, targetPath); err != nil {
				return err
			}
		} else {
			audit.Println("Restore file:", d)
			if err := core.CopyFile(tempPath, targetPath); err != nil {
				return err
			}
		}
	}

	// Unmount the partitions
	_ = core.Unmount(core.WritablePath)
	_ = core.Unmount(core.TempFSMount)

	return nil
}
