// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package core

import (
	"os/exec"

	"github.com/CanonicalLtd/flashback/audit"
	"github.com/CanonicalLtd/flashback/config"
)

// Partition identifies the path to the partitions
type Partition struct {
	SystemBoot string
	Restore    string
	Writable   string
}

// PartitionTable identifies the path to the partitions
var PartitionTable Partition

// // CopySystemData copies system-data to the new writable partition
// func CopySystemData(restore, newWritable string) error {
// 	// Mount the restore path
// 	if err := Mount(restore, RestorePath); err != nil {
// 		return err
// 	}

// 	// Mount the writable path
// 	if err := Mount(newWritable, TargetPath); err != nil {
// 		return err
// 	}

// 	// Copy the system data from the restore partition to the writable partition
// 	if err := CopyDirectory(SystemDataPath, TargetPath); err != nil {
// 		return err
// 	}

// 	// Backup log file to the writable partition
// 	targetLog := filepath.Join(TargetPath, SystemData, config.Store.LogFile)
// 	if err := CopyFile(audit.DefaultLogFile, targetLog); err != nil {
// 		return err
// 	}

// 	// Unmount the partitions
// 	_ = Unmount(RestorePath)
// 	_ = Unmount(TargetPath)

// 	//# close the device

// 	return nil
// }

// UnlockCryptoFS calls the hook to unlock full-disk encryption
func UnlockCryptoFS() (string, error) {
	out, err := exec.Command(config.Store.EncryptUnlockAction).CombinedOutput()
	return string(out), err
}

// LockCryptoFS calls the hook to unlock full-disk encryption
func LockCryptoFS() (string, error) {
	out, err := exec.Command(config.Store.EncryptLockAction).CombinedOutput()
	return string(out), err
}

// FindPartitions locates the three main partitions
func FindPartitions() error {
	// Find "writable" partition and matching disk device
	audit.Printf("Find the writable partition: %s", PartitionWritable)
	writable, err := FindFS(PartitionWritable)
	if err != nil {
		audit.Printf("Cannot find the writable partition: `%s` : %v\n", PartitionWritable, err)
		return err
	}
	audit.Println("Found writable partition at", writable)

	// Find "restore" partition and matching disk device
	audit.Printf("Find the restore partition: %s", PartitionRestore)
	restore, err := FindFS(PartitionRestore)
	if err != nil {
		audit.Printf("Cannot find the restore partition: `%s` : %v\n", PartitionRestore, err)
		return err
	}
	audit.Println("Found restore partition at", restore)

	// Find "system-boot" partition and matching disk device
	audit.Printf("Find the system-boot partition: %s", PartitionSystemBoot)
	systemboot, err := FindFS(PartitionSystemBoot)
	if err != nil {
		audit.Printf("Cannot find the system-boot partition: `%s` : %v\n", PartitionSystemBoot, err)
		return err
	}
	audit.Println("Found system-boot partition at", systemboot)

	// Save the partition device paths
	PartitionTable.Restore = restore
	PartitionTable.SystemBoot = systemboot
	PartitionTable.Writable = writable
	return nil
}
