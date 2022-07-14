package prometheus

import (
	"errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
	utilProcess "github.com/shirou/gopsutil/process"
	"net/http"
	"sync"
	"time"
	"xtest/util"
	_var "xtest/var"
)

type Resource struct {
	Mem  float64
	Cpu  float64
	Time float64
}

type Resources struct {
	resourceChan chan Resource
	resources    []Resource
	stopChan     chan bool
	wg           sync.WaitGroup
}

func NewResources() Resources {
	return Resources{
		resourceChan: make(chan Resource, 10000),
		resources:    []Resource{},
		stopChan:     make(chan bool, 100),
		wg:           sync.WaitGroup{},
	}
}

// Serve 启动Prometheus监听
func (rs *Resources) Serve() {
	server := http.NewServeMux()
	server.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2117", server)
}

// ReadMem 获取内存使用, 传入AiService的PID
func (rs *Resources) ReadMem(pid int) error {
	x, err := utilProcess.NewProcess(int32(pid))
	if err != nil {
		return errors.New("Pid Not Found! ")
	}
	memPer, _ := x.MemoryPercent()
	cpuPer, _ := x.CPUPercent()
	rs.resourceChan <- Resource{
		Mem:  float64(memPer),
		Cpu:  cpuPer,
		Time: float64(time.Now().UnixMicro()),
	}
	//fmt.Println(cpuPer, memPer)
	_var.CpuPer.Set(cpuPer)
	_var.MemPer.Set(float64(memPer))
	return nil
}

func (rs *Resources) GenerateData() {
	rs.wg.Add(1)
	go func() {
		for {
			select {
			case resource := <-rs.resourceChan:
				rs.resources = append(rs.resources, resource)
			case <-rs.stopChan:
				rs.wg.Done()
				return
			}
		}
	}()
}

// MetricValue 获取metric的Value值
func (rs *Resources) MetricValue(m prometheus.Gauge) (float64, error) {
	metric := dto.Metric{}
	err := m.Write(&metric)
	if err != nil {
		return 0, err
	}
	val := metric.Gauge.GetValue()
	return val, nil
}

// Draw 绘制图片
func (rs *Resources) Draw(dst string) error {
	n := len(rs.resources)
	cpus := make([]float64, n)
	mems := make([]float64, n)
	times := make([]float64, n)
	for i, r := range rs.resources {
		cpus[i] = r.Cpu
		mems[i] = r.Mem
		times[i] = r.Time
	}
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

func (rs *Resources) Stop() {
	rs.stopChan <- true
	rs.wg.Wait()
}

// bToMb bit转Mb
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
