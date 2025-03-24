//go:build linux && full_nvidia_support
// +build linux,full_nvidia_support

package utils

import (
	"errors"
	"sync"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

var (
	NVMLManager = &nvmlManager{
		initLock:        sync.Mutex{},
		nvmlInitialized: false,
	}
)

type nvmlManager struct {
	initLock        sync.Mutex
	nvmlInitialized bool
	userCount       int
}

func (manager *nvmlManager) Acquire() error {
	manager.initLock.Lock()
	defer manager.initLock.Unlock()

	if !manager.nvmlInitialized {
		ret := nvml.Init()
		if ret != nvml.SUCCESS {
			return errors.Join(ret, errors.New("failed to initialize NVML"))
		}
		manager.nvmlInitialized = true
	}
	manager.userCount++

	return nil
}

func (manager *nvmlManager) Release() error {
	manager.initLock.Lock()
	defer manager.initLock.Unlock()

	manager.userCount--
	if manager.userCount == 0 {
		ret := nvml.Shutdown()
		if ret != nvml.SUCCESS {
			return errors.Join(ret, errors.New("failed to shutdown NVML"))
		}
		manager.nvmlInitialized = false
	}

	return nil
}
