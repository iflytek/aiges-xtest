package main

import (
	"flag"
	"fmt"
	xsfcli "git.iflytek.com/AIaaS/xsf/client"
	"git.iflytek.com/AIaaS/xsf/utils"
	"github.com/pterm/pterm"
	"sync"
	"time"
	"xtest/analy"
	"xtest/prometheus"
	"xtest/request"
	"xtest/util"
	"xtest/var"
)

func main() {
	flag.Parse()
	// xrpc框架初始化;
	cli, e := xsfcli.InitClient(_var.CliName, utils.CfgMode(0), utils.WithCfgName(*_var.CmdCfg),
		utils.WithCfgURL(""), utils.WithCfgPrj(""), utils.WithCfgGroup(""),
		utils.WithCfgService(""), utils.WithCfgVersion(""))
	if e != nil {
		fmt.Println("cli xsf init fail with ", e.Error())
		return
	}

	// cli配置初始化;
	e = _var.ConfInit(cli.Cfg())
	if e != nil {
		fmt.Println("cli conf init fail with ", e.Error())
		return
	}
	x := NewXtest(cli)
	x.Run()
	return
}

type Xtest struct {
	cli *xsfcli.Client
}

func NewXtest(cli *xsfcli.Client) Xtest {
	return Xtest{cli}
}

func (x *Xtest) Run() {
	// 数据分析初始化、性能数据
	analy.ErrAnalyser.Start(_var.MultiThr, x.cli.Log)
	if _var.PerfConfigOn {
		analy.Perf = new(analy.PerfModule)
		analy.Perf.Log = x.cli.Log
		startErr := analy.Perf.Start()
		if startErr != nil {
			fmt.Println("failed to open req record file.", startErr.Error())
			return
		}
		defer analy.Perf.Stop()
	}
	// 启动异步输出打印&落盘
	var rwg sync.WaitGroup
	for i := 0; i < _var.DropThr; i++ {
		rwg.Add(1)
		go request.DownStreamWrite(&rwg, x.cli.Log)
	}

	var wg sync.WaitGroup

	// jbzhou5
	r := prometheus.NewResources()     // 开启资源监听实例
	stp := util.NewScheduledTaskPool() // 开启一个定时任务池
	if _var.PrometheusSwitch {
		go r.Serve() // jbzhou5 启动一个协程写入Prometheus
	}

	if _var.Plot {
		r.GenerateData()
	}

	// 启动一个系统资源定时任务
	stp.Start(time.Microsecond*50, func() {
		err := r.ReadMem(_var.ServicePid)
		if err != nil {
			return
		}
	})

	go util.ProgressShow(_var.LoopCnt)

	for i := 0; i < _var.MultiThr; i++ {
		wg.Add(1)
		go func() {
			for {
				if _var.LoopCnt.Load() <= 0 {
					break
				}
				switch _var.ReqMode {
				case 0:
					loopIndex := _var.LoopCnt.Load()
					info := request.OneShotCall(x.cli, loopIndex)
					_var.LoopCnt.Dec()
					analy.ErrAnalyser.PushErr(info)
				case 1:
					loopIndex := _var.LoopCnt.Load()
					info := request.SessionCall(x.cli, loopIndex) // loopIndex % len(stream.dataList)
					_var.LoopCnt.Dec()
					analy.ErrAnalyser.PushErr(info)
				case 2:
					loopIndex := _var.LoopCnt.Load()
					info := request.TextCall(x.cli, loopIndex) // loopIndex % len(stream.dataList)
					_var.LoopCnt.Dec()
					analy.ErrAnalyser.PushErr(info)
				case 3:
					loopIndex := _var.LoopCnt.Load()
					info := request.FileSessionCall(x.cli, loopIndex) // loopIndex % len(stream.dataList)
					_var.LoopCnt.Dec()
					analy.ErrAnalyser.PushErr(info)
				default:
					println("Unsupported Mode!")
				}
			}
			wg.Done()
		}()
		linearCtl() // 并发线性增长控制,防止瞬时并发请求冲击
	}
	wg.Wait()
	// 关闭异步落盘协程&wait
	close(_var.AsyncDrop)
	analy.ErrAnalyser.Stop()
	rwg.Wait()
	xsfcli.DestroyClient(x.cli)
	stp.Stop() // 关闭定时任务
	r.Stop()   // 关闭资源收集
	r.Draw(_var.PlotFile)
	pterm.DefaultBasicText.Println(pterm.LightGreen("\ncli finish "))
}

func linearCtl() {
	if _var.LinearNs > 0 {
		time.Sleep(time.Duration(time.Nanosecond) * time.Duration(_var.LinearNs))
	}
}
