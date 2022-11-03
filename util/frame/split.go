//go:build !linux
// +build !linux

package frameUtil

import "log"

func GetH264Frames(video []byte) (frameSizes []int) {
	// todo mac/windows do what
	log.Println("Not implement in this os...")
	return
}
