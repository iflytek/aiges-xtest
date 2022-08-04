package resources

import (
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
	utilProcess "github.com/shirou/gopsutil/process"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"xtest/util"
	_var "xtest/var"
)

type Resource struct {
	Mem  float64
	Cpu  float64
	Gpu  string
	Time float64
}

const (
	outputResourceFile = "./log/resource.csv"
)

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
func (rs *Resources) Serve(port int) error {
	server := http.NewServeMux()
	server.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), server)
	return err
}

// ReadMem 获取内存使用, 传入AiService的PID
func (rs *Resources) ReadMem(c *_var.Conf) (err error) {
	taddrs := c.Taddrs
	port, err := strconv.Atoi(strings.Split(taddrs, ":")[1])
	if err != nil {
		return  err
	}
	pid,  err := rs.DetectPort(port)
	if err != nil {
		return err
	}
	x, err := utilProcess.NewProcess(int32(pid))
	if err != nil {
		return errors.New("Pid Not Found! ")
	}
	processes, err := util.GpuProcesses()
	if err != nil {
		return errors.New("Nvidia-smi errors! ")
	}
	var gpu string
	for _, p := range processes {
		if p.Pid == pid {
			gpu = p.UsedMemory
		}
	}
	memPer, _ := x.MemoryPercent()
	cpuPer, _ := x.CPUPercent()

	c.CpuPer.Set(cpuPer)
	c.MemPer.Set(float64(memPer))
	r := Resource{
		Mem:  float64(memPer),
		Cpu:  cpuPer,
		Gpu: gpu,
		Time: float64(time.Now().UnixMicro()),
	}
	rs.resourceChan <- r
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

// Dump 持久化日志
func (rs *Resources) Dump() error {
	f, err := os.OpenFile(outputResourceFile, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	_, err = f.WriteString("CPU,MEMORY,GPU,TIME\n")
	if err != nil {
		return err
	}
	for _, r := range rs.resources {
		_, err = f.WriteString(fmt.Sprintf("%f,%f,%s,%s\n", r.Cpu, r.Mem, r.Gpu, time.UnixMicro(int64(r.Time)).Format("2006-01-02 15:04:05.000")))
		if err != nil {
			return err
		}
	}
	err = f.Close()
	if err != nil {
		return err
	}
	return nil
}

// DetectPort 通过端口号得出进程
func (rs *Resources) DetectPort(port int)  (int, error) {
	var pid int
	var err error
	for _, tcp := range util.Tcp() {
		if tcp.Port == int64(port) && tcp.State == "LISTEN"{
			pid, err = strconv.Atoi(tcp.Pid)
			if err != nil {
				return 0, err
			}
			break
		}
	}
	if pid == 0 {
		return 0, errors.New(fmt.Sprintf("No process listening port: ", port))
	}
	return pid, nil
}

// bToMb bit转Mb
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
