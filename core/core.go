// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package core

import (
	"os/exec"
	"path/filepath"

	"github.com/CanonicalLtd/flashback/audit"
	"github.com/CanonicalLtd/flashback/config"
)

// CopySystemData copies system-data to the new writable partition
func CopySystemData(restore, newWritable string) error {
	// Mount the restore path
	if err := Mount(restore, RestorePath); err != nil {
		return err
	}

	// Mount the writable path
	if err := Mount(newWritable, TargetPath); err != nil {
		return err
	}

	// Copy the system data from the restore partition to the writable partition
	if err := CopyDirectory(SystemDataPath, TargetPath); err != nil {
		return err
	}

	// Backup log file to the writable partition
	targetLog := filepath.Join(TargetPath, SystemData, config.Store.LogFile)
	if err := CopyFile(audit.DefaultLogFile, targetLog); err != nil {
		return err
	}

	// Unmount the partitions
	_ = Unmount(RestorePath)
	_ = Unmount(TargetPath)

	//# close the device

	return nil
}

// UnlockCryptoFS calls the hook to unlock full-disk encryption
func UnlockCryptoFS() (string, error) {
	out, err := exec.Command(config.Store.EncryptUnlockAction).Output()
	return string(out), err
}

// LockCryptoFS calls the hook to unlock full-disk encryption
func LockCryptoFS() (string, error) {
	out, err := exec.Command(config.Store.EncryptLockAction).Output()
	return string(out), err
}
