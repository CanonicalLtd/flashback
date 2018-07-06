// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package core

import (
	"github.com/CanonicalLtd/flashback/audit"
)

// Partition identifies the path to the partitions
type Partition struct {
	SystemBoot string
	Restore    string
	Writable   string
}

// PartitionTable identifies the path to the partitions
var PartitionTable Partition

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
