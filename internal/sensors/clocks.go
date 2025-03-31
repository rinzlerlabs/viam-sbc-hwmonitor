package sensors

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
)

func GetSysFsClock(ctx context.Context, path string) (int64, error) {
	return readIntFromFile(ctx, path)
}

func GetSysFsCpuPaths() ([]string, error) {
	paths, err := filepath.Glob("/sys/devices/system/cpu/cpu[0-9]*")
	if err != nil {
		return nil, err
	}
	validPaths := make([]string, 0)
	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}
		validPaths = append(validPaths, path)
	}

	return validPaths, nil
}

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
