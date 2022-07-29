package util

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"os/exec"
)

type NvidiaSmiLog struct {
	XMLName       xml.Name `xml:"nvidia_smi_log"`
	Timestamp     string   `xml:"timestamp"`
	DriverVersion string   `xml:"driver_version"`
	CudaVersion   string   `xml:"cuda_version"`
	AttachedGpus  int      `xml:"attached_gpus"`
	Gpus          []GpuS   `xml:"gpu"`
}

type MigModeS struct {
	XMLName    xml.Name `xml:"mig_mode"`
	CurrentMig string   `xml:"current_mig"`
	PendingMig string   `xml:"pending_mig"`
}

type DriverModelS struct {
	XMLName   xml.Name `xml:"driver_model"`
	CurrentDm string   `xml:"current_dm"`
	PendingDm string   `xml:"pending_dm"`
}

type InforomVersionS struct {
	XMLName    xml.Name `xml:"inforom_version"`
	ImgVersion string   `xml:"img_version"`
	OemObject  string   `xml:"oem_object"`
	EccObject  string   `xml:"ecc_object"`
	PwrObject  string   `xml:"pwr_object"`
}

type GpuOperationModeS struct {
	XMLName    xml.Name `xml:"gpu_operation_mode"`
	CurrentGom string   `xml:"current_gom"`
	PendingGom string   `xml:"pending_gom"`
}

type GpuVirtualizationModeS struct {
	XMLName            xml.Name `xml:"gpu_virtualization_mode"`
	VirtualizationMode string   `xml:"virtualization_mode"`
	HostVgpuMode       string   `xml:"host_vgpu_mode"`
}

type IbmnpuS struct {
	XMLName             xml.Name `xml:"ibmnpu"`
	RelaxedOrderingMode string   `xml:"relaxed_ordering_mode"`
}

type PciS struct {
	XMLName               xml.Name        `xml:"pci"`
	PciBus                string          `xml:"pci_bus"`
	PciDevice             string          `xml:"pci_device"`
	PciDomain             string          `xml:"pci_domain"`
	PciDeviceId           string          `xml:"pci_device_id"`
	PciBusId              string          `xml:"pci_bus_id"`
	PciSubSystemId        string          `xml:"pci_sub_system_id"`
	PciGpuLinkInfo        PciGpuLinkInfoS `xml:"pci_gpu_link_info"`
	PciBridgeChip         PciBridgeChipS  `xml:"pci_bridge_chip"`
	ReplayCounter         int             `xml:"replay_counter"`
	ReplayRolloverCounter int             `xml:"replay_rollover_counter"`
	TxUtil                string          `xml:"tx_util"`
	RxUtil                string          `xml:"rx_util"`
}

type PciGpuLinkInfoS struct {
	XMLName    xml.Name    `xml:"pci_gpu_link_info"`
	PcieGen    PcieGenS    `xml:"pcie_gen"`
	LinkWidths LinkWidthsS `xml:"link_widths"`
}

type PcieGenS struct {
	XMLName        xml.Name `xml:"pcie_gen"`
	MaxLinkGen     int      `xml:"max_link_gen"`
	CurrentLinkGen string   `xml:"current_link_gen"`
}

type LinkWidthsS struct {
	XMLName        xml.Name `xml:"link_widths"`
	MaxLinkGen     int      `xml:"max_link_gen"`
	CurrentLinkGen string   `xml:"current_link_gen"`
}

type PciBridgeChipS struct {
	XMLName        xml.Name `xml:"pci_bridge_chip"`
	BridgeChipType string   `xml:"bridge_chip_type"`
	BridgeChipFw   string   `xml:"bridge_chip_fw"`
}

type ClocksThrottleReasonsS struct {
	XMLName                                  xml.Name `xml:"clocks_throttle_reasons"`
	ClocksThrottleReasonGpuIdle              string   `xml:"clocks_throttle_reason_gpu_idle"`
	ClocksThrottleReasonApplications         string   `xml:"clocks_throttle_reason_applications"`
	ClocksThrottleReasonSwPowerCap           string   `xml:"clocks_throttle_reason_sw_power_cap"`
	ClocksThrottleReasonHwSlowdown           string   `xml:"clocks_throttle_reason_hw_slowdown"`
	ClocksThrottleReasonHwThermalSlowdown    string   `xml:"clocks_throttle_reason_hw_thermal_slowdown"`
	ClocksThrottleReasonHwPowerBrakeSlowdown string   `xml:"clocks_throttle_reason_hw_power_brake_slowdown"`
	ClocksThrottleReasonSyncBoost            string   `xml:"clocks_throttle_reason_sync_boost"`
	ClocksThrottleReasonSwThermalSlowdown    string   `xml:"clocks_throttle_reason_sw_thermal_slowdown"`
	ClocksThrottleReasonDisplayClocksSetting string   `xml:"clocks_throttle_reason_display_clocks_setting"`
}

type FbMemoryUsageS struct {
	XMLName xml.Name `xml:"fb_memory_usage"`
	Total   string   `xml:"total"`
	Used    string   `xml:"used"`
	Free    string   `xml:"free"`
}

type Bar1MemoryUsageS struct {
	XMLName xml.Name `xml:"bar1_memory_usage"`
	Total   string   `xml:"total"`
	Used    string   `xml:"used"`
	Free    string   `xml:"free"`
}

type UtilizationS struct {
	XMLName     xml.Name `xml:"utilization"`
	GpuUtil     string   `xml:"gpu_util"`
	MemoryUtil  string   `xml:"memory_util"`
	EncoderUtil string   `xml:"encoder_util"`
	DecoderUtil string   `xml:"decoder_util"`
}

type EncoderStatusS struct {
	XMLName        xml.Name `xml:"encoder_status"`
	SessionCount   int      `xml:"session_count"`
	AverageFps     int      `xml:"average_fps"`
	AverageLatency int      `xml:"average_latency"`
}

type FbcStatsS struct {
	XMLName        xml.Name `xml:"fbc_stats"`
	SessionCount   int      `xml:"session_count"`
	AverageFps     int      `xml:"average_fps"`
	AverageLatency int      `xml:"average_latency"`
}

type EccModeS struct {
	XMLName    xml.Name `xml:"ecc_mode"`
	CurrentEcc string   `xml:"current_ecc"`
	PendingEcc string   `xml:"pending_ecc"`
}

type EccErrorsS struct {
	XMLName   xml.Name   `xml:"ecc_errors"`
	Volatile  VolatileS  `xml:"volatile"`
	Aggregate AggregateS `xml:"aggregate"`
}

type VolatileS struct {
	XMLName   xml.Name   `xml:"volatile"`
	SingleBit SingleBitS `xml:"single_bit"`
	DoubleBit DoubleBitS `xml:"double_bit"`
}

type SingleBitS struct {
	XMLName       xml.Name `xml:"single_bit"`
	DeviceMemory  string   `xml:"device_memory"`
	RegisterFile  string   `xml:"register_file"`
	L1Cache       string   `xml:"l1_cache"`
	L2Cache       string   `xml:"l2_cache"`
	TextureMemory string   `xml:"texture_memory"`
	TextureShm    string   `xml:"texture_shm"`
	Cbu           string   `xml:"cbu"`
	Total         string   `xml:"total"`
}
type DoubleBitS struct {
	XMLName       xml.Name `xml:"double_bit"`
	DeviceMemory  string   `xml:"device_memory"`
	RegisterFile  string   `xml:"register_file"`
	L1Cache       string   `xml:"l1_cache"`
	L2Cache       string   `xml:"l2_cache"`
	TextureMemory string   `xml:"texture_memory"`
	TextureShm    string   `xml:"texture_shm"`
	Cbu           string   `xml:"cbu"`
	Total         string   `xml:"total"`
}

type AggregateS struct {
	XMLName   xml.Name   `xml:"aggregate"`
	SingleBit SingleBitS `xml:"single_bit"`
	DoubleBit DoubleBitS `xml:"double_bit"`
}

type RetiredPagesS struct {
	XMLName                     xml.Name                     `xml:"retired_pages"`
	MultipleSingleBitRetirement MultipleSingleBitRetirementS `xml:"multiple_single_bit_retirement"`
	DoubleBitRetirement         DoubleBitRetirementS         `xml:"double_bit_retirement"`
	PendingBlacklist            string                       `xml:"pending_blacklist"`
	PendingRetirement           string                       `xml:"pending_retirement"`
}

type MultipleSingleBitRetirementS struct {
	XMLName         xml.Name `xml:"multiple_single_bit_retirement"`
	RetiredCount    int      `xml:"retired_count"`
	RetiredPagelist string   `xml:"retired_pagelist"`
}

type DoubleBitRetirementS struct {
	XMLName         xml.Name `xml:"double_bit_retirement"`
	RetiredCount    int      `xml:"retired_count"`
	RetiredPagelist string   `xml:"retired_pagelist"`
}

type TemperatureS struct {
	XMLName                xml.Name `xml:"temperature"`
	GpuTemp                string   `xml:"gpu_temp"`
	GpuTempMaxThreshold    string   `xml:"gpu_temp_max_threshold"`
	GpuTempSlowThreshold   string   `xml:"gpu_temp_slow_threshold"`
	GpuTempMaxGpuThreshold string   `xml:"gpu_temp_max_gpu_threshold"`
	GpuTargetTemperature   string   `xml:"gpu_target_temperature"`
	MemoryTemp             string   `xml:"memory_temp"`
	GpuTempMaxMemThreshold string   `xml:"gpu_temp_max_mem_threshold"`
}

type SupportedGpuTargetTempS struct {
	XMLName          xml.Name `xml:"supported_gpu_target_temp"`
	GpuTargetTempMin string   `xml:"gpu_target_temp_min"`
	GpuTargetTempMax string   `xml:"gpu_target_temp_max"`
}

type PowerReadingsS struct {
	XMLName            xml.Name `xml:"power_readings"`
	PowerState         string   `xml:"power_state"`
	PowerManager       string   `xml:"power_manager"`
	PowerDraw          string   `xml:"power_draw"`
	PowerLimit         string   `xml:"power_limit"`
	DefaultPowerLimit  string   `xml:"default_power_limit"`
	EnforcedPowerLimit string   `xml:"enforced_power_limit"`
	MinPowerLimit      string   `xml:"min_power_limit"`
	MaxPowerLimit      string   `xml:"max_power_limit"`
}

type ClocksS struct {
	XMLName       xml.Name `xml:"clocks"`
	GraphicsClock string   `xml:"graphics_clock"`
	SmClock       string   `xml:"sm_clock"`
	MemClock      string   `xml:"mem_clock"`
	VideoClock    string   `xml:"video_clock"`
}

type ApplicationsClocksS struct {
	XMLName       xml.Name `xml:"applications_clocks"`
	GraphicsClock string   `xml:"graphics_clock"`
	MemClock      string   `xml:"mem_clock"`
}

type DefaultApplicationsClocksS struct {
	XMLName       xml.Name `xml:"default_applications_clocks"`
	GraphicsClock string   `xml:"graphics_clock"`
	MemClock      string   `xml:"mem_clock"`
}

type MaxClockS struct {
	XMLName       xml.Name `xml:"max_clocks"`
	GraphicsClock string   `xml:"graphics_clock"`
	SmClock       string   `xml:"sm_clock"`
	MemClock      string   `xml:"mem_clock"`
	VideoClock    string   `xml:"video_clock"`
}

type MaxCustomerBoostClocksS struct {
	XMLName       xml.Name `xml:"max_customer_boost_clocks"`
	GraphicsClock string   `xml:"graphics_clock"`
}

type ClockPolicyS struct {
	XMLName          xml.Name `xml:"clock_policy"`
	AutoBoost        string   `xml:"auto_boost"`
	AutoBoostDefault string   `xml:"auto_boost_default"`
}

type SupportedClocksS struct {
	XMLName           xml.Name             `xml:"supported_clocks"`
	SupportedMemClock []SupportedMemClockS `xml:"supported_mem_clock"`
}

type SupportedMemClockS struct {
	XMLName                xml.Name `xml:"supported_mem_clock"`
	Value                  string   `xml:"value"`
	SupportedGraphicsClock []string `xml:"supported_graphics_clock"`
}

type ProcessesS struct {
	XMLName      xml.Name       `xml:"processes"`
	ProcessInfos []ProcessInfoS `xml:"process_info"`
}

type ProcessInfoS struct {
	XMLName           xml.Name `xml:"process_info"`
	GpuInstanceId     string   `xml:"gpu_instance_id"`
	ComputeInstanceId string   `xml:"compute_instance_id"`
	Pid               int      `xml:"pid"`
	Type              string   `xml:"type"`
	ProcessName       string   `xml:"process_name"`
	UsedMemory        string   `xml:"used_memory"`
}

type AccountedProcessesS struct {
	XMLName xml.Name `xml:"accounted_processes"`
}

type GpuS struct {
	XMLName                   xml.Name                   `xml:"gpu"`
	ID                        string                     `xml:"id"`
	ProductName               string                     `xml:"product_name"`
	ProductBrand              string                     `xml:"product_brand"`
	DisplayMode               string                     `xml:"display_mode"`
	DisplayActive             string                     `xml:"display_active"`
	PersistenceMode           string                     `xml:"persistence_mode"`
	MigMode                   MigModeS                   `xml:"mig_mode"`
	MigDevices                string                     `xml:"mig_devices"`
	AccountingMode            string                     `xml:"accounting_mode"`
	AccountingModeBufferSize  int                        `xml:"accounting_mode_buffer_size"`
	DriverModel               DriverModelS               `xml:"driver_model"`
	Serial                    string                     `xml:"serial"`
	Uuid                      string                     `xml:"uuid"`
	MinorNumber               int                        `xml:"minor_number"`
	VbiosVersion              string                     `xml:"vbios_version"`
	MultigpuBoard             string                     `xml:"multigpu_board"`
	BoardId                   string                     `xml:"board_id"`
	GpuPartNumber             string                     `xml:"gpu_part_number"`
	InforomVersion            InforomVersionS            `xml:"inforom_version"`
	GpuOperationMode          GpuOperationModeS          `xml:"gpu_operation_mode"`
	GpuVirtualizationMode     GpuVirtualizationModeS     `xml:"gpu_virtualization_mode"`
	Ibmnpu                    IbmnpuS                    `xml:"ibmnpu"`
	Pci                       PciS                       `xml:"pci"`
	FanSpeed                  string                     `xml:"fan_speed"`
	PerformanceState          string                     `xml:"performance_state"`
	ClocksThrottleReasons     ClocksThrottleReasonsS     `xml:"clocks_throttle_reasons"`
	FbMemoryUsage             FbMemoryUsageS             `xml:"fb_memory_usage"`
	Bar1MemoryUsage           Bar1MemoryUsageS           `xml:"bar1_memory_usage"`
	ComputeMode               string                     `xml:"compute_mode"`
	Utilization               UtilizationS               `xml:"utilization"`
	EncoderStatus             EncoderStatusS             `xml:"encoder_status"`
	FbcStats                  FbcStatsS                  `xml:"fbc_stats"`
	EccMode                   EccModeS                   `xml:"ecc_mode"`
	EccErrors                 EccErrorsS                 `xml:"ecc_errors"`
	RetiredPages              RetiredPagesS              `xml:"retired_pages"`
	RamappedRows              string                     `xml:"ramapped_rows"`
	Temperature               TemperatureS               `xml:"temperature"`
	SupportedGpuTargetTemp    SupportedGpuTargetTempS    `xml:"supported_gpu_target_temp"`
	PowerReadings             PowerReadingsS             `xml:"power_readings"`
	Clocks                    ClocksS                    `xml:"clocks"`
	ApplicationsClocks        ApplicationsClocksS        `xml:"applications_clocks"`
	DefaultApplicationsClocks DefaultApplicationsClocksS `xml:"default_applications_clocks"`
	MaxClocks                 MaxClockS                  `xml:"max_clocks"`
	MaxCustomerBoostClocks    MaxCustomerBoostClocksS    `xml:"max_customer_boost_clocks"`
	ClockPolicy               ClockPolicyS               `xml:"clock_policy"`
	SupportedClocks           SupportedClocksS           `xml:"supported_clocks"`
	Processes                 ProcessesS                 `xml:"processes"`
	AccountedProcesses        AccountedProcessesS        `xml:"accounted_processes"`
}

// GpuProcesses 获得Nvidia的进程信息
func GpuProcesses() ([]ProcessInfoS, error) {
	c := exec.Command("nvidia-smi", "-x", "-q")
	stdout, err := c.StdoutPipe()
	if err != nil {
		log.Fatal(err)
		return []ProcessInfoS{}, err
	}
	defer stdout.Close()
	if err := c.Start(); err != nil {
		log.Fatal(err)
		return []ProcessInfoS{}, err
	}
	opBytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		log.Fatal(err)
		return []ProcessInfoS{}, err
	}

	x := NvidiaSmiLog{}
	err = xml.Unmarshal(opBytes, &x)
	var processes []ProcessInfoS
	for _, g := range x.Gpus {
		processes = append(processes, g.Processes.ProcessInfos...)
	}
	return processes, nil
}
