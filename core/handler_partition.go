// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package core

import (
	"errors"
	"fmt"
	"os/exec"

	humanize "github.com/dustin/go-humanize"
)

func partitionHandler(blockMeta PartitionCommand, item StorageItem) error {
	fmt.Printf("Create the partition `%s%d`\n", item.Device, item.Number)
	// Check the device details
	if len(item.Device) == 0 {
		return errors.New("device must be set for partition to be created")
	}
	if len(item.Size) == 0 {
		return errors.New("size must be specified for partition to be created")
	}
	if item.Number == 0 {
		return errors.New("partition number must be specified for partition to be created")
	}

	// Define the partition path using the different formats
	ptnPath := DiskPath{
		Device:       item.Device,
		DevicePath:   devicePathFromDevice(item.Device),
		SysBlockPath: sysBlockFromDevice(item.Device),
	}
	ptnName := fmt.Sprintf("%s%d", ptnPath.Device, item.Number) // sdd1

	lbs := logicalBlockSize(ptnPath.SysBlockPath)

	// Get the previous partition details
	var (
		prevSizeSectors  int
		prevStartSectors int
		offsetSectors    int
	)

	// Get the partition start point if this is not the first partition
	if item.Number > 1 {
		prevPtnName := fmt.Sprintf("%s%d", ptnPath.Device, item.Number-1) // sdd1
		prevPtn := sysBlockFromDevice(prevPtnName)                        // /sys/class/block/sdd1
		prevSize := partitionSize(prevPtn)
		prevStart := partitionStart(prevPtn)
		prevSizeSectors = int(prevSize * 512 / lbs)
		prevStartSectors = int(prevStart * 512 / lbs)
	}

	alignmentOffset := int((1 << 20)) / lbs
	if item.Number == 1 {
		offsetSectors = alignmentOffset
	} else {
		// We're only handling GPT and physical partitions (not extended/logical ones)
		offsetSectors = prevStartSectors + prevSizeSectors
	}

	// Start sector is part of the sectors that define the partitions size so length has to be "size in sectors - 1"
	lengthBytes, err := humanize.ParseBytes(item.Size)
	if err != nil {
		return err
	}
	lengthSectors := int(int(lengthBytes)/lbs) - 1

	if len(item.Wipe) > 0 {
		fmt.Printf("  Not implemented: ignore the `wipe` flag for partition `%s%d`", ptnPath.Device, item.Number)
	}

	// Run sgdisk to create the new partition
	fmt.Printf("  Create partition `%s` starting at sector %d-%d\n", ptnName, offsetSectors, lengthSectors+offsetSectors)
	typeCode := sgDiskFlag(item.Flag)
	out, err := exec.Command("sgdisk", "--new",
		fmt.Sprintf("%d:%d:%d", item.Number, offsetSectors, lengthSectors+offsetSectors),
		fmt.Sprintf("--typecode=%d:%s", item.Number, typeCode),
		ptnPath.DevicePath).Output()
	if err != nil {
		return fmt.Errorf("  Error formatting partition `%s`: %v", ptnName, err)
	}
	fmt.Printf("  Partition `%s` formatted: %s\n", ptnName, out)
	return nil
}
