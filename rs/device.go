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

// CameraInfo 封装了 RS2_CAMERA_INFO_* 常量
type CameraInfo int

const (
	CameraInfoName                CameraInfo = C.RS2_CAMERA_INFO_NAME
	CameraInfoSerialNumber        CameraInfo = C.RS2_CAMERA_INFO_SERIAL_NUMBER
	CameraInfoFirmwareVersion     CameraInfo = C.RS2_CAMERA_INFO_FIRMWARE_VERSION
	CameraInfoRecommendedFirmware CameraInfo = C.RS2_CAMERA_INFO_RECOMMENDED_FIRMWARE_VERSION
	CameraInfoPhysicalPort        CameraInfo = C.RS2_CAMERA_INFO_PHYSICAL_PORT
	CameraInfoDebugOpCode         CameraInfo = C.RS2_CAMERA_INFO_DEBUG_OP_CODE
	CameraInfoAdvancedMode        CameraInfo = C.RS2_CAMERA_INFO_ADVANCED_MODE
	CameraInfoProductId           CameraInfo = C.RS2_CAMERA_INFO_PRODUCT_ID
	CameraInfoCameraLocked        CameraInfo = C.RS2_CAMERA_INFO_CAMERA_LOCKED
	CameraInfoUsbTypeDescriptor   CameraInfo = C.RS2_CAMERA_INFO_USB_TYPE_DESCRIPTOR
	CameraInfoProductLine         CameraInfo = C.RS2_CAMERA_INFO_PRODUCT_LINE
	CameraInfoAsicSerialNumber    CameraInfo = C.RS2_CAMERA_INFO_ASIC_SERIAL_NUMBER
	CameraInfoFirmwareUpdateId    CameraInfo = C.RS2_CAMERA_INFO_FIRMWARE_UPDATE_ID
)

// GetInfo 获取设备的特定信息字符串
// info: CameraInfo 枚举值，如 CameraInfoSerialNumber
func (d *Device) GetInfo(info CameraInfo) (string, error) {
	var err *C.rs2_error

	// 检查是否支持该信息查询
	if C.rs2_supports_device_info(d.ptr, C.rs2_camera_info(info), &err) == 0 {
		return "", fmt.Errorf("device info %d not supported", info)
	}

	val := C.rs2_get_device_info(d.ptr, C.rs2_camera_info(info), &err)
	if err != nil {
		return "", errorFromC(err)
	}

	return C.GoString(val), nil
}

// GetUSBTypeDescriptor 获取 USB 类型描述符 (例如 "3.2" 或 "2.1")
func (d *Device) GetUSBTypeDescriptor() (string, error) {
	return d.GetInfo(CameraInfoUsbTypeDescriptor)
}

// GetPhysicalPort 获取 USB 物理端口路径
func (d *Device) GetPhysicalPort() (string, error) {
	return d.GetInfo(CameraInfoPhysicalPort)
}

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

// GetSensors 获取设备的所有传感器
// 注意：返回的 Sensor 切片中的每个元素都需要手动 Close
func (d *Device) GetSensors() ([]*Sensor, error) {
	var err *C.rs2_error

	// 1. 查询设备上的所有传感器列表
	sensorsList := C.rs2_query_sensors(d.ptr, &err)
	if err != nil {
		return nil, errorFromC(err)
	}
	defer C.rs2_delete_sensor_list(sensorsList)

	count := int(C.rs2_get_sensors_count(sensorsList, &err))
	if err != nil {
		return nil, errorFromC(err)
	}

	var sensors []*Sensor
	for i := 0; i < count; i++ {
		// 创建传感器对象
		sensorPtr := C.rs2_create_sensor(sensorsList, C.int(i), &err)
		if err != nil {
			// 如果出错，清理已创建的传感器
			for _, s := range sensors {
				s.Close()
			}
			return nil, errorFromC(err)
		}
		sensors = append(sensors, &Sensor{ptr: sensorPtr})
	}

	return sensors, nil
}

// GetDepthSensor 获取第一个深度传感器
// 这是一个便捷函数，用于快速获取深度传感器进行控制
func (d *Device) GetDepthSensor() (*Sensor, error) {
	sensors, err := d.GetSensors()
	if err != nil {
		return nil, err
	}

	for _, s := range sensors {
		var err *C.rs2_error
		// 检查是否支持深度扩展
		if C.rs2_is_sensor_extendable_to(s.ptr, C.RS2_EXTENSION_DEPTH_SENSOR, &err) != 0 {
			// 找到了深度传感器
			// 释放其他传感器
			for _, other := range sensors {
				if other != s {
					other.Close()
				}
			}
			return s, nil
		}
	}

	// 没找到，清理所有
	for _, s := range sensors {
		s.Close()
	}
	return nil, fmt.Errorf("depth sensor not found")
}

// Close 释放设备资源
func (d *Device) Close() {
	if d.ptr != nil {
		C.rs2_delete_device(d.ptr)
		d.ptr = nil
	}
}
