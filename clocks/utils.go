package clocks

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/rinzlerlabs/viam-raspi-sensors/utils"
)

func readIntFromFile(ctx context.Context, path string) (int64, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()
	file, err := utils.ReadFileWithContext(ctxWithTimeout, path)
	if err != nil {
		return 0, err
	}
	i, err := strconv.ParseInt(strings.TrimSpace(string(file)), 10, 64)
	if err != nil {
		return 0, err
	}
	return i, nil
}
