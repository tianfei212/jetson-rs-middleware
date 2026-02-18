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

// GetData 返回帧的原始字节数据
// 注意：这只是一个指向 C 内存的引用，必须在 Frame 释放前使用
func (f *Frame) GetData() []byte {
	var err *C.rs2_error
	dataPtr := C.rs2_get_frame_data(f.ptr, &err)
	if checkError(err) != nil || dataPtr == nil {
		return nil
	}

	// 获取帧的分辨率和步长来计算总字节数
	width := int(C.rs2_get_frame_width(f.ptr, &err))
	height := int(C.rs2_get_frame_height(f.ptr, &err))
	stride := int(C.rs2_get_frame_stride_in_bytes(f.ptr, &err))
	if checkError(err) != nil {
		return nil
	}
	_ = width

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
	header.Data = uintptr(dataPtr)
	header.Len = size
	header.Cap = size

	return slice
}

// Close 极其重要！必须手动释放每一帧，否则 Jetson 会迅速崩溃
func (f *Frame) Close() {
	if f.ptr != nil {
		C.rs2_release_frame(f.ptr)
		f.ptr = nil
	}
}

// GetFrame 从 FrameSet 中提取特定类型的帧
func (fs *FrameSet) GetFrame(stream StreamType) (*Frame, error) {
	var err *C.rs2_error
	count := int(C.rs2_embedded_frames_count(fs.ptr, &err))
	if e := checkError(err); e != nil {
		return nil, e
	}
	for i := 0; i < count; i++ {
		frame := C.rs2_extract_frame(fs.ptr, C.int(i), &err)
		if e := checkError(err); e != nil {
			return nil, e
		}

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

		if StreamType(cstream) == stream || stream == StreamAny {
			C.rs2_frame_add_ref(frame, &err)
			if e := checkError(err); e != nil {
				C.rs2_release_frame(frame)
				return nil, e
			}
			C.rs2_release_frame(frame)
			return &Frame{ptr: frame}, nil
		}

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
