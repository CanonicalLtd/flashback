// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package execute

// Command defines the execution options for the application
type Command struct {
	ConfigPath   string `short:"c" long:"config" description:"read configuration from cfg"`
	FactoryReset bool   `long:"factory-reset" description:"read configuration from cfg"`
}

// Execution is the implementation of the execution options
var Execution Command
