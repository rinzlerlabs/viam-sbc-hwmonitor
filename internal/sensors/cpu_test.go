package sensors

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetProcCmdlineParseing(t *testing.T) {
	data := "/root/.viam/packages-local/data/module/synthetic-rinzlerlabs_sbc-hwmonitor_from_reload-0_0_21/bin/rinzlerlabs-sbc-hwmonitor\x00/tmp/viam-module-2615955240/rinzlerlabs_sbc-hwmonitor_from_reload-sJYoJ.sock\x00"
	// The cmdline is null-separated, so split it and return the first argument
	args := strings.Split(string(data), "\x00")
	if len(args) > 0 {

		fmt.Printf("First argument in cmdline: %s\n", filepath.Base(args[0]))
		return
	}
	t.FailNow()
}
