package rs

/*
#include <librealsense2/rs.h>
#include <librealsense2/h/rs_option.h>
*/
import "C"
import (
	"fmt"
)

// TelemetryData 包含从设备获取的遥测数据
type TelemetryData struct {
	AsicTemperature      float32 // ASIC 温度 (摄氏度)
	ProjectorTemperature float32 // 投影模组温度 (摄氏度)
}

// GetTelemetry 获取设备的遥测数据 (温度等)
// 注意：并不是所有设备都支持所有遥测数据，不支持的字段将为 0
func (d *Device) GetTelemetry() (*TelemetryData, error) {
	// 获取深度传感器，因为温度通常绑定在深度传感器上
	sensor, err := d.GetDepthSensor()
	if err != nil {
		return nil, fmt.Errorf("failed to get depth sensor for telemetry: %v", err)
	}
	defer sensor.Close()

	data := &TelemetryData{}

	// 1. 获取 ASIC 温度 (RS2_OPTION_ASIC_TEMPERATURE)
	if val, err := sensor.GetOption(OptionAsicTemperature); err == nil {
		data.AsicTemperature = val
	}

	// 2. 获取投影仪温度 (RS2_OPTION_PROJECTOR_TEMPERATURE)
	if val, err := sensor.GetOption(OptionProjectorTemperature); err == nil {
		data.ProjectorTemperature = val
	}

	return data, nil
}

// SyncStatus 包含相机同步状态
type SyncStatus struct {
	Mode          float32 // 同步模式 (0=Default, 1=Master, 2=Slave)
	TimestampDiff float64 // 与主相机的硬件时间戳差值 (仅 Slave 有效，需要手动计算)
}

// GetSyncMode 获取相机的同步模式
// 返回值对应: 0 (Default), 1 (Master), 2 (Slave), 3 (Full Slave) 等
func (d *Device) GetSyncMode() (float32, error) {
	sensor, err := d.GetDepthSensor()
	if err != nil {
		return 0, err
	}
	defer sensor.Close()

	return sensor.GetOption(OptionInterCamSyncMode)
}

// GetOption implementation is already in sensor.go, so we remove it from here.
