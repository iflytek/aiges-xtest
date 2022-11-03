//go:build !linux
// +build !linux

package resources

func LookUpGpu(pid int) (gpu string, err error) {
	return "", nil
}
