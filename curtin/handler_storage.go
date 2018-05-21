// -*- Mode: Go; indent-tabs-mode: t -*-
// Curtin Core
// Copyright 2018 Canonical Ltd.  All rights reserved.

package curtin

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

const (
	pTableGPT          = "gpt"
	pTableDOS          = "dos"
	pTableMSDOS        = "msdos"
	wipeModeSuperblock = "superblock"
	zeroBufLen         = 1024
	zeroCount          = 1024
	zeroSize           = zeroBufLen * zeroCount
)

func diskHandler(blockMeta PartitionCommand, item StorageItem) error {
	if item.Preserve {
		fmt.Println("Not implemented: ignore the `preserve` flag")
	}

	if !item.Preserve && len(item.Wipe) > 0 {
		fmt.Println("Not implemented: wipe the disk")
	}

	if !item.Preserve && len(item.PTable) > 0 {
		if item.PTable == pTableGPT {
			// Wipe both MBR and GPT that may be present on the disk.
			// n.b. wipe_volume wipes 1M at front and end of the disk.
			// This could destroy disk data in filesystems that lived
			// there.
			// Note: using the `path` instead of the `id`
			fmt.Println("---wipeVolume", item.PTable, item.Path)
			if err := wipeVolume(item.Path, wipeModeSuperblock, true); err != nil {
				fmt.Println(err)
				return err
			}
		}
		fmt.Println("Not implemented: label the disk")
	}

	return nil
}

// wipeVolume wipes a volume/block device
// path: a path to a block device
// mode: how to wipe it.
//	pvremove: wipe a lvm physical volume
//	zero: write zeros to the entire volume
//	random: write random data (/dev/urandom) to the entire volume
//	superblock: zero the beginning and the end of the volume
//	superblock-recursive: zero the beginning of the volume, the end of the
//		volume and beginning and end of any partitions that are known to be on this device.
//	exclusive: boolean to control how path is opened
func wipeVolume(path, mode string, exclusive bool) error {
	if mode == wipeModeSuperblock {
		return quickZero(path, false, exclusive)
	}
	return fmt.Errorf("Wipe not implemented for mode `%s`", mode)
}

// quickZero zeroes 1M at front, 1M at end, and 1M at front if this is a block device and
// partitions is true, then zero 1M at front and end of each partition.
func quickZero(path string, partitions, exclusive bool) error {
	fmt.Println("---quickZero", path)
	//offsets := []int{0, -zeroSize}

	// Check this path is a block device or file
	isBlk, err := isBlockDevice(path)
	if err != nil {
		return err
	}
	isFilePath := false
	if _, err := os.Stat(path); os.IsNotExist(err) {
		isFilePath = true
	}
	if !isBlk && !isFilePath {
		return fmt.Errorf("%s: not an existing block device", path)
	}

	if partitions {
		fmt.Println("Zeroing partitions is not implemented")
		deviceName := filepath.Base(path)
		partitions, err := sysfsPartitions(deviceName)
		if err != nil {
			fmt.Printf("Error retrieving partitions for the device: %v", err)
			return err
		}

		for _, p := range partitions {
			partPath := filepath.Join("/dev", p)
			fmt.Println("---zero partition", partPath)
			// if err = quickZero(partPath, false, false); err != nil {
			// 	return err
			// }
		}
	}

	fmt.Println("---isBlockDevice", "YES!")

	// err = zeroFileAtOffsets(path, offsets, zeroBufLen, zeroCount, false, exclusive)
	// if err != nil {
	// 	fmt.Printf("Error zeroing the path `%s`: %v\n", path, err)
	// 	return err
	// }
	fmt.Printf("Successfully zeroed path `%s`\n", path)
	return nil
}

func isBlockDevice(path string) (bool, error) {
	var stat syscall.Stat_t

	err := syscall.Stat(path, &stat)
	if err != nil {
		fmt.Printf("Error checking block device: %v\n", err)
		return false, err
	}

	// Check if this is a block device
	if (stat.Mode & syscall.S_IFMT) == syscall.S_IFBLK {
		return true, nil
	}

	return false, nil
}

// zeroFileAtOffsets writes zeroes to file at specified offsets
func zeroFileAtOffsets(path string, offsets []int, buflen, count int, strict, exclusive bool) error {
	f, err := exclusiveOpen(path, exclusive)
	if err != nil {
		return err
	}

	// Get the size by seeking to the end
	size, err := f.Seek(0, os.SEEK_END)
	if err != nil {
		return err
	}

	// Create an zero-ised buffer of the write length
	buf := make([]byte, buflen)

	for _, offset := range offsets {
		// Position the pointer at the offset position
		pos := int64(offset)
		if offset < 0 {
			pos = pos + size
		}
		if pos > size || pos < 0 {
			return fmt.Errorf("invalid file offset: %d", offset)
		}
		if pos+zeroSize > size {
			return fmt.Errorf("shortened to size: %d", size)
		}
		pos, err := f.Seek(pos, os.SEEK_SET)
		if err != nil {
			return err
		}

		// Write the zero-ised buffer the required number of times
		for i := 0; i < count; i++ {
			if pos+int64(buflen) > size {
				f.Write(buf[:size-pos])
			} else {
				f.Write(buf)
			}
			pos, err = f.Seek(0, os.SEEK_CUR)
		}
	}

	return nil
}

// exclusiveOpen obtains an exclusive file-handle to the file/device specified
// unless caller specifics exclusive=False.
func exclusiveOpen(path string, exclusive bool) (*os.File, error) {
	flags := os.O_RDWR
	if exclusive {
		flags = flags | os.O_EXCL
	}

	return os.OpenFile(path, flags, 0644)
}
