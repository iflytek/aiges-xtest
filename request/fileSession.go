package request

import (
	"errors"
	"frame"
	xsfcli "git.iflytek.com/AIaaS/xsf/client"
	"github.com/golang/protobuf/proto"
	"protocol"
	"strconv"
	"sync"
	"time"
	"xtest/analy"
	"xtest/util"
	_var "xtest/var"
)

func FileSessionCall(cli *xsfcli.Client, index int64) (info analy.ErrInfo) {
	// 下行结果缓存

	// go routine 区分不同frame slice数据流
	var thrRslt []protocol.LoaderOutput = make([]protocol.LoaderOutput, 0, 1)
	var thrLock sync.Mutex
	reqSid := util.NewSid(_var.TestSub)
	_var.ConcurrencyCnt.Add(1) // jbzhou5 启动协程时+1
	hdl, status, info := FilesessAIIn(cli, index, &thrRslt, &thrLock, reqSid)
	if info.ErrStr != nil {
		if len(hdl) != 0 {
			_ = FilesessAIExcp(cli, hdl, reqSid)
			return
		}
	} else if status != protocol.LoaderOutput_END {
		info = FilesessAIOut(cli, hdl, reqSid, &thrRslt)
		if info.ErrStr != nil {
			_ = FilesessAIExcp(cli, hdl, reqSid)
			return info
		}
	}
	_ = FilesessAIExcp(cli, hdl, reqSid)
	_var.ConcurrencyCnt.Dec() // jbzhou5 任务完成时-1
	// 结果落盘
	tmpMerge := make(map[string] /*streamId*/ *protocol.Payload)
	for k, _ := range thrRslt {
		for _, d := range thrRslt[k].Pl {
			meta, exist := tmpMerge[d.Meta.Name]
			if exist {
				tmpMerge[d.Meta.Name].Data = append(meta.Data, d.Data...)
			} else {
				tmpMerge[d.Meta.Name] = d
			}
		}
	}

	for _, v := range tmpMerge {
		var outType string = "invalidType"
		switch v.Meta.DataType {
		case protocol.MetaDesc_TEXT:
			outType = "text"
		case protocol.MetaDesc_AUDIO:
			outType = "audio"
		case protocol.MetaDesc_IMAGE:
			outType = "image"
		case protocol.MetaDesc_VIDEO:
			outType = "video"
		}

		select {
		case _var.AsyncDrop <- _var.OutputMeta{reqSid, outType, v.Meta.Name, v.Meta.Attribute, v.Data}:
		default:
			// 异步channel满, 同步写;	key: sid-type-format-encoding, value: data
			key := reqSid + "-" + outType + "-" + v.Meta.Name
			downOutput(key, v.Data, cli.Log)
		}
	}
	return
}

func FilesessAIIn(cli *xsfcli.Client, indexs int64, thrRslt *[]protocol.LoaderOutput, thrLock *sync.Mutex, reqSid string) (hdl string, status protocol.LoaderOutput_RespStatus, info analy.ErrInfo) {
	// request构包；构造首包SeqNo=1,同加载器建立会话上下文信息; 故首帧不携带具体数据
	req := xsfcli.NewReq()
	req.SetParam("SeqNo", "1")
	req.SetParam("baseId", "0")
	req.SetParam("version", "v2")
	req.SetParam("waitTime", strconv.Itoa(_var.TimeOut))
	dataIn := protocol.LoaderInput{}
	dataIn.State = protocol.LoaderInput_STREAM
	dataIn.ServiceId = _var.SvcId
	dataIn.ServiceName = _var.SvcName
	// 平台参数header
	dataIn.Headers = make(map[string]string)
	dataIn.Headers["sid"] = reqSid
	dataIn.Headers["status"] = "0"
	for k, v := range _var.Header {
		dataIn.Headers[k] = v
	}
	// 能力参数params
	dataIn.Params = make(map[string]string)
	for k, v := range _var.Params {
		dataIn.Params[k] = v
	}
	// 期望输出expect
	for k, _ := range _var.DownExpect {
		dataIn.Expect = append(dataIn.Expect, &_var.DownExpect[k])
	}

	input, err := proto.Marshal(&dataIn)
	if err != nil {
		cli.Log.Errorw("sessAIIn marshal create request fail", "err", err.Error(), "params", dataIn.Params)
		return hdl, status, analy.ErrInfo{ErrCode: -1, ErrStr: err}
	}

	rd := xsfcli.NewData()
	rd.Append(input)
	req.AppendData(rd)

	caller := xsfcli.NewCaller(cli)
	analy.Perf.Record(reqSid, "", analy.DataBegin, analy.SessBegin, analy.UP, 0, "")
	resp, ecode, err := caller.SessionCall(xsfcli.CREATE, _var.SvcName, "AIIn", req, time.Duration(_var.TimeOut+_var.LossDeviation)*time.Millisecond)
	if err != nil {
		cli.Log.Errorw("sessAIIn Create request fail", "err", err.Error(), "code", ecode, "params", dataIn.Params)
		analy.Perf.Record(reqSid, resp.Handle(), analy.DataBegin, analy.SessBegin, analy.DOWN, int(ecode), err.Error())
		return hdl, status, analy.ErrInfo{ErrCode: int(ecode), ErrStr: err}
	}
	hdl = resp.Session()
	analy.Perf.Record(reqSid, resp.Handle(), analy.DataBegin, analy.SessBegin, analy.DOWN, 0, "")

	//主体数据发送
	// data stream: 相同UpInterval合并发送;

	errChan := make(chan analy.ErrInfo, 1) // 仅保存首个错误码;
	defer close(errChan)
	var rwg sync.WaitGroup

	FilemultiUpStream(cli, &rwg, hdl, thrRslt, thrLock, errChan)

	rwg.Wait() // 异步协程上行数据交互结束
	select {
	case einfo := <-errChan:
		return hdl, status, einfo
	default:
		// unblock; check status
		for k, _ := range *thrRslt {
			if (*thrRslt)[k].Status == protocol.LoaderOutput_END {
				status = (*thrRslt)[k].Status
			}
		}
	}
	return
}

func FilemultiUpStream(cli *xsfcli.Client, swg *sync.WaitGroup, session string, pm *[]protocol.LoaderOutput, sm *sync.Mutex, errchan chan analy.ErrInfo) {

	for dataId := 1; dataId <= len(_var.UpStreams[0].DataList); dataId++ {

		println("send data ")

		sendData := _var.UpStreams[0].DataList[dataId-1]
		sTime := time.Now()

		req := xsfcli.NewReq()
		req.SetParam("baseId", "0")
		req.SetParam("version", "v2")
		req.SetParam("waitTime", strconv.Itoa(_var.TimeOut))
		_ = req.Session(session)
		dataIn := protocol.LoaderInput{}
		dataIn.SyncId = int32(dataId)
		upStatus := protocol.EngInputData_CONTINUE
		if dataId == len(_var.UpStreams[0].DataList) {
			upStatus = protocol.EngInputData_END
		}
		desc := make(map[string]string)
		for dk, dv := range _var.UpStreams[0].DataDesc {
			desc[dk] = dv
		}
		md := protocol.MetaDesc{
			Name:      _var.UpStreams[0].Name,
			DataType:  _var.UpStreams[0].DataType,
			Attribute: desc}
		md.Attribute["seq"] = strconv.Itoa(dataId)
		md.Attribute["status"] = strconv.Itoa(int(upStatus))
		inputmeta := protocol.Payload{Meta: &md, Data: sendData}
		dataIn.Pl = append(dataIn.Pl, &inputmeta)
		input, err := proto.Marshal(&dataIn)
		if err != nil {
			cli.Log.Errorw("multiUpStream marshal create request fail", "err", err.Error(), "params", dataIn.Params)
			FileunBlockChanWrite(errchan, analy.ErrInfo{ErrCode: -1, ErrStr: err})
			return
		}

		rd := xsfcli.NewData()
		rd.Append(input)
		req.AppendData(rd)
		caller := xsfcli.NewCaller(cli)

		analy.Perf.Record("", req.Handle(), analy.DataContinue, analy.SessContinue, analy.UP, 0, "")

		resp, ecode, err := caller.SessionCall(xsfcli.CONTINUE, _var.SvcName, "AIIn", req, time.Duration(_var.TimeOut+_var.LossDeviation)*time.Millisecond)
		if err != nil && ecode != frame.AigesErrorEngInactive {
			cli.Log.Errorw("multiUpStream Create request fail", "err", err.Error(), "code", ecode, "params", dataIn.Params)
			FileunBlockChanWrite(errchan, analy.ErrInfo{ErrCode: int(ecode), ErrStr: err})
			analy.Perf.Record("", req.Handle(), analy.DataContinue, analy.SessContinue, analy.DOWN, int(ecode), err.Error())
			return
		}
		// 下行结果输出
		dataOut := protocol.LoaderOutput{}
		err = proto.Unmarshal(resp.GetData()[0].Data, &dataOut)
		if err != nil {
			cli.Log.Errorw("multiUpStream Resp Unmarshal fail", "err", err.Error(), "respData", resp.GetData()[0].Data)
			FileunBlockChanWrite(errchan, analy.ErrInfo{ErrCode: -1, ErrStr: err})
			return
		}

		switch dataOut.Code {
		case 0: // nothing to do
		case frame.AigesErrorEngInactive:
			return
		default:
			cli.Log.Errorw("multiUpStream get engine err", "err", dataOut.Err, "code", dataOut.Code, "params", dataIn.Params)
			FileunBlockChanWrite(errchan, analy.ErrInfo{ErrCode: int(dataOut.Code), ErrStr: errors.New(dataOut.Err)})
			analy.Perf.Record("", req.Handle(), analy.DataContinue, analy.SessContinue, analy.DOWN, int(dataOut.Code), dataOut.Err)
			return // engine err but not 10101
		}

		// 同步下行数据
		if len(dataOut.Pl) > 0 {
			(*sm).Lock()
			*pm = append(*pm, dataOut)
			cli.Log.Debugw("multiUpStream get resp result", "hdl", session, "result", dataOut)
			(*sm).Unlock()
			analy.Perf.Record("", req.Handle(), analy.DataStatus(int(dataOut.Status)), analy.SessContinue, analy.DOWN, 0, "")
		}
		if dataOut.Status == protocol.LoaderOutput_END {
			return // last result
		}

		// wait ms.动态调整校准上行数据实时率, 考虑其他接口耗时.
		FilertCalibration(dataId, _var.UpStreams[0].UpInterval, sTime)
	}

}

// 实时性校准,用于校准发包大小及发包时间间隔之间的实时性.
func FilertCalibration(curReq int, interval int, sTime time.Time) {
	cTime := int(time.Now().Sub(sTime).Nanoseconds() / (1000 * 1000)) // ssb至今绝对时长.ms
	expect := interval * (curReq + 1)                                 // 期望发包时间
	if expect > cTime {
		time.Sleep(time.Millisecond * time.Duration(expect-cTime))
	}
}

// downStream 下行调用单线程;
func FilesessAIOut(cli *xsfcli.Client, hdl string, sid string, rslt *[]protocol.LoaderOutput) (info analy.ErrInfo) {
	// loop read downstream result
	for {
		req := xsfcli.NewReq()
		req.SetParam("baseId", "0")
		req.SetParam("version", "v2")
		req.SetParam("waitTime", strconv.Itoa(_var.TimeOut))
		_ = req.Session(sid)
		dataIn := protocol.LoaderInput{}

		input, err := proto.Marshal(&dataIn)
		if err != nil {
			cli.Log.Errorw("sessAIOut marshal create request fail", "err", err.Error(), "params", dataIn.Params)
			return analy.ErrInfo{ErrCode: -1, ErrStr: err}
		}

		rd := xsfcli.NewData()
		rd.Append(input)
		req.AppendData(rd)
		_ = req.Session(hdl)

		caller := xsfcli.NewCaller(cli)
		resp, ecode, err := caller.SessionCall(xsfcli.CONTINUE, _var.SvcName, "AIOut", req, time.Duration(_var.TimeOut+_var.LossDeviation)*time.Millisecond)
		if err != nil {
			cli.Log.Errorw("sessAIOut request fail", "err", err.Error(), "code", ecode, "params", dataIn.Params)
			if ecode == frame.AigesErrorEngInactive { // reset 10101 inactive
				err = nil
			}
			analy.Perf.Record("", req.Handle(), analy.DataContinue, analy.SessContinue, analy.DOWN, int(ecode), err.Error())

			return analy.ErrInfo{ErrCode: int(ecode), ErrStr: err}
		}

		// 解析结果、输出落盘
		dataOut := protocol.LoaderOutput{}
		err = proto.Unmarshal(resp.GetData()[0].Data, &dataOut)
		if err != nil {
			cli.Log.Errorw("sessAIOut Resp Unmarshal fail", "err", err.Error(), "respData", resp.GetData()[0].Data)
			return analy.ErrInfo{ErrCode: -1, ErrStr: err}
		}

		*rslt = append(*rslt, dataOut)
		analy.Perf.Record("", req.Handle(), analy.DataStatus(int(dataOut.Status)), analy.SessContinue, analy.DOWN, int(dataOut.Code), dataOut.Err)
		cli.Log.Debugw("sessAIOut get resp result", "hdl", sid, "result", dataOut)
		if dataOut.Status == protocol.LoaderOutput_END {
			return // last result
		}
	}

	return
}

func FilesessAIExcp(cli *xsfcli.Client, hdl string, sid string) (err error) {

	req := xsfcli.NewReq()
	req.SetParam("baseId", "0")
	req.SetParam("waitTime", strconv.Itoa(_var.TimeOut))
	dataIn := protocol.LoaderInput{}
	input, err := proto.Marshal(&dataIn)
	if err != nil {
		cli.Log.Errorw("sessAIExcp marshal create request fail", "err", err.Error(), "params", dataIn.Params)
		return
	}

	rd := xsfcli.NewData()
	rd.Append(input)
	req.AppendData(rd)
	_ = req.Session(hdl)

	caller := xsfcli.NewCaller(cli)
	_, _, err = caller.SessionCall(xsfcli.CONTINUE, _var.SvcName, "AIExcp", req, time.Duration(_var.TimeOut+_var.LossDeviation)*time.Millisecond)
	return
}

// upStream first error
func FileunBlockChanWrite(ch chan analy.ErrInfo, err analy.ErrInfo) {
	select {
	case ch <- err:
	default:
		// ch full, return. save first err code
	}
}
