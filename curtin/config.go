// -*- Mode: Go; indent-tabs-mode: t -*-
// Curtin Core
// Copyright 2018 Canonical Ltd.  All rights reserved.

package curtin

import (
	"io/ioutil"
	"log"

	yaml "gopkg.in/yaml.v2"
)

// Config defines the configuration parameters
type Config struct {
	PartitioningCommands map[string]string `yaml:"partitioning_commands"`
	EarlyCommands        map[string]string `yaml:"early_commands"`
	NetworkCommands      map[string]string `yaml:"network_commands"`

	Storage struct {
		Config  []StorageItem `yaml:"config"`
		Version int           `yaml:"version"`
	} `yaml:"storage"`
}

// StorageItem defines a storage element
type StorageItem struct {
	ID         string `yaml:"id"`
	Type       string `yaml:"type"`
	PTable     string `yaml:"ptable,omitempty"`
	Path       string `yaml:"path,omitempty"`
	GrubDevice bool   `yaml:"grub_device,omitempty"`
	Preserve   bool   `yaml:"preserve,omitempty"`
	Number     int    `yaml:"number,omitempty"`
	Device     string `yaml:"device,omitempty"`
	Size       string `yaml:"size,omitempty"`
	Flag       string `yaml:"flag,omitempty"`
	FsType     string `yaml:"fstype,omitempty"`
	Wipe       string `yaml:"wipe,omitempty"`
	Label      string `yaml:"label,omitempty"`
	Volume     string `yaml:"volume,omitempty"`
}

// ReadConfig fetches the store config
func ReadConfig(path string) (Config, error) {
	c := Config{}

	dat, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("Error reading config parameters: %v", err)
		return c, err
	}

	err = yaml.Unmarshal(dat, &c)
	if err != nil {
		log.Printf("Error parsing config parameters: %v", err)
		return c, err
	}

	return c, nil
}
