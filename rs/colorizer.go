package rs

/*
#include <librealsense2/rs.h>
#include <stdlib.h>
*/
import "C"

// Colorizer 封装了伪彩色处理器
// 用于将深度图（Z16）转换为可视化友好的彩虹图（RGB8）
type Colorizer struct {
	ptr   *C.rs2_processing_block
	queue *C.rs2_frame_queue
}

// NewColorizer 创建一个新的 Colorizer
func NewColorizer() (*Colorizer, error) {
	var err *C.rs2_error
	ptr := C.rs2_create_colorizer(&err)
	if err != nil {
		return nil, errorFromC(err)
	}

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

	return &Colorizer{ptr: ptr, queue: queue}, nil
}

// Process 处理帧，将深度帧转换为彩色帧
// 注意：返回的 Frame 需要手动 Close
func (c *Colorizer) Process(frame *Frame) (*Frame, error) {
	var err *C.rs2_error

	// 增加引用计数，因为 rs2_process_frame 会消耗一个引用
	// 否则原 frame 在 Go 端 Close 时会导致 double free，或者在这里被消耗掉
	C.rs2_frame_add_ref(frame.ptr, &err)
	if err != nil {
		return nil, errorFromC(err)
	}

	C.rs2_process_frame(c.ptr, frame.ptr, &err)
	if err != nil {
		return nil, errorFromC(err)
	}

	// 获取结果
	result := C.rs2_wait_for_frame(c.queue, 5000, &err)
	if err != nil {
		return nil, errorFromC(err)
	}

	return &Frame{ptr: result}, nil
}

// Close 释放资源
func (c *Colorizer) Close() {
	if c.ptr != nil {
		C.rs2_delete_processing_block(c.ptr)
		c.ptr = nil
	}
	if c.queue != nil {
		C.rs2_delete_frame_queue(c.queue)
		c.queue = nil
	}
}
