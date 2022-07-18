package main

import (
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
	f := _var.NewFlag()
	f.Parse()
	//if *f.XTestVersion {
	//	fmt.Println("2.5.2")
	//}
	// xrpc框架初始化;
	cli, e := xsfcli.InitClient(_var.CliName, utils.CfgMode(0), utils.WithCfgName(*f.CmdCfg),
		utils.WithCfgURL(""), utils.WithCfgPrj(""), utils.WithCfgGroup(""),
		utils.WithCfgService(""), utils.WithCfgVersion(""))
	if e != nil {
		fmt.Println("cli xsf init fail with ", e.Error())
		return
	}

	// cli配置初始化;
	conf := _var.NewConf()
	e = conf.ConfInit(cli.Cfg())
	if e != nil {
		fmt.Println("cli conf init fail with ", e.Error())
		return
	}
	//fmt.Printf("%+v\n", conf)
	x := NewXtest(cli, conf)
	x.Run()
	return
}

type Xtest struct {
	r   request.Request
	cli *xsfcli.Client
}

func NewXtest(cli *xsfcli.Client, conf _var.Conf) Xtest {
	return Xtest{r: request.Request{C: conf}, cli: cli}
}

func (x *Xtest) Run() {
	// 数据分析初始化、性能数据
	analy.ErrAnalyser.Start(x.r.C.MultiThr, x.cli.Log, x.r.C.ErrAnaDst)
	if x.r.C.PerfConfigOn {
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
	for i := 0; i < x.r.C.DropThr; i++ {
		rwg.Add(1)
		go x.r.DownStreamWrite(&rwg, x.cli.Log)
	}

	var wg sync.WaitGroup

	// jbzhou5
	r := prometheus.NewResources()     // 开启资源监听实例
	stp := util.NewScheduledTaskPool() // 开启一个定时任务池
	if x.r.C.PrometheusSwitch {
		go r.Serve() // jbzhou5 启动一个协程写入Prometheus
	}

	if x.r.C.Plot {
		r.GenerateData()
	}

	// 启动一个系统资源定时任务
	stp.Start(time.Microsecond*50, func() {
		err := r.ReadMem(x.r.C.ServicePid)
		if err != nil {
			return
		}
	})

	go util.ProgressShow(x.r.C.LoopCnt, x.r.C.LoopCnt.Load())

	for i := 0; i < x.r.C.MultiThr; i++ {
		wg.Add(1)
		go func() {
			for {
				loopIndex := x.r.C.LoopCnt.Load()
				if x.r.C.LoopCnt.Load() <= 0 {
					break
				}
				switch x.r.C.ReqMode {
				case 0:
					x.r.C.LoopCnt.Dec()
					info := x.r.OneShotCall(x.cli, loopIndex)
					analy.ErrAnalyser.PushErr(info)
				case 1:
					x.r.C.LoopCnt.Dec()
					info := x.r.SessionCall(x.cli, loopIndex) // loopIndex % len(stream.dataList)
					analy.ErrAnalyser.PushErr(info)
				case 2:
					x.r.C.LoopCnt.Dec()
					info := x.r.TextCall(x.cli, loopIndex) // loopIndex % len(stream.dataList)
					analy.ErrAnalyser.PushErr(info)
				case 3:
					x.r.C.LoopCnt.Dec()
					info := x.r.FileSessionCall(x.cli, loopIndex) // loopIndex % len(stream.dataList)
					analy.ErrAnalyser.PushErr(info)
				default:
					println("Unsupported Mode!")
				}
			}
			wg.Done()
		}()
		x.linearCtl() // 并发线性增长控制,防止瞬时并发请求冲击
	}
	wg.Wait()
	// 关闭异步落盘协程&wait
	close(x.r.C.AsyncDrop)
	analy.ErrAnalyser.Stop()
	rwg.Wait()
	xsfcli.DestroyClient(x.cli)
	stp.Stop() // 关闭定时任务
	r.Stop()   // 关闭资源收集
	r.Draw(x.r.C.PlotFile)
	pterm.DefaultBasicText.Println(pterm.LightGreen("\ncli finish "))
}

func (x *Xtest) linearCtl() {
	if x.r.C.LinearNs > 0 {
		time.Sleep(time.Duration(time.Nanosecond) * time.Duration(x.r.C.LinearNs))
	}
}
