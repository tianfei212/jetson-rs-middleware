package rs

/*
#include <librealsense2/rs.h>
#include <librealsense2/h/rs_pipeline.h>
#include <librealsense2/h/rs_config.h>
#include <stdlib.h>
*/
import "C"

// NewPipeline 创建一个数据流管道
func NewPipeline(ctx *Context) (*Pipeline, error) {
	var err *C.rs2_error
	ptr := C.rs2_create_pipeline(ctx.ptr, &err)
	if err != nil {
		return nil, errorFromC(err)
	}
	return &Pipeline{ptr: ptr}, nil
}

// Start 启动相机流
// 如果有特定的 config（分辨率、FPS等），在这里传入
func (p *Pipeline) Start(cfg *Config) error {
	var err *C.rs2_error

	var profile *C.rs2_pipeline_profile

	if cfg != nil {
		profile = C.rs2_pipeline_start_with_config(p.ptr, cfg.ptr, &err)
	} else {
		profile = C.rs2_pipeline_start(p.ptr, &err)
	}

	if err != nil {
		return errorFromC(err)
	}

	// 释放 pipeline profile，避免内存泄漏
	// pipeline 内部已经持有了配置，这里返回的 profile 只是一个句柄
	if profile != nil {
		C.rs2_delete_pipeline_profile(profile)
	}

	return nil
}

// WaitForFrames 等待并获取下一组传感器数据
// timeout 是等待时间（毫秒），通常设为 5000
func (p *Pipeline) WaitForFrames(timeout uint) (*FrameSet, error) {
	var err *C.rs2_error

	// 这个函数返回的是一个 frameset (一组帧)
	ptr := C.rs2_pipeline_wait_for_frames(p.ptr, C.uint(timeout), &err)
	if err != nil {
		return nil, errorFromC(err)
	}

	// rs2_pipeline_wait_for_frames 返回的是 *rs2_frame
	return &FrameSet{ptr: ptr}, nil
}

// Stop 停止相机流
func (p *Pipeline) Stop() {
	var err *C.rs2_error
	if p.ptr != nil {
		C.rs2_pipeline_stop(p.ptr, &err)
	}
}

// Close 释放管道资源
func (p *Pipeline) Close() {
	if p.ptr != nil {
		C.rs2_delete_pipeline(p.ptr)
		p.ptr = nil
	}
}
