package common

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestServerSideEncryptionConfiguration(t *testing.T) {
	assert.Assert(t, ServerSideEncryptionConfiguration{}.IsEmpty())
}
