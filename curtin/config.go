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

	BlockMeta struct {
		Devices        []string        `yaml:"devices"`
		BootPartitions []BootPartition `yaml:"boot-partition"`
	} `yaml:"block-meta"`

	Sources []interface{} `yaml:"sources"`
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

// BootPartition defines the details of a boot partition
type BootPartition struct {
	Enabled bool   `yaml:"enabled"`
	Format  string `yaml:"format"`
	FsType  string `yaml:"fstype"`
	Label   string `yaml:"label"`
}

// SourceItem defines a source image
type SourceItem struct {
	DDImg string `yaml:"dd-img"`
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
