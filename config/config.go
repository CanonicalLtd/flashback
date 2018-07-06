// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package config

import (
	"fmt"
	"io/ioutil"

	"github.com/CanonicalLtd/flashback/audit"
	yaml "gopkg.in/yaml.v2"
)

// Config defines the configuration parameters
type Config struct {
	Backup struct {
		Size int      `yaml:"size"`
		Data []string `yaml:"data"`
	} `yaml:"retain"`
}

// Default constants
const (
	defaultBackupSize      = 32
	restorePartitionLabel  = "restore"
	writablePartitionLabel = "writable"
	LogFileBootprint       = "/var/log/flashback/bootprint.log"
	LogFileReset           = "/var/log/flashback/reset.log"
)

// Store the stored configuration from the file
var Store Config

// Read parses the yaml config file
func Read(path string) error {
	Store = Config{}

	dat, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading config parameters: %v\n", err)
		return err
	}

	err = yaml.Unmarshal(dat, &Store)
	if err != nil {
		fmt.Printf("Error parsing config parameters: %v\n", err)
		return err
	}

	// Default the missing parameters
	setDefaults()

	return nil
}

func setDefaults() {
	if Store.Backup.Size <= 0 {
		audit.Printf("Default the retained data size to `%d`\n", defaultBackupSize)
		Store.Backup.Size = defaultBackupSize
	}
}
