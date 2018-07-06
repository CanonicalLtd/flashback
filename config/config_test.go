// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package config_test

import (
	"testing"

	"github.com/CanonicalLtd/flashback/config"
	check "gopkg.in/check.v1"
)

type SuiteTest struct {
	path    string
	success bool
}

type configSuite struct{}

var _ = check.Suite(&configSuite{})

func TestConfig(t *testing.T) { check.TestingT(t) }

func (s *configSuite) TestRead(c *check.C) {
	tests := []SuiteTest{
		{"../example.yaml", true},
		{"bad path", false},
		{"../README.md", false},
	}

	for _, t := range tests {
		err := config.Read(t.path)
		if t.success {
			c.Assert(err, check.IsNil)
		} else {
			c.Assert(err, check.NotNil)
		}
	}
}
