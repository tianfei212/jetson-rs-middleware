package rs

/*
#include <librealsense2/rs.h>
*/
import "C"

// NewContext 创建一个 RealSense 上下文
func NewContext() (*Context, error) {
	var err *C.rs2_error
	ptr := C.rs2_create_context(C.RS2_API_VERSION, &err)
	if e := checkError(err); e != nil {
		return nil, e
	}
	return &Context{ptr: ptr}, nil
}

// Close 释放上下文资源
func (ctx *Context) Close() {
	if ctx.ptr != nil {
		C.rs2_delete_context(ctx.ptr)
		ctx.ptr = nil
	}
}
