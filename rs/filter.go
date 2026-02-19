package rs

/*
#include <librealsense2/rs.h>
#include <stdlib.h>
*/
import "C"
import "unsafe"

// Filter 封装了各类图像处理过滤器
type Filter struct {
	ptr   *C.rs2_processing_block
	queue *C.rs2_frame_queue
}

// newFilter 内部辅助函数，统一初始化过滤器
func newFilter(ptr *C.rs2_processing_block) (*Filter, error) {
	var err *C.rs2_error

	// 创建帧队列用于接收处理结果
	queue := C.rs2_create_frame_queue(1, &err)
	if err != nil {
		C.rs2_delete_processing_block(ptr)
		return nil, errorFromC(err)
	}

	// 启动处理块，将结果输出到队列
	C.rs2_start_processing_queue(ptr, queue, &err)
	if err != nil {
		C.rs2_delete_processing_block(ptr)
		C.rs2_delete_frame_queue(queue)
		return nil, errorFromC(err)
	}

	return &Filter{ptr: ptr, queue: queue}, nil
}

// NewDecimationFilter 创建降采样过滤器
// magnitude: 降采样倍数 (2-8)
func NewDecimationFilter() (*Filter, error) {
	var err *C.rs2_error
	ptr := C.rs2_create_decimation_filter_block(&err)
	if err != nil {
		return nil, errorFromC(err)
	}
	return newFilter(ptr)
}

// NewSpatialFilter 创建空间过滤器
// 用于平滑深度数据，保留边缘
func NewSpatialFilter() (*Filter, error) {
	var err *C.rs2_error
	ptr := C.rs2_create_spatial_filter_block(&err)
	if err != nil {
		return nil, errorFromC(err)
	}
	return newFilter(ptr)
}

// NewTemporalFilter 创建时间过滤器
// 利用多帧数据进行平滑，减少噪点
func NewTemporalFilter() (*Filter, error) {
	var err *C.rs2_error
	ptr := C.rs2_create_temporal_filter_block(&err)
	if err != nil {
		return nil, errorFromC(err)
	}
	return newFilter(ptr)
}

// NewHoleFillingFilter 创建孔洞填充过滤器
func NewHoleFillingFilter() (*Filter, error) {
	var err *C.rs2_error
	ptr := C.rs2_create_hole_filling_filter_block(&err)
	if err != nil {
		return nil, errorFromC(err)
	}
	return newFilter(ptr)
}

// Process 处理帧
// 注意：会增加输入帧的引用计数，原帧仍需调用者释放
func (f *Filter) Process(frame *Frame) (*Frame, error) {
	var err *C.rs2_error

	// 增加引用计数，因为 rs2_process_frame 会消耗一个引用
	// 如果不增加，Go 层的 frame.Close() 会导致 double free
	C.rs2_frame_add_ref(frame.ptr, &err)
	if err != nil {
		return nil, errorFromC(err)
	}

	// 处理
	C.rs2_process_frame(f.ptr, frame.ptr, &err)
	if err != nil {
		return nil, errorFromC(err)
	}

	// 等待结果
	result := C.rs2_wait_for_frame(f.queue, 5000, &err)
	if err != nil {
		return nil, errorFromC(err)
	}

	return &Frame{ptr: result}, nil
}

// SetOption 设置过滤器参数
// option: 选项枚举 (如 C.RS2_OPTION_FILTER_MAGNITUDE)
// value: 值
func (f *Filter) SetOption(option int, value float32) error {
	var err *C.rs2_error
	// processing block 也是一种 options interface
	// 直接转换为 options 指针
	opts := (*C.rs2_options)(unsafe.Pointer(f.ptr))

	if C.rs2_supports_option(opts, C.rs2_option(option), &err) == 0 {
		return nil // 或者返回错误说不支持
	}

	C.rs2_set_option(opts, C.rs2_option(option), C.float(value), &err)
	if err != nil {
		return errorFromC(err)
	}
	return nil
}

// Close 释放资源
func (f *Filter) Close() {
	if f.ptr != nil {
		C.rs2_delete_processing_block(f.ptr)
		f.ptr = nil
	}
	if f.queue != nil {
		C.rs2_delete_frame_queue(f.queue)
		f.queue = nil
	}
}
