package main

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestUpdateConfiguration(t *testing.T) {
	testCases := []struct {
		Dir string
	}{
		{
			Dir: t.TempDir(),
		},
		{
			Dir: "../tests/configuration",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Dir, func(t *testing.T) {
			assert.NilError(t, UpdateConfig(tc.Dir))
		})
	}

}
