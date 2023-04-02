//go:build linux
// +build linux

package resources

import (
	"errors"
	"xtest/util"
)

func LookUpGpu(pid int) (gpu string, err error) {
	processes, err := util.NVMLGpuProcesses()
	if err != nil {
		return "", errors.New("NVML lib errors! ")
	}
	for _, p := range processes {
		if p.Pid == pid {
			gpu = p.UsedMemory
		}
	}
	return
}
