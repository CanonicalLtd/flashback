// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package bootprint

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"

	"github.com/CanonicalLtd/flashback/audit"
	"github.com/CanonicalLtd/flashback/core"
)

// backupPartition makes a raw backup of a partition
func backupPartition(devicePath, imagePath string) error {
	// Mount the restore path
	err := core.Mount(core.PartitionTable.Restore, core.RestorePath)
	if err != nil {
		return err
	}

	// Unmount the boot path
	_ = core.Unmount(devicePath)

	// Back up the partition to img file so we keep the exact filesystem
	// without having to parse gadget.yaml or worrying about ABI compatibility
	// to ubuntu-image's dosfstools
	err = core.ReadAndGzipToFile(devicePath, imagePath)

	// Unmount the restore partition
	_ = core.Unmount(core.RestorePath)

	return err
}

// backupWritable makes a backup of the files on the writable partition
// We don't use an image as we'd need to regenerate the encryption key
func backupWritable() error {
	audit.Println("Backup writable partition to the restore partition")
	// Mount the writable path
	if err := core.Mount(core.PartitionTable.Writable, core.WritablePath); err != nil {
		return err
	}

	// Mount the restore path
	err := core.Mount(core.PartitionTable.Restore, core.RestorePath)
	if err != nil {
		return err
	}

	// Create the tar file
	tarfile, err := os.Create(core.BackupImageWritable)
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

	// Check the path to system-data
	source := filepath.Join(core.WritablePath, core.SystemData)
	if _, err := os.Stat(source); os.IsNotExist(err) {
		audit.Println("Directory not found:", core.SystemData)
		return err
	}

	// Add the directory to the archive
	audit.Println("Backup directory:", core.SystemData)
	if err := core.Tar(source, tw); err != nil {
		return err
	}

	// Unmount the writable partition
	_ = core.Unmount(core.WritablePath)
	_ = core.Unmount(core.RestorePath)

	return nil
}

// backupSystemBoot makes a raw backup of system-boot partition
func backupSystemBoot() error {
	return backupPartition(core.PartitionTable.SystemBoot, core.BackupImageSystemBoot)
}
