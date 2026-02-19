package rs

/*
#include <librealsense2/rs.h>
#include <stdlib.h>
*/
import "C"

// 必须引入此包

// Align 结构体封装了对齐处理器
type Align struct {
	ptr   *C.rs2_processing_block
	queue *C.rs2_frame_queue // 用于接收处理后的帧
}

// NewAlign 创建一个新的对齐处理器
// alignTo 参数指定对齐的基准流，通常是 StreamColor (彩色图)
func NewAlign(alignTo StreamType) (*Align, error) {
	var err *C.rs2_error
	// 创建对齐处理块
	ptr := C.rs2_create_align(C.rs2_stream(alignTo), &err)
	if err != nil {
		return nil, errorFromC(err)
	}

	// 创建帧队列用于接收处理结果
	queue := C.rs2_create_frame_queue(1, &err)
	if err != nil {
		C.rs2_delete_processing_block(ptr)
		return nil, errorFromC(err)
	}

	// 将处理块的输出定向到队列
	C.rs2_start_processing_queue(ptr, queue, &err)
	if err != nil {
		C.rs2_delete_processing_block(ptr)
		C.rs2_delete_frame_queue(queue)
		return nil, errorFromC(err)
	}

	return &Align{ptr: ptr, queue: queue}, nil
}

// Process 处理并对齐帧集
func (a *Align) Process(frames *FrameSet) (*FrameSet, error) {
	var err *C.rs2_error

	// 增加引用计数，因为 rs2_process_frame 会接管所有权
	// 如果不增加引用计数，frames.ptr 在 C 层被释放后，Go 层的 Frame.Close() 会导致 double-free
	C.rs2_frame_add_ref(frames.ptr, &err)
	if err != nil {
		return nil, errorFromC(err)
	}

	// 执行对齐处理
	// rs2_process_frame 返回 void，结果会放入绑定的队列中
	C.rs2_process_frame(a.ptr, frames.ptr, &err)
	if err != nil {
		return nil, errorFromC(err)
	}

	// 从队列中获取结果帧
	result := C.rs2_wait_for_frame(a.queue, 5000, &err) // 5秒超时
	if err != nil {
		return nil, errorFromC(err)
	}

	// 此时 result 是一个新的 frame 引用（通常是一个 frameset）
	return &FrameSet{ptr: result}, nil
}

// Close 释放对齐处理器的内存
func (a *Align) Close() {
	if a.ptr != nil {
		C.rs2_delete_processing_block(a.ptr)
		a.ptr = nil
	}
	if a.queue != nil {
		C.rs2_delete_frame_queue(a.queue)
		a.queue = nil
	}
}

// GetSceneFrame 这是一个辅助函数，帮助从对齐后的帧集中提取特定流
func (f *FrameSet) GetSceneFrame(stream C.rs2_stream) (*Frame, error) {
	var err *C.rs2_error

	// 从对齐后的帧集中提取单个流的帧
	ptr := C.rs2_extract_frame(f.ptr, 0, &err) // 索引通常为0，或根据流类型查找
	if err != nil {
		return nil, errorFromC(err)
	}

	return &Frame{ptr: ptr}, nil
}
