package rs

/*
#include <librealsense2/rs.h>
#include <librealsense2/h/rs_sensor.h>
#include <stdlib.h>
*/
import "C"

// StreamProfile 描述了一个具体的流配置能力
type StreamProfile struct {
	Stream    string `json:"stream"`     // 流类型 (Color, Depth, Infrared)
	Format    string `json:"format"`     // 像素格式 (Z16, RGB8, Y8)
	Width     int    `json:"width"`      // 宽度
	Height    int    `json:"height"`     // 高度
	FPS       int    `json:"fps"`        // 帧率
	IsDefault bool   `json:"is_default"` // 是否为推荐配置
}

// GetCapabilities 获取设备支持的所有流配置组合
func (d *Device) GetCapabilities() ([]StreamProfile, error) {
	var err *C.rs2_error
	var profiles []StreamProfile

	// 1. 获取所有传感器
	sensors, goErr := d.GetSensors()
	if goErr != nil {
		return nil, goErr
	}
	// 确保所有传感器都被关闭，避免内存泄漏
	defer func() {
		for _, s := range sensors {
			s.Close()
		}
	}()

	// 2. 遍历每个传感器
	for _, sensor := range sensors {
		// 获取流配置列表
		profileList := C.rs2_get_stream_profiles(sensor.ptr, &err)
		if err != nil {
			return nil, errorFromC(err)
		}
		// C.rs2_delete_stream_profiles_list(profileList) // Do not delete here, it's owned by sensor or context? Actually rs2_get_stream_profiles returns a const list usually owned by sensor.
		// Wait, the C API rs2_get_stream_profiles returns a list that needs to be released?
		// Checking librealsense2 docs: rs2_get_stream_profiles returns a list of stream profiles. The list should be released by rs2_delete_stream_profiles_list.

		count := int(C.rs2_get_stream_profiles_count(profileList, &err))
		if err != nil {
			C.rs2_delete_stream_profiles_list(profileList)
			return nil, errorFromC(err)
		}

		// 3. 遍历每个配置
		for i := 0; i < count; i++ {
			profile := C.rs2_get_stream_profile(profileList, C.int(i), &err)
			if err != nil {
				continue
			}

			// 获取流类型
			var streamType C.rs2_stream
			var format C.rs2_format
			var index C.int
			var uniqueId C.int
			var fps C.int

			C.rs2_get_stream_profile_data(profile, &streamType, &format, &index, &uniqueId, &fps, &err)
			if err != nil {
				continue
			}

			// 尝试获取视频流分辨率，如果失败则不是视频流
			var width, height C.int
			// 在调用 get_video_stream_resolution 之前，我们无法直接判断是否为视频流（CGO 环境下缺少某些宏或内联函数）
			// 因此，我们尝试调用，如果出错则忽略
			C.rs2_get_video_stream_resolution(profile, &width, &height, &err)
			if err != nil {
				// 假设这是因为该 profile 不是视频流 profile
				err = nil // 清除错误
				continue
			}

			isDefault := C.rs2_is_stream_profile_default(profile, &err)
			if err != nil {
				err = nil // 重置错误
			}

			p := StreamProfile{
				Stream:    C.GoString(C.rs2_stream_to_string(streamType)),
				Format:    C.GoString(C.rs2_format_to_string(format)),
				Width:     int(width),
				Height:    int(height),
				FPS:       int(fps),
				IsDefault: isDefault != 0,
			}
			profiles = append(profiles, p)
		}
		C.rs2_delete_stream_profiles_list(profileList)
	}

	return profiles, nil
}
