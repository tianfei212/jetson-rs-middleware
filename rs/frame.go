package rs

/*
#include <librealsense2/rs.h>
*/
import "C"

import (
	"fmt"
	"reflect"
	"unsafe"
)

// GetRawData 返回帧的原始字节数据
// 注意：这只是一个指向 C 内存的引用，必须在 Frame 释放前使用
func (f *Frame) GetRawData() []byte {
	var err *C.rs2_error
	dataPtr := C.rs2_get_frame_data(f.ptr, &err)
	if checkError(err) != nil || dataPtr == nil {
		return nil
	}

	// 获取帧的分辨率和步长来计算总字节数
	// width := int(C.rs2_get_frame_width(f.ptr, &err))
	height := int(C.rs2_get_frame_height(f.ptr, &err))
	stride := int(C.rs2_get_frame_stride_in_bytes(f.ptr, &err))
	if checkError(err) != nil {
		return nil
	}

	size := stride * height

	// 将 C 指针转换为 Go 的字节切片（无拷贝）
	var slice []byte
	header := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
	header.Data = uintptr(dataPtr)
	header.Len = size
	header.Cap = size

	return slice
}

// GetDepthData 将深度帧转换为 uint16 切片
func (f *Frame) GetDepthData() []uint16 {
	var err *C.rs2_error
	dataPtr := C.rs2_get_frame_data(f.ptr, &err)
	if checkError(err) != nil || dataPtr == nil {
		return nil
	}

	width := int(C.rs2_get_frame_width(f.ptr, &err))
	height := int(C.rs2_get_frame_height(f.ptr, &err))
	if checkError(err) != nil {
		return nil
	}
	size := width * height

	var slice []uint16
	header := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
	header.Data = uintptr(unsafe.Pointer(dataPtr)) // 需要转换类型
	header.Len = size
	header.Cap = size

	return slice
}

// GetWidth 获取帧宽度
func (f *Frame) GetWidth() int {
	var err *C.rs2_error
	w := C.rs2_get_frame_width(f.ptr, &err)
	if checkError(err) != nil {
		return 0
	}
	return int(w)
}

// GetHeight 获取帧高度
func (f *Frame) GetHeight() int {
	var err *C.rs2_error
	h := C.rs2_get_frame_height(f.ptr, &err)
	if checkError(err) != nil {
		return 0
	}
	return int(h)
}

// GetTimestamp 获取帧的硬件时间戳（毫秒）
func (f *Frame) GetTimestamp() (float64, error) {
	var err *C.rs2_error
	ts := C.rs2_get_frame_timestamp(f.ptr, &err)
	if err != nil {
		return 0, errorFromC(err)
	}
	return float64(ts), nil
}

// GetTimestampDomain 获取时间戳域
// RS2_TIMESTAMP_DOMAIN_HARDWARE_CLOCK (1) 表示硬件时间戳
// RS2_TIMESTAMP_DOMAIN_SYSTEM_TIME (2) 表示系统时间
func (f *Frame) GetTimestampDomain() (int, error) {
	var err *C.rs2_error
	domain := C.rs2_get_frame_timestamp_domain(f.ptr, &err)
	if err != nil {
		return 0, errorFromC(err)
	}
	return int(domain), nil
}

// Close 极其重要！必须手动释放每一帧，否则 Jetson 会迅速崩溃
func (f *Frame) Close() {
	if f.ptr != nil {
		C.rs2_release_frame(f.ptr)
		f.ptr = nil
	}
}

// GetFrame 从 FrameSet 中提取特定类型的帧
// 注意：返回的 Frame 必须手动 Close，否则会导致内存泄漏
func (fs *FrameSet) GetFrame(stream StreamType) (*Frame, error) {
	var err *C.rs2_error
	count := int(C.rs2_embedded_frames_count(fs.ptr, &err))
	if e := checkError(err); e != nil {
		return nil, e
	}

	for i := 0; i < count; i++ {
		// 提取每一帧
		frame := C.rs2_extract_frame(fs.ptr, C.int(i), &err)
		if e := checkError(err); e != nil {
			return nil, e
		}

		// 检查该帧的流类型
		// 获取 profile
		profile := C.rs2_get_frame_stream_profile(frame, &err)
		if e := checkError(err); e != nil {
			C.rs2_release_frame(frame)
			return nil, e
		}

		var cstream C.rs2_stream
		var format C.rs2_format
		var index C.int
		var uniqueID C.int
		var framerate C.int
		C.rs2_get_stream_profile_data(profile, &cstream, &format, &index, &uniqueID, &framerate, &err)
		if e := checkError(err); e != nil {
			C.rs2_release_frame(frame)
			return nil, e
		}

		// 匹配流类型
		if StreamType(cstream) == stream || stream == StreamAny {
			// rs2_extract_frame 返回的 frame 引用计数已经是 +1 的
			// 我们直接封装返回
			return &Frame{ptr: frame}, nil
		}

		// 不匹配，释放该帧
		C.rs2_release_frame(frame)
	}

	return nil, fmt.Errorf("frame not found for stream %v", stream)
}

// Close 释放帧集
func (fs *FrameSet) Close() {
	if fs.ptr != nil {
		C.rs2_release_frame(fs.ptr)
		fs.ptr = nil
	}
}

// GetDepthFrame 获取深度帧的快捷方法
func (fs *FrameSet) GetDepthFrame() (*Frame, error) {
	return fs.GetFrame(StreamDepth)
}

// GetColorFrame 获取彩色帧的快捷方法
func (fs *FrameSet) GetColorFrame() (*Frame, error) {
	return fs.GetFrame(StreamColor)
}
