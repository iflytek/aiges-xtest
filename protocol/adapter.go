package protocol

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"strconv"
	"strings"
	"xtest/frame"
)

const (
	// pb协议版本, v1:AIaaS v2:AIPaaS
	ReqV1  = "v1"
	ReqV2  = "v2"
	ReqSrc = "source"
	SrcV1  = "aiaas"
	SrcV2  = "aipaas"

	// v1协议能力标记(serviceId)
	sampleRateV1  = "rate"
	frameSizeV1   = "spx_fsize"
	audioEncoding = "aue"
	serviceIat    = "ent"
	serviceTts    = "vcn"
	ttsDefaultAue = "speex-wb"
	ttsParamAuf   = "auf"
	ttsParamSfl   = "sfl"
	ttsParamGao   = "gao"
	ttsParamTte   = "tte"
	ttsParamOrg   = "speex-org"
)

type compressibility struct {
	original string
	encFrame string
}

var SpeexEncRate map[string]*compressibility = map[string]*compressibility{
	"1":  &compressibility{original: "640", encFrame: "15"},
	"2":  &compressibility{original: "640", encFrame: "20"},
	"3":  &compressibility{original: "640", encFrame: "25"},
	"4":  &compressibility{original: "640", encFrame: "32"},
	"5":  &compressibility{original: "640", encFrame: "42"},
	"6":  &compressibility{original: "640", encFrame: "52"},
	"7":  &compressibility{original: "640", encFrame: "60"},
	"8":  &compressibility{original: "640", encFrame: "70"},
	"9":  &compressibility{original: "640", encFrame: "86"},
	"10": &compressibility{original: "640", encFrame: "106"},
}

var AufFlag map[string]string = map[string]string{
	"3":  "8000",
	"4":  "16000",
	"46": "24000",
}

// 实现v1至v2协议的输入兼容适配;
func InputAdapter(version string, input []byte, ei *LoaderInput) (code int, err error) {
	switch version {
	case ReqV1:
		engInput := EngInputData{}
		err = proto.Unmarshal(input, &engInput)
		if err != nil {
			return frame.AigesErrorPbUnmarshal, err
		}

		// AIaaS protocol -> AIPaaS protocol
		ei.Headers = make(map[string]string)
		ei.Headers[ReqSrc] = SrcV1
		for k, v := range engInput.GetEngParam() {
			ei.Headers[k] = v
		}
		ei.Params = make(map[string]string)
		for k, v := range engInput.GetEngParam() {
			ei.Params[k] = v
		}

		ent := ei.Headers[serviceIat]
		vcn := ei.Headers[serviceTts]
		if len(vcn) != 0 {
			// 合成tts
			ei.ServiceId = vcn
			ei.ServiceName = ei.ServiceId
			ei.SyncId = int32(engInput.GetSyncId())
			ei.State = LoaderInput_ONCE // LoaderInput_ONCE
			for _, data := range engInput.GetDataList() {
				if data.Status != MetaData_ONCE {
					ei.State = LoaderInput_STREAM
					break
				}
			}

			aue := ei.Params[audioEncoding]
			if aue == "" {
				ei.Params[audioEncoding] = ttsDefaultAue // 合成缺省编解码:speex-wb
			}
			// 合成原生speex编解码，缺省level:8
			if strings.Contains(ei.Params[audioEncoding], ttsParamOrg) {
				arr := strings.Split(ei.Params[audioEncoding], ";")
				if len(arr) == 1 {
					ei.Params[audioEncoding] = ei.Params[audioEncoding] + ";8"
				}
			}
			sfl := ei.Params[ttsParamSfl]
			if sfl == "0" && aue == "lame" {
				ei.Params[ttsParamGao] = "1"
			}
			auf := ei.Params[ttsParamAuf]
			ei.Params[sampleRateV1] = AufFlag[auf]
			if len(ei.Params[sampleRateV1]) == 0 {
				return frame.AigesErrorPbAdapter, frame.ErrorPbAdapter
			}

			// 数据实体适配
			datalist := engInput.GetDataList()
			for k, _ := range datalist {
				var mt MetaDesc
				mt.Name = datalist[k].DataId
				mt.DataType = MetaDesc_DataType(datalist[k].DataType)
				mt.Attribute = make(map[string]string)
				for dk, dv := range datalist[k].Desc {
					mt.Attribute[dk] = string(dv)
				}
				mt.Attribute[Encoding] = ei.Params[ttsParamTte]
				mt.Attribute[Sequence] = strconv.Itoa(int(datalist[k].FrameId))
				ei.SyncId = int32(datalist[k].FrameId) // v1场景仅需适配单数据流场景;
				mt.Attribute[Status] = strconv.Itoa(int(datalist[k].Status))
				if ei.State == LoaderInput_STREAM {
					mt.Attribute[Status] = "2" // 合成输入适配,流式请求输入状态:2
				}
				pl := Payload{Meta: &mt, Data: datalist[k].Data}
				ei.Pl = append(ei.Pl, &pl)
			}
			// 期望输出适配
			md := MetaDesc{Name: "audio", DataType: MetaDesc_AUDIO}
			md.Attribute = make(map[string]string)
			md.Attribute[Encoding] = ei.Params[audioEncoding]
			md.Attribute[SampleRate] = ei.Params[sampleRateV1]
			ei.Expect = append(ei.Expect, &md)
		} else {
			// 听写iat
			ei.ServiceId = ent
			ei.ServiceName = ei.ServiceId
			ei.SyncId = int32(engInput.GetSyncId())
			ei.State = LoaderInput_STREAM // LoaderInput_ONCE
			for _, data := range engInput.GetDataList() {
				if data.Status != MetaData_ONCE {
					ei.State = LoaderInput_STREAM
					break
				}
			}

			var paramEncoding string
			if taue,exist:=ei.Params[audioEncoding];exist{
				paramEncoding=taue
			}
			var paramRate string
			if trate,exist:=ei.Params[sampleRateV1];exist{
				switch trate {
				case "16000","16k":
					paramRate="16000"
				case "8k","8000":
					paramRate="8000"
				default:
					return frame.AigesErrorPbAdapterIatParamInvalid,
					errors.New(frame.ErrorPbAdapterIatParamInvalid.Error()+":"+sampleRateV1)
				}
			}
			if trate,exist:=ei.Params["sample_rate"];exist{
				switch trate {
				case "16000","16k":
					paramRate="16000"
				case "8k","8000":
					paramRate="8000"
				default:
					return frame.AigesErrorPbAdapterIatParamInvalid,
						errors.New(frame.ErrorPbAdapterIatParamInvalid.Error()+":"+sampleRateV1)
				}
			}
			// 数据实体适配
			datalist := engInput.GetDataList()
			for k, _ := range datalist {
				var mt MetaDesc
				mt.Name = datalist[k].DataId
				mt.DataType = MetaDesc_DataType(datalist[k].DataType)
				mt.Attribute = make(map[string]string)
				for dk, dv := range datalist[k].Desc {
					mt.Attribute[dk] = string(dv)
				}
				if strings.TrimSpace(datalist[k].Encoding)==""{
					mt.Attribute[Encoding]=paramEncoding
				}else{
					mt.Attribute[Encoding] = datalist[k].Encoding
				}
				mt.Attribute[Sequence] = strconv.Itoa(int(datalist[k].FrameId))
				ei.SyncId = int32(datalist[k].FrameId) // v1场景仅需适配单数据流场景;
				mt.Attribute[Status] = strconv.Itoa(int(datalist[k].Status))
				mt.Attribute[FrameSize] = ei.Params[frameSizeV1]
				if paramRate!=""{
					mt.Attribute[SampleRate] = paramRate
				}else{
					fmts := strings.Split(datalist[k].Format, ";") // 拆解音频描述:采样率
					if len(fmts) == 2 {
						pairs := strings.Split(fmts[1], "=")
						if len(pairs) == 2 {
							mt.Attribute[SampleRate] = pairs[1]
						}
					}
				}
				pl := Payload{Meta: &mt, Data: datalist[k].Data}
				if pl.Meta.DataType == MetaDesc_AUDIO && len(pl.Data) == 1 {
					pl.Data = nil // 适配v1 AIaaS 听写业务输入单字节音频场景
				}
				ei.Pl = append(ei.Pl, &pl)
			}
			// 期望输出适配
			ei.Expect = nil // inst.go 适配AIaaS输出编码缺失dataId对齐问题;
		}

	case ReqV2:
		err = proto.Unmarshal(input, ei)
		if err != nil {
			return frame.AigesErrorPbUnmarshal, err
		}
		if ei.Headers == nil {
			ei.Headers = make(map[string]string)
		}
		if ei.Params==nil{
			ei.Params=make(map[string]string)
		}
		ei.Headers[ReqSrc] = SrcV2
	default:
		return frame.AigesErrorPbVersion, frame.ErrorPbVersion
	}
	return
}

// 实现v2至v1协议的输出兼容适配;
func OutputAdapter(version string, eo *LoaderOutput) (output []byte, code int, err error) {
	switch version {
	case ReqV1:
		engOutput := EngOutputData{}
		engOutput.Status = EngOutputData_DataStatus(eo.Status)
		engOutput.Err = eo.Err
		engOutput.Ret = eo.Code
		for k, _ := range eo.Pl {
			md := MetaData{
				Data:     eo.Pl[k].Data,
				DataId:   eo.Pl[k].Meta.Name,
				DataType: MetaData_DataType(eo.Pl[k].Meta.DataType),
				Encoding: eo.Pl[k].Meta.Attribute[Encoding],
			}
			fid, _ := strconv.Atoi(eo.Pl[k].Meta.Attribute[Sequence])
			md.FrameId = uint32(fid)
			md.Status = MetaData_DataStatus(eo.Status)
			if status, fg := eo.Pl[k].Meta.Attribute[Status]; fg {
				st, _ := strconv.Atoi(status)
				md.Status = MetaData_DataStatus(st)
			}
			md.Desc = make(map[string][]byte)
			for dk, dv := range eo.Pl[k].Meta.Attribute {
				md.Desc[dk] = []byte(dv)
			}
			// 输出数据为音频 & 编码为原生speex  TODO 输出 eo.Pl[k].Meta.Attribute[Encoding]与输入参数保持一致;
			if strings.Contains(md.Encoding, ttsParamOrg) {
				encRate := "8"
				arr := strings.Split(md.Encoding, ";")
				if len(arr) == 2 {
					encRate = arr[1]
				}
				tmp := SpeexEncRate[encRate]
				if tmp == nil {
					rn, _ := strconv.Atoi(encRate)
					if rn > 10 {
						encRate = "10"
					} else if rn < 1 {
						encRate = "1"
					} else {
						encRate = "8"
					}
					tmp = SpeexEncRate[encRate]
				}
				md.Desc["speex_level"] = []byte(encRate)
				md.Desc["speex_from"] = []byte(tmp.original)
				md.Desc["speex_to"] = []byte(tmp.encFrame)
				value, ok := eo.Pl[k].Meta.Attribute["key1"]
				if ok && value != "" {
					md.Desc["audio_url"] = []byte(eo.Pl[k].Meta.Attribute[AudioUrl])
				}
			}
			engOutput.DataList = append(engOutput.DataList, &md)
		}
		output, err = proto.Marshal(&engOutput)
		if err != nil {
			return nil, frame.AigesErrorPbMarshal, frame.ErrorPbMarshal
		}
	case ReqV2:
		output, err = proto.Marshal(eo)
		if err != nil {
			return nil, frame.AigesErrorPbMarshal, frame.ErrorPbMarshal
		}
	default:
		return nil, frame.AigesErrorPbVersion, frame.ErrorPbVersion
	}
	return
}
