package request

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// SigRegister 用于优雅退出测试
func (r *Request) SigRegister() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM)
	signal.Notify(sigChan, syscall.SIGINT)
	go func() {
		for {
			sig := <-sigChan
			switch sig {
			case syscall.SIGTERM:
				// 当前正在进行的请求或会话持续请求至正常结束, 剩余请求清零
				r.C.LoopCnt.Store(0)
			case syscall.SIGINT:
				//_var.LoopCnt.Store(0)
				fmt.Println("\nSIGINT\n")
				// TODO 可区别于SIGTERM, 当前进行的会话暴力结束,不计入统计数据,防止会话最长时间等待(eg: iat 60s)
			}
		}
	}()
}
