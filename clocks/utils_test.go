package clocks

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFoo(t *testing.T) {
	ctx := context.Background()
	clocks, err := getSystemClocks(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, clocks)
	t.Logf("%v", clocks)
}
