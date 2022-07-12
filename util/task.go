package util

import (
	"fmt"
	"time"
)

var (
	TaskNum      = 0
	TimeStopChan = make(chan int, 100) // 通知定时任务结束协程
)

// ScheduledTask 启动一个定时任务 jbzhou5
func ScheduledTask(d time.Duration, f func()) {
	TaskNum++
	go func() {
		ticker := time.NewTicker(d)
		for {
			select {
			case <-ticker.C:
				f()
			case x := <-TimeStopChan:
				fmt.Println("关闭咯: ", x)
				return
			}
		}
	}()
}

// StopTask 结束定时任务
func StopTask() {
	for i := 0; i < TaskNum<<1; i++ {
		TimeStopChan <- i
	}
}
