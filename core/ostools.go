// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package core

import (
	"fmt"
	"os/exec"
	"strings"
)

// FindFS locates a filesystem by label
func FindFS(label string) (string, error) {
	out, err := exec.Command("findfs", fmt.Sprintf("LABEL=%s", label)).Output()

	// Remove non-printable characters from the response
	cleaned := cleanOutput(string(out))
	return cleaned, err
}

// FreePartitionSpace finds the free space on a device
func FreePartitionSpace(device string) (int, error) {
	out, err := exec.Command("parted", "-ms", device, "unit B print").Output()
	if err != nil {
		return 0, err
	}

	// Parse the output to get the unit number
	s := string(out)
	last := strings.Split(s, ":")

	return stringToInt(last[7])
}

// RelabelDisk changes the label of a disk
func RelabelDisk(device, label string) (string, error) {
	out, err := exec.Command("tune2fs", "-L", label, device).Output()

	// Remove non-printable characters from the response
	cleaned := cleanOutput(string(out))
	return cleaned, err
}
