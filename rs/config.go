package rs

/*
#include <librealsense2/rs.h>
#include <librealsense2/h/rs_config.h>
#include <stdlib.h>
*/
import "C"

// Format 定义数据格式
type Format int

const (
	FormatAny  Format = C.RS2_FORMAT_ANY
	FormatZ16  Format = C.RS2_FORMAT_Z16  // 深度图标准格式 [cite: 54]
	FormatRGB8 Format = C.RS2_FORMAT_RGB8 // 彩色图标准格式 [cite: 54]
)

// NewConfig 初始化配置容器[cite:29,30]
func NewConfig() (*Config, error) {
	var err *C.rs2_error
	ptr := C.rs2_create_config(&err)
	if err != nil {
		return nil, errorFromC(err)
	}
	return &Config{ptr: ptr}, nil
}

// EnableStream 设置流的具体参数（分辨率、FPS、格式） [cite: 31, 32]
func (c *Config) EnableStream(stype StreamType, w, h, fps int, format Format) error {
	var err *C.rs2_error

	C.rs2_config_enable_stream(
		c.ptr,
		C.rs2_stream(stype),
		C.int(0), // 传感器索引，通常为 0
		C.int(w),
		C.int(h),
		C.rs2_format(format),
		C.int(fps),
		&err,
	)

	if err != nil {
		return errorFromC(err)
	}
	return nil
}

// Close 释放配置对象内存
func (c *Config) Close() {
	if c.ptr != nil {
		C.rs2_delete_config(c.ptr)
		c.ptr = nil
	}
}
