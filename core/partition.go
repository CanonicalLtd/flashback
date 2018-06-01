// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package core

import (
	"errors"
	"fmt"
	"os"
)

const (
	modeCustom          = "custom"
	modeSimple          = "simple"
	modeSimpleBoot      = "simple-boot"
	envTargetMountPoint = "TARGET_MOUNT_POINT"
)

// PartitionCommand defines the execution options for partitioning
type PartitionCommand struct {
	Devices    []string `short:"D" long:"devices" description:"read configuration from cfg"`
	FSType     string   `long:"fstype" description:"root partition filesystem type" choice:"ext4" choice:"ext3" default:"ext4"`
	BootFSType string   `long:"boot-fstype" description:"boot partition filesystem type" choice:"ext4" choice:"ext3"`
	Target     string   `short:"t" long:"target" description:"chroot to target. default is env[TARGET_MOUNT_POINT]"`
	UMount     bool     `long:"umount" description:"unmount any mounted filesystems before exit"`
	Mode       string   `positional_arg_name:"mode" description:"meta-mode to use"`
}

// Partition run a partitioning command
func Partition(blockMeta PartitionCommand, config Config) error {
	var err error

	switch blockMeta.Mode {
	case modeCustom:
		err = metaCustom(blockMeta, config)
		if err != nil {
			fmt.Printf("Error with partitioning: %v\n", err)
		}
	case modeSimple:
		err = metaSimple(blockMeta, config)
		if err != nil {
			fmt.Printf("Error with restore: %v\n", err)
		}
	}

	return err
}

// metaCustom run the `custom` commands for partitioning
func metaCustom(blockMeta PartitionCommand, config Config) error {
	for _, item := range config.Storage.Config {
		if err := metaCustomItem(blockMeta, item); err != nil {
			return err
		}
	}
	return nil
}

// metaSimple run the `simple` commands for partitioning
func metaSimple(blockMeta PartitionCommand, config Config) error {
	// Set the target mount point, if it's not set
	if len(blockMeta.Target) == 0 {
		v, ok := os.LookupEnv(envTargetMountPoint)
		if !ok {
			return errors.New("Unable to find target. Use --target or set TARGET_MOUNT_POINT")
		}
		blockMeta.Target = v
	}

	return simpleHandler(blockMeta, config)
}

func metaCustomItem(blockMeta PartitionCommand, item StorageItem) error {
	switch item.Type {
	case "disk":
		return diskHandler(blockMeta, item)
	case "partition":
		return partitionHandler(blockMeta, item)
	case "format":
		return formatHandler(blockMeta, item)
	default:
		return fmt.Errorf("The storage type `%s` is not implemented", item.Type)
	}
}
