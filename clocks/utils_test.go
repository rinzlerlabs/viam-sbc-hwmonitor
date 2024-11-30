package clocks

import (
	"context"
	"testing"

	. "github.com/rinzlerlabs/sbcidentify/test"
	"github.com/stretchr/testify/assert"
)

func TestGetSystemClocks(t *testing.T) {
	Test().RequiresRoot().RequiresSbc().ShouldSkip(t)
	ctx := context.Background()
	clocks, err := getSystemClocks(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, clocks)
	t.Logf("%v", clocks)
}
