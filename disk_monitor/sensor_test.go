package disk_monitor

import (
	"context"
	"testing"
	"time"

	. "github.com/rinzlerlabs/sbcidentify/test"
	"github.com/stretchr/testify/assert"
	"go.viam.com/rdk/logging"
)

func TestGetDiskInfo(t *testing.T) {
	Test().ShouldSkip(t)
	logger := logging.NewTestLogger(t)

	tests := []struct {
		name              string
		includeIOCounters bool
	}{
		{
			name:              "no_counters",
			includeIOCounters: false,
		},
		{
			name:              "with_counters",
			includeIOCounters: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			parts, err := getRealDisks(ctx)
			assert.NoError(t, err)
			assert.NotEmpty(t, parts)
			for _, part := range parts {
				logger.Infof("Disk device: %v mount: %v", part.Device, part.Mountpoint)
			}

			sensor := &Config{
				logger:            logger,
				disks:             parts,
				includeIOCounters: tt.includeIOCounters,
			}
			now := time.Now()
			readings, err := sensor.Readings(ctx, nil)
			elapsed := time.Since(now)
			logger.Infof("Elapsed time: %v", elapsed)
			assert.NoError(t, err)
			assert.NotNil(t, readings)
			assert.NotEmpty(t, readings)
			if tt.includeIOCounters {
				assert.Len(t, readings, len(parts)*17)
			} else {
				assert.Len(t, readings, len(parts)*4)
			}
			for k, v := range readings {
				logger.Infof("%v: %v", k, v)
			}
		})
	}
}

func TestGetDisksFromConfig(t *testing.T) {
	Test().RequiresSbc().ShouldSkip(t)
	tests := []struct {
		name           string
		disks          []string
		expectedCount  int
		expectedMount  string
		expectedDevice string
	}{
		{
			name:           "empty",
			disks:          []string{},
			expectedCount:  1,
			expectedMount:  "/",
			expectedDevice: "mmcblk0p2",
		},
		{
			name:           "slash",
			disks:          []string{"/"},
			expectedCount:  1,
			expectedMount:  "/",
			expectedDevice: "mmcblk0p2",
		},
		{
			name:           "slash_dev_slash_mmcblk0p2",
			disks:          []string{"/dev/mmcblk0p2"},
			expectedCount:  1,
			expectedMount:  "/",
			expectedDevice: "mmcblk0p2",
		},
		{
			name:           "mmcblk0p2",
			disks:          []string{"mmcblk0p2"},
			expectedCount:  1,
			expectedMount:  "/",
			expectedDevice: "mmcblk0p2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			conf := &ComponentConfig{
				Disks: tt.disks,
			}
			disks, err := getDisks(ctx, conf.Disks)
			assert.NoError(t, err)
			assert.Len(t, disks, tt.expectedCount)
			assert.Equal(t, tt.expectedDevice, disks[0].Device)
			assert.Equal(t, tt.expectedMount, disks[0].Mountpoint)
		})
	}
}
