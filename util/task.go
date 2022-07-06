package util

import (
	"time"
)

// ScheduledTask 启动一个定时任务 jbzhou5
func ScheduledTask(d time.Duration, f func()) {
	go func() {
		ticker := time.NewTicker(d)
		for {
			select {
			case <-ticker.C:
				f()
			}
		}
	}()
}
