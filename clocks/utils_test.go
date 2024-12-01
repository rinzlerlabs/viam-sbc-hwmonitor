package clocks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSysFsCpuPaths(t *testing.T) {
	paths, err := getSysFsCpuPaths()
	assert.NoError(t, err)
	assert.NotEmpty(t, paths)
	t.Logf("%v", paths)
}
