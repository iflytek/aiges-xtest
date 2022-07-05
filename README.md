# Xtest测试框架

## 一、 目录总览

[TOC]

## 二、代码目录

### 2.1、根目录

#### Ⅰ.  xtest.go

> xtest为项目的入口，拥有main函数与linearCtl函数，
>
> main函数负责解析参数、初始化客户端，并实现多协程异步记录测试日志
>
> ```go
> // 函数负责并发线性增长控制,防止瞬时并发请求冲击
> func linearCtl()
> ```

### 2.2、analy文件夹

#### Ⅰ. errdist.go： 运行错误信息有关的结构体与函数定义

> ```go
> type errInfo struct {
> 	errCode int
> 	errStr  error
> }
> 
> // errDistAnalyser 记录错误数据相关的信息
> type errDistAnalyser struct {
> 	errCnt   map[int]int64 // map[error]count 错误计数
> 	errDsc   map[int]error // 错误描述
> 	errTmp   []errInfo     // 临时存储区,用于channel满阻塞的极端场景;
> 	errMutex sync.Mutex	   // 互斥锁
> 	errChan  chan errInfo // errorInfo管道
> 	swg      sync.WaitGroup // 并发同步原语
> 	log      *utils.Logger // 日志写入
> }
> 
> // 初始化 errDistAnalyser 结构体成员，并启动一个count计数协程计算errCnt
> func (eda *errDistAnalyser) Start(clen int, logger *utils.Logger) 
> // 将错误信息发送到errChan中，如果errchan满了，则发送到临时存储区errTmp
> func (eda *errDistAnalyser) PushErr(code int, err error)
> // 关闭errChan并等待swg 并发协程执行完成
> func (eda *errDistAnalyser) Stop()
> // 首先读取errChan错误信息，再读取errTmp临时存储区错误信息，分别统计不同错误的次数，并落盘
> func (eda *errDistAnalyser) count()
> // 错误分布数据落盘
> func (eda *errDistAnalyser) dumpLog()
> ```

#### Ⅱ. perf.go： 性能指标统计有关的函数定义

> ```go
> // 计时类型,用于控制计时开关
> const (
>  FIFOPERF = 1 << iota // 首结果耗时
>  LILOPERF             // 尾结果耗时
>  SESSPERF             // 会话耗时
>  INTFPERF             // 接口耗时
> )
> 
> // 计时定点,用于标记计时位置
> const (
>  pointCreate int = 1 << iota
>  pointUpBegin
>  pointDownBegin
>  pointUpEnd
>  pointDownEnd
>  pointDestroy
> )
> 
> type PerfDetail struct {
>  cTime     time.Time // create 时间
>  dTime     time.Time // destroy 时间
>  firstUp   time.Time
>  lastUp    time.Time
>  firstDown time.Time
>  lastDown  time.Time
>  upCost    []time.Time // 上行接口耗时
>  downCost  []time.Time // 下行接口耗时
> }
> 
> type perfDist struct {
>  level   int  // 性能统计等级, 最高：FIFOPERF | LILOPERF | SESSPERF | INTFPERF
>  details map[string] /*sid*/ PerfDetail // 分布数据需要保存全量会话数据
> 
> }
> 
> func (pc *perfDist) Start(perfLevel int)
> // TODO check type and point, 根据性能等级判定当前point是否需要获取时间
> // TODO write to channel
> func (pc *perfDist) TickPoint(point int)
> func (pc *perfDist) Stop()
> // TODO read from channel
> // lock map
> func (pc *perfDist) analysis()
> // 性能指标落盘
> func (pc *perfDist) perfDump()
> ```

#### Ⅲ performance.go

> ```go
> type direction int
> type DataStatus int
> type SessStatus int
> 
> const (
>     UP   direction = 1
>     DOWN direction = 2
> 
>     DataBegin    DataStatus = 0
>     DataContinue DataStatus = 1
>     DataEnd      DataStatus = 2
>     DataTotal    DataStatus = 3
> 
>     SessBegin    SessStatus = 0
>     SessContinue SessStatus = 1
>     SessEnd      SessStatus = 2
>     SessOnce     SessStatus = 3
> 
>     outputPerfFile   = "perf.txt"
>     outputRecordFile = "perfReqRecord.csv"
>     outputPerfImg    = "perf.jpg"
> )
> 
> /*
> xtest 性能检测工具
> */
> type callDetail struct {
>     ID       string     //uuid
>     Handle   string     //会话模式时的hdl
>     Tm       time.Time  //时间戳
>     dataStat DataStatus //数据状态 ，0,1,2,3
>     sessStat SessStatus //会话状态,0,1,2,3
>     Dire     direction  //输入 还是输出
>     ErrCode  int
>     ErrInfo  string
> }
> 
> type performance struct {
>     Max         float32 `json:"max"`
>     Min         float32 `json:"min"`
>     FailRate    float32 `json:"failRate"`
>     SuccessRate float32 `json:"successRate"`
>     //平均值95 99线
>     Delay95      float32 `json:"delay95"`
>     Delay99      float32 `json:"delay99"`
>     DelayAverage float32 `json:"delayAverage"`
>     //首结果95 99线
>     DelayFirstMin     float32 `json:"delayFirstMin"`
>     DelayFirstMax     float32 `json:"delayFirstMax"`
>     DelayFirst95      float32 `json:"delayFirst95"`
>     DelayFirst99      float32 `json:"delayFirst99"`
>     DelayFirstAverage float32 `json:"delayFirstAverage"`
>     //尾结果95 99线
>     DelayLastMin     float32 `json:"delayLatMin"`
>     DelayLastMax     float32 `json:"delayLatMax"`
>     DelayLast95      float32 `json:"delayLast95"`
>     DelayLast99      float32 `json:"delayLast99"`
>     DelayLastAverage float32 `json:"delayLastAverage"`
> }
> 
> type singlePerfCost struct {
>     id        string
>     cost      float32
>     firstCost float32 //首个结果耗时
>     lastCost  float32 //最后一个结果耗时
> }
> 
> type errMsg struct {
>     ErrInfo string `json:"errInfo"`
>     Handle  string `json:"handle"`
> }
> 
> type PerfModule struct {
>     idx            int
>     collectChan    chan callDetail
>     mtx            sync.Mutex
>     control        chan bool
>     correctReqPath map[string][]callDetail //正确的请求路径图
> 
>     errReqRecord map[int][]errMsg //错误的请求记录
> 
>     correctReqCost []singlePerfCost //正确的请求花费的时间记录
> 
>     perf performance //性能结果
> 
>     reqRecordFile *os.File
> 
>     Log *utils.Logger
> }
> 
> var Perf *PerfModule
> 
> // 初始化Performance实例，并启动一个collect协程收集性能日志
> func (pf *PerfModule) Start() (err error) 
> // 关闭请求记录文件，calcDelay计算请求的性能指标，dump将性能指标数据落盘并关闭collectChan收集管道
> func (pf *PerfModule) Stop() 
> // 将采集详细数据写入collectChan收集管道
> func (pf *PerfModule) Record(id, handle string, stat DataStatus, stat2 SessStatus, dire direction, errCode int, errInfo string) 
> // 读取collectChan收集管道，将信息分类为正确与错误信息并记录correctReqPath和errReqRecord
> func (pf *PerfModule) collect() 
> // 计算correctReqPath[id]请求响应的时间开销
> func (pf *PerfModule) pretreatment(id string) 
> // 从性能日志文件中解析数据到实例
> func (pf *PerfModule) loadRecord() error 
> // 计算最后才能知道的性能指标，例如正确率、失败率、95、99指标
> func (pf *PerfModule) calcDelay()
> // 写入性能指标到outputPerfFile日志文件
> func (pf *PerfModule) dump()
> // 从data数据中计算出性能指标
> func (pf *PerfModule) anallyArray(data []float32) (min, max, average, aver95, aver99 float32) 
> ```

### 2.3、 inclue文件夹

#### Ⅰ.h264_nalu_spilt.h

#### Ⅱ. type.h

### 2.4、 lib文件夹

#### Ⅰ. libh264bitstream.so.0

### 2.5、request文件夹

#### Ⅰ. fileSession.go

> ```go
> // 文件session请求
> func FileSessionCall(cli *xsfcli.Client, index int64) (code int, err error) 
> // 文件AI上行请求
> func FilesessAIIn(cli *xsfcli.Client, indexs int64, thrRslt *[]protocol.LoaderOutput, thrLock *sync.Mutex, reqSid string) (hdl string, status protocol.LoaderOutput_RespStatus, code int, err error) 
> // 多线程文件上传流请求
> func FilemultiUpStream(cli *xsfcli.Client, swg *sync.WaitGroup, session string, pm *[]protocol.LoaderOutput, sm *sync.Mutex, errchan chan struct {
> code int
> err  error
> }) 
> 
> // 实时性校准,用于校准发包大小及发包时间间隔之间的实时性.
> func FilertCalibration(curReq int, interval int, sTime time.Time)
> 
> // downStream 下行调用单线程;
> func FilesessAIOut(cli *xsfcli.Client, hdl string, sid string, rslt *[]protocol.LoaderOutput) (code int, err error) 
> // 文件session报错
> func FilesessAIExcp(cli *xsfcli.Client, hdl string, sid string) (err error)
> 
> // upStream first error ，将上传流错误写入ch管道
> func FileunBlockChanWrite(ch chan struct {
> code int
> err  error
> }, err struct {
> code int
> err  error
> }) 
> ```

#### Ⅱ. oneShot.go：与RPC通信有关的函数定义

> ```go
> // 使用xsf框架发起RPC通信，设置协议参数、上行数据键值对，使用ONESHORT方式发起SessionCall，然后
> // 下行数据解析到AsyncDrop下行数据异步落盘同步通道，如果通道满了，使用downOutput函数写入本地文件。
> func OneShotCall(cli *xsfcli.Client, index int64) (code int, err error)
> ```

#### Ⅲ. output.go：下行数据输出有关的函数定义

> ```go
> // 读取AsyncDrop通道中的下行数据，调用downOutput函数写入本地文件
> func DownStreamWrite(wg *sync.WaitGroup, log *utils.Logger) 
> // 写入数据到本地文件（_var.OutputDst和_var.Output配置相关文件路径）
> func downOutput(key string, data []byte, log *utils.Logger)
> ```

#### Ⅳ. session.go

> ```go
> // Session调用
> func SessionCall(cli *xsfcli.Client, index int64) (code int, err error) 
> 
> // AI调用输入
> func sessAIIn(cli *xsfcli.Client, indexs []int, thrRslt *[]protocol.EngOutputData, thrLock *sync.Mutex, reqSid string) (hdl string, status protocol.EngOutputData_DataStatus, code int, err error)
> 
> // 多线程上传流
> func multiUpStream(cli *xsfcli.Client, swg *sync.WaitGroup, session string, interval int, indexs map[int]int, sid string, pm *[]protocol.EngOutputData, sm *sync.Mutex, errchan chan struct {
> 	code int
> 	err  error
> })
> // 实时性校准,用于校准发包大小及发包时间间隔之间的实时性.
> func rtCalibration(curReq int, interval int, sTime time.Time)
> // downStream 下行调用单线程;
> func sessAIOut(cli *xsfcli.Client, hdl string, sid string, rslt *[]protocol.EngOutputData) (code int, err error) 
> // 
> func sessAIExcp(cli *xsfcli.Client, hdl string, sid string) (err error)
> // upStream first error 将错误信息发送至ch管道
> func unBlockChanWrite(ch chan struct {
> 	code int
> 	err  error
> }, err struct {
> 	code int
> 	err  error
> })
> ```

#### Ⅴ. signal.go：xtest退出有关的函数定义

> ```go
> // 通过signal.Notify转发信号，优雅退出程序
> func SigRegister() 
> ```

#### Ⅵ. splitFrame.go

> ```go
> func GetH264Frames(video []byte) (frameSizes []int)
> ```

#### Ⅶ. textLine.go

> ```go
> // 文本session请求
> func TextCall(cli *xsfcli.Client, index int64) (code int, err error)
> // 文本AI上行
> func TextAIIn(cli *xsfcli.Client, indexs int64, thrRslt *[]protocol.LoaderOutput, thrLock *sync.Mutex, reqSid string) (hdl string, status protocol.LoaderOutput_RespStatus, code int, err error)
> 
> // 多线程文本上行数据流
> func TextmultiUpStream(cli *xsfcli.Client, swg *sync.WaitGroup, session string, pm *[]protocol.LoaderOutput, sm *sync.Mutex, errchan chan struct {
> code int
> err  error
> }) 
> 
> // 实时性校准,用于校准发包大小及发包时间间隔之间的实时性.
> func TextrtCalibration(curReq int, interval int, sTime time.Time)
> 
> // downStream 下行调用单线程;
> func TextsessAIOut(cli *xsfcli.Client, hdl string, sid string, rslt *[]protocol.LoaderOutput) (code int, err error) 
> 
> func TextsessAIExcp(cli *xsfcli.Client, hdl string, sid string) (err error)
> 
> // upStream first error，错误日志写入管道
> func TextunBlockChanWrite(ch chan struct {
> code int
> err  error
> }, err struct {
> code int
> err  error
> }) 
> ```

### 2.6、script文件夹

#### Ⅰ. test.sh： 运行脚本

> 启动xtest脚本

#### Ⅱ. xtest.toml：配置文件

> ```toml
> [xtest]
> #测试目标服务配置，配置格式如下,注意分割符的差异. 业务1@ip1:port1;ipn:portn,业务2@ip2:port2;ipn:portn
> taddrs="AIservice@127.0.0.1:5090"
> trace-ip = "172.16.51.13"
> 
> [svcMode]
> service = "AIservice"           # 请求目标服务名称, eg:sms
> timeout = 1000                  # 服务超时时间, 对应服务端waitTime
> multiThr = 10                  # 请求并发数
> loopCnt = 100                 # 请求总次数
> reqMode = 0                     # 服务请求模式. 0:非会话模式 1:会话模式
> reqPara = "k1=v1,k2=v2,k3=v3"   # 服务请求参数对, 多个参数对以","分隔
> linearMs = 5000                 # 并发压测线性增长区间,单位:ms
> 
> 
> [upstream]
> inputSrc = "./test"               # 上行数据流数据源, 配置文件路径(配置为目录则循环读取目录中文件)
> sliceOn = 1                     # 切片开关, 0:关闭切换, !0:开启切片
> sliceSize = 1280                # 上行数据切片大小,用于会话模式: byte
> interval = 40                   # 上行数据发包间隔,用于会话模式: ms. 注：动态校准,每个包间隔非严格interval
> type = "audio"                  # 数据类型
> format = "audio/L16;rate=16000" # 数据格式
> encoding = "raw"                # 数据编码
> describe = "k=v;k=v"            # 数据描述信息
> 
> 
> #[upstream-N]                    # 用于实现多数据流上行配置, 对于多个数据流可按照upstream-N规则叠加配置
> #inputSrc = "path-N"             # 同upstream
> #sliceSize = 1280                # 同upstream
> #interval = 40                   # 同upstream, 注:多个upstream发包间隔相同时,数据流发送包合并
> #type = "audio"                  # 同upstream
> #format = "audio/L16;rate=16000" # 同upstream
> #encoding = "raw"                # 同upstream
> #describe = "k=v;k=v"            # 同upstream
> 
> 
> [downstream]                    # 下行数据流存储输出
> output = 0                      # 输出方式： 0:输出至公共文件outputDst 1:以独立文件形式输出至目录outputDst(文件名:sid+**) 2：输出至终端
> outputDst = "./log/result"          # 响应数据输出目标, 若output=0则配置目录, output=1则配置文件
> 
> 
> [log]
> file = "./log/xtest.log"              # 日志文件名
> level = "debug"                 # 日志打印等级
> size = 100                      # 日志文件大小
> count = 20                      # 日志备份数量
> async = 0                       # 异步开关
> die = 30
> 
> [trace]
> able = 0
> ```

#### Ⅲ. xtest_example.toml文件

### 2.7、testdata文件夹

### 2.8、util文件夹

#### Ⅰ sid.go：与sid生成有关的函数定义

> ```go
> var (
>     index        int64  = 0		// 生成的SID索引
>     Location     string = "dx"
>     LocalIP      string 			// 本地IP
>     ShortLocalIP string			// 本地短IP
>     Port         string			// 端口
> )
> // 获取本地ip地址与短地址ip
> func init() 
> // 生成sid
> func NewSid(sub string) string 
> ```

### 2.9、var文件夹

#### Ⅰ. cmd.go：命令行输入相关

> ```go
> var (
> // default 缺省配置模式为native
>     CmdCfg = flag.String("f", "xtest.toml", "client cfg name") // 配置文件选项
>     XTestVersion = flag.String("v", "2.5.2", "xtest version") // Xtest版本号
> )
> // 打印xtest用法配置选项
> func Usage() 
> ```

#### Ⅱ. conf.go：xtest配置相关定义和函数

> ```go
> package _var
> 
> import (
> 	"errors"
> 	"fmt"
> 	"git.iflytek.com/AIaaS/xsf/utils"
> 	"go.uber.org/atomic"
> 	"io/ioutil"
> 	"os"
> 	"protocol"
> 	"reflect"
> 	"strconv"
> 	"strings"
> )
> 
> const (
> 	CliName = "xtest"
> 	CliVer  = "2.0.1"
> )
> 
> type InputMeta struct {
> 	Name       string                     // 上行数据流key值
> 	DataSrc    string                     // 上行实体数据来源;数据集则配置对应目录
> 	SliceOn    int                        // 上行数据切片开关, !0:切片. 0:不切片
> 	UpSlice    int                        // 上行数据切片大小: byte
> 	UpInterval int                        // slice发包间隔: ms
> 	DataType   protocol.MetaDesc_DataType // audio/text/image/video
> 	DataDesc   map[string]string
> 
> 	// DataList map[string/*file*/] []byte /*data*/
> 	DataList [][]byte /*data*/
> }
> 
> type OutputMeta struct {
> 	Name string            // 下行数据流key
> 	Sid  string            // sid
> 	Type string            // 下行数据类型
> 	Desc map[string]string // 数据描述
> 	Data []byte            // 下行数据实体
> }
> 
> var (
> 	// [svcMode]
> 	SvcId         string        = "s12345678"
> 	SvcName       string        = "AIservice"            // dst service name
> 	TimeOut       int           = 1000                   // 超时时间: ms, 对应加载器waitTime
> 	LossDeviation int           = 50                     // 自身性能损耗误差, ms.
> 	MultiThr      int           = 100                    // 请求并发数
> 	DropThr                     = 100                    // 下行数据异步输出线程数
> 	LoopCnt       *atomic.Int64 = atomic.NewInt64(10000) // 请求总次数
> 	ReqMode       int           = 0                      // 0: 非会话模式, 1: 常规会话模式 2.文本按行会话模式 3.文件会话模式
> 	LinearNs      int           = 0                      // 并发模型线性增长时间,用于计算并发增长斜率(单位：ns). default:0,瞬时并发压测.
> 	TestSub       string        = "ase"                  // 测试业务sub, 缺省test
> 
> 	PerfConfigOn bool = false //true: 开启性能检测 false: 不开启性能检测
> 	PerfLevel    int  = 0     //非会话模式默认0
> 	//会话模式0: 从发第一帧到最后一帧的性能
> 	//会话模式1:首结果(发送第一帧到最后一帧的性能)
> 	//会话模式2:尾结果(发送最后一帧到收到最后一帧的性能)
> 	// 请求参数对
> 	Header map[string]string = make(map[string]string)
> 	Params map[string]string = make(map[string]string)
> 
> 	Payload []string // 上行数据流
> 	Expect  []string // 下行数据流
> 
> 	// 上行数据流配置, 多数据流通过section [data]中payload进行配置
> 	UpStreams []InputMeta = make([]InputMeta, 0, 1)
> 
> 	DownExpect []protocol.MetaDesc
> 
> 	// [downstream]
> 	Output = 0 // 0：输出至公共文件outputDst(sid+***:data)
> 	// 1：以独立文件形式输出至目录outputDst(文件名:sid+***)
> 	// 2：输出至终端
> 	//-1：不输出
> 	OutputDst = "./log/result" // output=0时,该项配置输出文件名; output=1时,该项配置输出目录名
> 	ErrAnaDst = "./log/errDist"
> 	AsyncDrop chan OutputMeta // 下行数据异步落盘同步通道
> )
> 
> // 参数初始化
> func ConfInit(conf *utils.Configure) error
> // 解析下行数据字段
> func secParseEp(conf *utils.Configure) error 
> //解析上行数据字段
> func secParsePl(conf *utils.Configure) error
> // 解析RPC服务字段
> func secParseSvc(conf *utils.Configure) error 
> // 解析请求头字段
> func secParseHeader(conf *utils.Configure) error
> // 解析请求参数字段
> func secParseParams(conf *utils.Configure) error
> // 解析请求数据字段
> func secParseData(conf *utils.Configure) error
> // 解析下行数据流字段
> func secParseDStream(conf *utils.Configure) error
> 
> ```

## 三、需求分析

### 3.1 功能需求

- 模式-支持流式、非流式、异步回调三种模式
- 非流式模式下
    - [x] 读取文件输入
    - [x] 配置中心手动输入数据 √ 使用scan输入文本数据   success
- 流式模式下
    - [x] 单一文件一次输入
    - [x] 单一文件按照固定长度输入
    - [ ] 文本文件按行读取 √  优化代码
    - [ ] 多个文件，每个文件一帧输入
        - [ ] 文件有序 √ 优化代码
        - [ ] 文件无序：map √ 实现功能

### 3.2 性能需求

- [x] 并发，显示当前路数 √ 实现功能 success
- [ ] 成功率，性能数据及性能分布，输出本地相关数据 √ 优化代码
- [x] 内存、显存定时统计： 内存：syscall 包调用。显存：cmd执行 nvidia-smi ？node_exporter？

### 3.3 其他需求

- 文档说明

- demo样例

  