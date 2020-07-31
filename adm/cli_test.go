package main

import (
	"testing"

	"github.com/aluzme/go-admin/modules/system"
	"github.com/magiconair/properties/assert"
)

func TestGetLatestVersion(t *testing.T) {
	assert.Equal(t, getLatestVersion(), system.Version())
}
