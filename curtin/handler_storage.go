// -*- Mode: Go; indent-tabs-mode: t -*-
// Curtin Core
// Copyright 2018 Canonical Ltd.  All rights reserved.

package curtin

import (
	"fmt"
	"os"
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

// wipeVolume wipes a volume/block device
// path: a path to a block device
// mode: how to wipe it.
//	superblock: zero the beginning and the end of the volume
//	-- pvremove: wipe a lvm physical volume
//	-- zero: write zeros to the entire volume
//	-- random: write random data (/dev/urandom) to the entire volume
//	-- superblock-recursive: zero the beginning of the volume, the end of the
//		volume and beginning and end of any partitions that are known to be on this device.
//	exclusive: boolean to control how path is opened
func wipeVolume(diskPath DiskPath, mode string, exclusive bool) error {
	if mode == wipeModeSuperblock {
		return quickZeroSuperblock(diskPath, exclusive)
	}
	return fmt.Errorf("  Wipe not implemented for mode `%s`", mode)
}

// quickZeroSuperblock zeroes 1M at front, 1M at end, and 1M at front if this is a block device
func quickZeroSuperblock(diskPath DiskPath, exclusive bool) error {
	offsets := []int{0, -zeroSize}

	// Check this path is a block device or file
	isBlk, err := isBlockDevice(diskPath.DevicePath)
	if err != nil {
		return err
	}
	if !isBlk {
		return fmt.Errorf("  %s: not an existing block device", diskPath.DevicePath)
	}

	// Zero out the first and last 1M of the disk
	err = zeroFileAtOffsets(diskPath, offsets, zeroBufLen, zeroCount, false, exclusive)
	if err != nil {
		fmt.Printf("  Error zeroing the path `%s`: %v\n", diskPath.DevicePath, err)
		return err
	}
	fmt.Printf("  Successfully zeroed path `%s`\n", diskPath.DevicePath)
	return nil
}

func isBlockDevice(path string) (bool, error) {
	var stat syscall.Stat_t

	err := syscall.Stat(path, &stat)
	if err != nil {
		fmt.Printf("  Error checking block device: %v\n", err)
		return false, err
	}

	// Check if this is a block device
	if (stat.Mode & syscall.S_IFMT) == syscall.S_IFBLK {
		return true, nil
	}

	return false, nil
}

// zeroFileAtOffsets writes zeroes to file at specified offsets
func zeroFileAtOffsets(diskPath DiskPath, offsets []int, buflen, count int, strict, exclusive bool) error {
	f, err := exclusiveOpen(diskPath.DevicePath, exclusive)
	if err != nil {
		return err
	}
	defer f.Close()

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
