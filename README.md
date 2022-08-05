# Xtest测试工具
> xtest 是一个基于Aiges平台的自动化测试工具，可支持单元测试、压力测试、功能测试等测试功能，并提供了基本的资源监控，可配合prometheus和grafana使用。

## 一、使用说明
1. 启动Aiservice，注意监听的端口是否有改变。
2. 根据自己的AI模型修改xtest.toml文件，例如Aiservice端口、参数，测试轮数等配置，具体请参考说明六。
3. 由于监听资源需要管理员权限，请保证本地已经安装**netstat**命令，然后执行： ```sudo ./xtest``` 或```sudo ./xtest -f *.toml ```命令启动， 否则资源文件将为空，但并不影响其他任务。
4. GPU 监控为待开放功能，源码已实现，但要求用户拥有英伟达显卡监控（Nvidia-smi）


## 二、需求分析

### 2.1 功能需求

- 模式-支持流式、非流式、异步回调三种模式
- 非流式模式下
    - [x] 读取文件输入
    - [x] 配置中心手动输入数据
- 流式模式下
    - [x] 单一文件一次输入
    - [x] 单一文件按照固定长度输入
    - [x] 文本文件按行读取 √  优化代码
    - [x] 多个文件，每个文件一帧输入
        - [x] 文件有序
        - [x] 文件无序

### 2.2 性能需求

- [x] 并发，显示当前路数
- [x] 成功率，性能数据及性能分布，输出本地相关数据
- [x] 内存、显存定时统计
  - [x]CPU
  - [x]内存
  - [x]GPU

### 2.3 其他需求

- 文档说明
- demo样例


## 三、配置说明

### 3.1 样例配置

```toml
[xtest]
taddrs="AIservice@127.0.0.1:5090"
trace-ip = "172.16.51.13"

[svcMode]
service = "AIservice"           # 请求目标服务名称, eg:sms
svcId = "s12345678"             # 服务id
timeout = 1000                  # 服务超时时间, 对应服务端waitTime
multiThr = 100                  # 请求并发数
loopCnt = 100000                # 请求总次数
sessMode = 0                    # 0: 非会话模式, 1: 常规会话模式 2.文本按行会话模式 3.文件会话模式
linearMs = 5000                 # 并发压测线性增长区间,单位:ms
perfOn=true                     # 是否开启性能测试
perfLevel=0                     # 非会话模式默认0
                                # 会话模式0: 从发第一帧到最后一帧的性能
                                # 会话模式1:首结果(发送第一帧到最后一帧的性能)
                                # 会话模式2:尾结果(发送最后一帧到收到最后一帧的性能)
inputCmd = false 				# 切换为命令行输入，仅在非会话模式生效
prometheus_switch = true  		# Prometheus开关， 开启后开启双写，同时写入prometheus与本地日志
prometheus_port = 2117    # jbzhou5 Prometheus指标暴露端口
plot = true  					# 绘制资源图， 默认开启
plot_file = "./log/line.png"    # 绘制图形保存路径
file_sorted = 0  				# 传入文件是否排序， 0： 随机， 1： 升序， 2： 降序
file_name_seq = "_" 			# 传入文件名分割方式 例如传入'_', 则1_2.txt -> 1，2_2.txt -> 2, 为空或者传入非法则不处理
[header]
"appid" = "100IME"
"uid" = "1234567890"

[parameter]
"key" = 2
"x" = 1

[data]
payload = "dataKey2"   			# 输入数据流配置段,多个数据流以";"分割， 如果开启了inputCmd， 该值会被清空
expect = "dataKey3"             # 输出数据流配置段,多个数据流以";"分割

[dataKey1]  # 输入数据流dataKey1描述
inputSrc = "path"               # 上行数据流数据源, 配置文件路径(配置为目录则循环读取目录中文件)
sliceOn = false                 # 切片开关, false:关闭切换, true:开启切片
sliceSize = 1280                # 上行数据切片大小,用于会话模式: byte
interval = 40                   # 上行数据发包间隔,用于会话模式: ms. 注：动态校准,每个包间隔非严格interval
name = "input1"                 # 输入数据流key值
type = "image"                  # 数据类型，支持"audio","text","image","video"
describe = "k1=v1;k2=v2"        # 数据描述信息,多个描述信息以";"分割
                                # 图像支持如下属性："encoding", 如"encoding=jpg"
[dataKey2]
inputSrc = "./testdata/text2"   # 上行数据流数据源, 配置文件路径(配置为目录则循环读取目录中文件)
sliceOn = false                 # 切片开关, false:关闭切换, true:开启切片
sliceSize = 1280                # 上行数据切片大小,用于会话模式: byte
interval = 40                   # 上行数据发包间隔,用于会话模式: ms. 注：动态校准,每个包间隔非严格interval
name = "input2"                 # 输入数据流key值
type = "text"                  	# 数据类型，支持"audio","text","image","video"
describe = "encoding=utf8"      # 数据描述信息,多个描述信息以";"分割
                                # 图像支持如下属性："encoding", 如"encoding=jpg"
[dataKey3]
name = "result"                 # 输入数据流key值
type = "text"                   # 输出数据类型，支持"audio","text","image","video"
describe = "k1=v1;k2=v2"        # 数据描述信息,多个描述信息以";"分割
                                # 文本支持如下属性："encoding","compress", 如"encoding=utf8;compress=gzip"


[downstream]                    # 下行数据流存储输出
output = 0                      # 输出方式： 0:输出至公共文件outputDst 1:以独立文件形式输出至目录outputDst(文件名:sid+**) 2：输出至终端
outputDst = "./log/result"      # 响应数据输出目标, 若output=0则配置目录, output=1则配置文件


[log]
file = "./log/xtest.log"        # 日志文件名
level = "debug"                 # 日志打印等级
size = 100                      # 日志文件大小
count = 20                      # 日志备份数量
async = 0                       # 异步开关
die = 30

[trace]
able = 0
```
## 四、字段说明
> **xtest.toml 中大部分字段一般保持不变，下面仅对常用字段进行说明解释。**
- ```[xtest]```
  - ```taddrs="AIservice@127.0.0.1:5090"```： 与Aiservice的通信地址，与AIservice的启动端口对应，其中端口会被解析用于获取Aiservice的进程，监听其使用资源信息。
  - ``` trace-ip = "172.16.51.13```
- ```[svcMode]```
  - ```service = "AIservice"``` ： 请求目标服务名称, eg:sms
  - ```svcId = "s12345678"```   ：服务id
  - ```timeout = 1000``` ：服务超时时间, 对应服务端waitTime
  - ```multiThr = 100``` ：请求并发数，即同时开启多个协程发送请求测试
  - ```loopCnt = 100000``` ： multiThr个协程发送的请求总次数
  - ```sessMode = 0``` ： 0: 非会话模式, 1: 常规会话模式 2.文本按行会话模式 3.文件会话模式
  - ```linearMs = 5000``` ：并发压测线性增长区间,单位:ms
  - ```perfOn=true``` ： 是否开启性能测试，即是否在log文件夹底下记录perf.txt和PerfRecord.csv，用于记录成功率、失败率、发送延迟等性能指标。
  - ```perfLevel=0 ```：与sessMode字段对应，非会话模式默认0，会话模式0: 从发第一帧到最后一帧的性能，会话模式1:首结果(发送第一帧到最后一帧的性能)，会话模式2:尾结果(发送最后一帧到收到最后一帧的性能)
  - ```inputCmd = false ```：切换为命令行输入，仅在非会话模式生效，配置该字段时，所配置的[data] 字段将失效，仅读取用户命令行输入数据
  - ```prometheus_switch = true``` ：Prometheus开关，开启后会开放一个Prometheus监控端口，可使用grafana进行数据的展示。关闭/打开都会在Log目录生成一个Resource.csv 资源监听文件。
  - ```prometheus_port = 2117```：Prometheus指标暴露端口
  - ```plot = true``` ：绘制资源图，默认开启，将绘制Aiservice使用的资源变化折线图
  - ```plot_file = "./log/line.png"``` ：绘制图形保存路径
  - ```file_sorted = 0``` ：[data] 传入文件是否按名称排序， 0： 随机， 1： 升序， 2： 降序
  - ```file_name_seq = "_"``` ： 传入文件名分割方式 例如传入'_', 则1_2.txt -> 1，2_2.txt -> 2, 为空或者传入非法（即不能作为文件名的字符）则不处理， 注意此处分割为仅保留前半部分，若文件名为1_2_3.txt， 则得到的分割文件名为1。
- 
- ```[parameter]```：使用的AI模型需要传入的字段，根据自己需要填写
  - ```"key" = 2```
  - ```"x" = 1```

- ```[data]```
  - ```payload = "dataKey2"``` ：输入数据流配置段,多个数据流以";"分割， jbzhou5 如果开启了inputCmd， 该值会被清空
  - ```expect = "dataKey3"```：输出数据流配置段,多个数据流以";"分割

- ```[dataKey1] ``` ：名称可自定义，主要用于在[data] 字段的payload属性中方便标记加载数据，输入数据流dataKey1描述
  - ```inputSrc = "path" ```：上行数据流数据源, 配置文件路径(配置为目录则循环读取目录中文件)
  - ```sliceOn = false```：切片开关, false:关闭切换, true:开启切片
  - ```sliceSize = 1280```：上行数据切片大小,用于会话模式: byte
  - ```interval = 40```： 上行数据发包间隔,用于会话模式: ms. 注：动态校准,每个包间隔非严格interval
  - ```name = "input1"```： 输入数据流key值
  - ```type = "image"```： 数据类型，支持"audio","text","image","video"
  - ```describe = "k1=v1;k2=v2"```： 数据描述信息,多个描述信息以";"分割，图像支持如下属性："encoding", 如"encoding=jpg"

- ```[dataKey2]```
  - ```inputSrc = "./testdata/text2"```：上行数据流数据源, 配置文件路径(配置为目录则循环读取目录中文件)
  - ```sliceOn = false```： 切片开关, false:关闭切换, true:开启切片
  - ```sliceSize = 1280```：上行数据切片大小,用于会话模式: byte
  - ```interval = 40```： 上行数据发包间隔,用于会话模式: ms. 注：动态校准,每个包间隔非严格interval
  - ```name = "input2"```： 输入数据流key值
  - ```type = "text"``` ： 数据类型，支持"audio","text","image","video"
  - ```describe = "encoding=utf8"``` ： 数据描述信息,多个描述信息以";"分割图像支持如下属性："encoding", 如"encoding=jpg"

- ``` [dataKey3]```
  - ```name = "result"```： 输入数据流key值
  - ```type = "text"```：输出数据类型，支持"audio","text","image","video"
  - ```describe = "k1=v1;k2=v2"```： 数据描述信息,多个描述信息以";"分割，文本支持如下属性："encoding","compress", 如"encoding=utf8;compress=gzip"


- ```[downstream] ``` ：下行数据流存储输出
  - ```output = 0 ```： 输出方式： 0:输出至公共文件outputDst 1:以独立文件形式输出至目录outputDst(文件名:sid+**) 2：输出至终端
  - ```outputDst = "./log/result"```：响应数据输出目标, 若output=0则配置目录, output=1则配置文件


- ```[log]```
  - ```file = "./log/xtest.log"```：日志文件名
  - ```level = "debug" ```：日志打印等级
  - ```size = 100 ```：日志文件大小，单个超过size，将会写入新文件。
  - ```count = 20``` ： 日志备份数量
  - ```async = 0```： 异步开关
  - ```die = 30```

- ```[trace]```
  - ```able = 0```

五、[代码说明](https://github.com/tupig-7/aiges-xtest/tree/main/doc)