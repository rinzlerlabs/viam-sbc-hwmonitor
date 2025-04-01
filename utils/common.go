package utils

import (
	"context"
	"errors"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
)

var (
	ErrBoardNotSupported    = errors.New("board not supported")
	ErrPlatformNotSupported = errors.New("platform not supported")
)

func ReadFileWithContext(ctx context.Context, path string) (string, error) {
	fileChan := make(chan []byte, 1)
	errChan := make(chan error, 1)
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	go func() {
		defer f.Close()
		data, err := io.ReadAll(f)
		if err != nil {
			errChan <- err
			return
		}
		fileChan <- data
	}()
	select {
	case <-ctx.Done():
		// Force close the file to unblock the ReadAll
		err = f.Close()
		if err != nil {
			return "", errors.Join(err, ctx.Err())
		}
		return "", ctx.Err()
	case data := <-fileChan:
		return strings.TrimSpace(string(data)), nil
	case err := <-errChan:
		return "", err
	}
}

func ReadInt64FromFileWithContext(ctx context.Context, path string) (int64, error) {
	data, err := ReadFileWithContext(ctx, path)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.TrimSpace(data), 10, 64)
}

func ReadBoolFromFileWithContext(ctx context.Context, path string) (bool, error) {
	data, err := ReadFileWithContext(ctx, path)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(data) == "1", nil
}

func ParseInt64(data string) (int64, error) {
	value, err := strconv.ParseInt(strings.TrimSpace(data), 10, 64)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func ParseFloat64(data string) (float64, error) {
	value, err := strconv.ParseFloat(strings.TrimSpace(data), 64)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func RoundValue(f float64, places int) float64 {
	shift := math.Pow(10, float64(places))
	return math.Round(f*shift) / shift
}
