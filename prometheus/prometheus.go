package prometheus

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	utilProcess "github.com/shirou/gopsutil/process"
	"net/http"
	"syscall"
	_var "xtest/var"
)

func Start() {
	server := http.NewServeMux()
	server.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2117", server)
}

// ReadMem 获取内存使用
func ReadMem() {
	pid := syscall.Getpid() // 获取xtest的运行id
	x, _ := utilProcess.NewProcess(int32(pid))
	memPer, _ := x.MemoryPercent()
	cpuPer, _ := x.CPUPercent()
	_var.CpuPer.Set(cpuPer)
	_var.MemPer.Set(float64(memPer))
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

//func main() {
//	Start()
//}
