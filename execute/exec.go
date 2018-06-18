// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package execute

// Command defines the execution options for the application
type Command struct {
	ConfigPath   string `short:"c" long:"config" description:"read configuration from cfg"`
	FactoryReset bool   `long:"factory-reset" description:"run a factory reset of the device"`
	Bootprint    bool   `long:"bootprint" description:"create a recovery image for the device"`
	Check        bool   `long:"check" description:"check that a recovery image does not exist (used with the --bootprint option)"`
}

// Execution is the implementation of the execution options
var Execution Command
