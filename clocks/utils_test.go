package clocks

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	. "github.com/thegreatco/gotestutils"
)

func TestGetSystemClocks(t *testing.T) {
	Test().RequiresSbc().ShouldSkip(t)
	ctx := context.Background()
	clocks, err := getSystemClocks(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, clocks)
	t.Logf("%v", clocks)
}
