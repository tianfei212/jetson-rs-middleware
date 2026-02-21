package rs

/*
#include <librealsense2/rs.h>
#include <librealsense2/h/rs_option.h>
#include <stdlib.h>
*/
import "C"
import "unsafe"

// 传感器选项常量
const (
	OptionExposure             = C.RS2_OPTION_EXPOSURE
	OptionGain                 = C.RS2_OPTION_GAIN
	OptionLaserPower           = C.RS2_OPTION_LASER_POWER
	OptionEnableAutoExposure   = C.RS2_OPTION_ENABLE_AUTO_EXPOSURE
	OptionFilterMagnitude      = C.RS2_OPTION_FILTER_MAGNITUDE
	OptionHolesFill            = C.RS2_OPTION_HOLES_FILL
	OptionVisualPreset         = C.RS2_OPTION_VISUAL_PRESET
	OptionAsicTemperature      = C.RS2_OPTION_ASIC_TEMPERATURE
	OptionProjectorTemperature = C.RS2_OPTION_PROJECTOR_TEMPERATURE
	OptionInterCamSyncMode     = C.RS2_OPTION_INTER_CAM_SYNC_MODE
)

// VisualPreset 定义 D400 系列相机的视觉预设模式
type VisualPreset int

const (
	VisualPresetCustom          VisualPreset = C.RS2_RS400_VISUAL_PRESET_CUSTOM
	VisualPresetDefault         VisualPreset = C.RS2_RS400_VISUAL_PRESET_DEFAULT
	VisualPresetHand            VisualPreset = C.RS2_RS400_VISUAL_PRESET_HAND
	VisualPresetHighAccuracy    VisualPreset = C.RS2_RS400_VISUAL_PRESET_HIGH_ACCURACY
	VisualPresetHighDensity     VisualPreset = C.RS2_RS400_VISUAL_PRESET_HIGH_DENSITY
	VisualPresetMediumDensity   VisualPreset = C.RS2_RS400_VISUAL_PRESET_MEDIUM_DENSITY
	VisualPresetRemoveIRPattern VisualPreset = C.RS2_RS400_VISUAL_PRESET_REMOVE_IR_PATTERN
)

// SetVisualPreset 设置视觉预设模式
func (s *Sensor) SetVisualPreset(preset VisualPreset) error {
	return s.SetOption(OptionVisualPreset, float32(preset))
}

// GetVisualPreset 获取当前视觉预设模式
func (s *Sensor) GetVisualPreset() (VisualPreset, error) {
	val, err := s.GetOption(OptionVisualPreset)
	if err != nil {
		return VisualPresetCustom, err
	}
	return VisualPreset(val), nil
}

// GetDepthScale 获取深度传感器的缩放比例
// 仅对深度传感器有效
func (s *Sensor) GetDepthScale() (float32, error) {
	var err *C.rs2_error

	// 检查是否为深度传感器
	if C.rs2_is_sensor_extendable_to(s.ptr, C.RS2_EXTENSION_DEPTH_SENSOR, &err) == 0 {
		return 0, errorFromC(err) // 或者返回错误：不是深度传感器
	}

	scale := C.rs2_get_depth_scale(s.ptr, &err)
	if err != nil {
		return 0, errorFromC(err)
	}
	return float32(scale), nil
}

// SetOption 设置传感器参数
// 比如设置曝光: SetOption(OptionExposure, 1000)
func (s *Sensor) SetOption(option int, value float32) error {
	var err *C.rs2_error

	opts := (*C.rs2_options)(unsafe.Pointer(s.ptr))

	// 检查是否支持该选项
	if C.rs2_supports_option(opts, C.rs2_option(option), &err) == 0 {
		return nil // 不支持，静默失败或返回错误
	}

	C.rs2_set_option(opts, C.rs2_option(option), C.float(value), &err)
	if err != nil {
		return errorFromC(err)
	}
	return nil
}

// GetOption 获取传感器参数
func (s *Sensor) GetOption(option int) (float32, error) {
	var err *C.rs2_error

	opts := (*C.rs2_options)(unsafe.Pointer(s.ptr))

	if C.rs2_supports_option(opts, C.rs2_option(option), &err) == 0 {
		return 0, nil // 不支持
	}

	val := C.rs2_get_option(opts, C.rs2_option(option), &err)
	if err != nil {
		return 0, errorFromC(err)
	}
	return float32(val), nil
}

// Close 释放传感器资源
func (s *Sensor) Close() {
	if s.ptr != nil {
		C.rs2_delete_sensor(s.ptr)
		s.ptr = nil
	}
}
