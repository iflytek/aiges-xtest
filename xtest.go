package main

import (
	"fmt"
	"log"
	"sync"
	"time"
	"xtest/analy"
	"xtest/conf"
	"xtest/request"
	"xtest/resources"
	"xtest/util"

	"github.com/pterm/pterm"
	xsfcli "github.com/xfyun/xsf/client"
)

type Xtest struct {
	r   request.Request
	cli *xsfcli.Client
}

func NewXtest(cli *xsfcli.Client, c conf.Conf) Xtest {
	return Xtest{r: request.Request{C: c}, cli: cli}
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
	x.cli.Log.Debugw("dropThr", "length of dropThr", x.r.C.DropThr)
	for i := 0; i < x.r.C.DropThr; i++ {
		rwg.Add(1)
		go x.r.DownStreamWrite(&rwg, x.cli.Log)
	}

	var wg sync.WaitGroup

	//
	r := resources.NewResources()      // 开启资源监听实例
	stp := util.NewScheduledTaskPool() // 开启一个定时任务池
	if x.r.C.PrometheusSwitch {
		go func() {
			if err := r.Serve(x.r.C.PrometheusPort); err != nil {
				log.Printf("failed to serve prometheus metrics: %v", err)
			}
		}()
	}

	if x.r.C.Plot {
		r.GenerateData()
	}

	go util.ProgressShow(x.r.C.LoopCnt, x.r.C.LoopCnt.Load())

	x.cli.Log.Debugw("multiThr", "length of multiThr", x.r.C.MultiThr)
	for i := 0; i < x.r.C.MultiThr; i++ {
		wg.Add(1)
		go func() {
			for {
				loopIndex := x.r.C.LoopCnt.Load()
				x.r.C.LoopCnt.Dec()
				if x.r.C.LoopCnt.Load() < 0 {
					break
				}

				switch x.r.C.ReqMode {
				case 0:
					info := x.r.OneShotCall(x.cli, loopIndex)
					analy.ErrAnalyser.PushErr(info)
				case 1:
					info := x.r.SessionCall(x.cli, loopIndex) // loopIndex % len(stream.dataList)
					analy.ErrAnalyser.PushErr(info)
				case 2:
					info := x.r.TextCall(x.cli, loopIndex) // loopIndex % len(stream.dataList)
					analy.ErrAnalyser.PushErr(info)
				case 3:
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
	// 持久化资源日志
	if err := r.Dump(); err != nil {
		x.cli.Log.Errorw("resource dump failed", "err", err)
	}
	if x.r.C.Plot {
		if err := r.Draw(x.r.C.PlotFile); err != nil {
			x.cli.Log.Errorw("draw plot file failed", "err", err)
		}
	}
	pterm.DefaultBasicText.Println(pterm.LightGreen("\ncli finish "))
}

func (x *Xtest) linearCtl() {
	if x.r.C.LinearNs > 0 {
		time.Sleep(time.Nanosecond * time.Duration(x.r.C.LinearNs))
	}
}
