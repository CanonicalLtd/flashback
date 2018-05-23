// -*- Mode: Go; indent-tabs-mode: t -*-
// Curtin Core
// Copyright 2018 Canonical Ltd.  All rights reserved.

package curtin

import (
	"fmt"
)

const (
	modeCustom     = "custom"
	modeSimple     = "simple"
	modeSimpleBoot = "simple-boot"
)

// PartitionCommand defines the execution options for partitioning
type PartitionCommand struct {
	Device     string `short:"d" long:"device" description:"read configuration from cfg"`
	FSType     string `long:"fstype" description:"root partition filesystem type" choice:"ext4" choice:"ext3" default:"ext4"`
	BootFSType string `long:"boot-fstype" description:"boot partition filesystem type" choice:"ext4" choice:"ext3"`
	Target     string `short:"t" long:"target" description:"chroot to target. default is env[TARGET_MOUNT_POINT]"`
	UMount     bool   `long:"umount" description:"unmount any mounted filesystems before exit"`
	Mode       string `positional_arg_name:"mode" description:"meta-mode to use"`
}

// Partition run a partitioning command
func Partition(blockMeta PartitionCommand, config Config) error {
	if blockMeta.Mode == modeCustom {
		return metaCustom(blockMeta, config)
	}

	return nil
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
