package rs

/*
#include <librealsense2/rs.h>
#include <librealsense2/h/rs_pipeline.h>
#include <librealsense2/h/rs_device.h>
#include <librealsense2/h/rs_sensor.h>
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
)

// GetDevice 从管道获取当前活动的设备
// 通常在 pipeline.Start() 之后调用，用于获取硬件参数
func (p *Pipeline) GetDevice() (*Device, error) {
	var err *C.rs2_error

	// 从管道的当前配置文件中提取设备
	selection := C.rs2_pipeline_get_active_profile(p.ptr, &err)
	if err != nil {
		return nil, errorFromC(err)
	}
	defer C.rs2_delete_pipeline_profile(selection)

	dev := C.rs2_pipeline_profile_get_device(selection, &err)
	if err != nil {
		return nil, errorFromC(err)
	}

	return &Device{ptr: dev}, nil
}

// GetDepthScale 获取深度传感器的缩放比例
// D455 通常返回 0.001，意味着数值 1000 代表 1米
func (d *Device) GetDepthScale() (float32, error) {
	var err *C.rs2_error

	// 1. 查询设备上的所有传感器
	sensors := C.rs2_query_sensors(d.ptr, &err)
	if err != nil {
		return 0, errorFromC(err)
	}
	defer C.rs2_delete_sensor_list(sensors)

	count := int(C.rs2_get_sensors_count(sensors, &err))
	if err != nil {
		return 0, errorFromC(err)
	}

	// 2. 遍历传感器寻找深度传感器
	for i := 0; i < count; i++ {
		sensorPtr := C.rs2_create_sensor(sensors, C.int(i), &err)
		if err != nil {
			continue
		}

		// 检查该传感器是否支持深度缩放
		// 在 C API 中，这通常通过检查传感器是否能提供深度流来判断
		if C.rs2_is_sensor_extendable_to(sensorPtr, C.RS2_EXTENSION_DEPTH_SENSOR, &err) != 0 {
			scale := C.rs2_get_depth_scale(sensorPtr, &err)
			C.rs2_delete_sensor(sensorPtr) // 及时释放
			if err != nil {
				return 0, errorFromC(err)
			}
			return float32(scale), nil
		}
		C.rs2_delete_sensor(sensorPtr)
	}

	return 0, fmt.Errorf("未在设备中找到深度传感器")
}

// Close 释放设备资源
func (d *Device) Close() {
	if d.ptr != nil {
		C.rs2_delete_device(d.ptr)
		d.ptr = nil
	}
}
