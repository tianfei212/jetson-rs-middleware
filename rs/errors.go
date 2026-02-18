package rs

/*
#include <librealsense2/rs.h>
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
)

// errorFromC 将 RealSense C API 的错误转换为 Go error
func errorFromC(err *C.rs2_error) error {
	if err == nil {
		return nil
	}

	// 获取错误信息
	msg := C.rs2_get_error_message(err)
	// 获取出错的函数名
	fn := C.rs2_get_failed_function(err)
	// 获取出错的参数信息
	args := C.rs2_get_failed_args(err)

	goErr := fmt.Errorf("realsense error: %s (在函数 %s(%s) 中出错)",
		C.GoString(msg), C.GoString(fn), C.GoString(args))

	// 重要：必须释放 C 端的错误对象内存
	C.rs2_free_error(err)

	return goErr
}
