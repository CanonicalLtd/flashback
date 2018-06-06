// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package reset

import (
	"github.com/CanonicalLtd/flashback/audit"
	"github.com/CanonicalLtd/flashback/config"
	"github.com/CanonicalLtd/flashback/core"
)

// Run starts the factory reset
func Run() error {
	audit.Println("Start a factory reset of the device")

	// Find the writable partition
	audit.Printf("Find the writable partition: %s", config.Store.WritablePartitionLabel)
	writable, err := core.FindFS(config.Store.WritablePartitionLabel)
	if err != nil {
		audit.Printf("Cannot find the writable partition: `%s` : %v\n", config.Store.WritablePartitionLabel, err)
		return nil
	}
	audit.Println("Found partition at", writable)

	// Find the restore partition
	audit.Printf("Find the restore partition: %s", config.Store.RestorePartitionLabel)
	restore, err := core.FindFS(config.Store.RestorePartitionLabel)
	if err != nil {
		audit.Printf("Cannot find the restore partition: `%s` : %v\n", config.Store.RestorePartitionLabel, err)
		return nil
	}
	audit.Println("Found partition at", restore)

	// Back up the requested data
	if err := core.BackupUserData(writable); err != nil {
		return err
	}

	// Format the new partition
	audit.Println("Format the writable partition:", writable)
	if err = core.FormatDisk(writable, "ext4", config.Store.WritablePartitionLabel); err != nil {
		return err
	}

	// Copy content from restore partition (renamed writable) to the new writable partition
	audit.Println("Copy the system data to the writable partition")
	if err = core.CopySystemData(restore, writable); err != nil {
		return err
	}

	// Restore system-boot to virgin state

	// Restore backed up data
	if err := core.RestoreUserData(writable); err != nil {
		return err
	}

	// Initiate reboot
	return nil
}
