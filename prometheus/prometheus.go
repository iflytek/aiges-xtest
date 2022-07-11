package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
	utilProcess "github.com/shirou/gopsutil/process"
	"net/http"
	"syscall"
	"time"
	"xtest/util"
	_var "xtest/var"
)

var (
	times []float64
	cpus  []float64
	mems  []float64
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

// bToMb bit转Mb
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

// MetricValue 获取metric的Value值
func MetricValue(m prometheus.Gauge) (float64, error) {
	metric := dto.Metric{}
	err := m.Write(&metric)
	if err != nil {
		return 0, err
	}
	val := metric.Gauge.GetValue()
	return val, nil
}

// GenerateData 获取资源数据绘制折线图
func GenerateData(cv, mv float64) {
	times = append(times, float64(time.Now().UnixMicro()))
	cpus = append(cpus, cv)
	mems = append(mems, mv)
}

// Run 绘制图片
func Run(dst string) error {
	c := util.Charts{
		Vals: util.LinesData{
			Title: "Resource Record",
			BarValues: []util.LineYValue{
				{"cpus", cpus},
				{"mem", mems},
			},
		},
		Dst:     dst,
		XValues: times,
	}
	err := c.Draw()
	if err != nil {
		return err
	}
	return nil
}
