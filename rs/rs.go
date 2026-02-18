package rs

/*
#include <librealsense2/rs.h>
*/
import "C" // 注意：必须紧贴着上面的注释块，中间不能有任何文字或空行！
import (
	"fmt"
)

// 定义基础结构体封装 C 指针
type Context struct {
	ptr *C.rs2_context
}

type Pipeline struct {
	ptr *C.rs2_pipeline
}

type Config struct {
	ptr *C.rs2_config
}

type FrameSet struct {
	ptr *C.rs2_frame
}

type Frame struct {
	ptr *C.rs2_frame
}

// StreamType 映射 C 的流类型
type StreamType int

const (
	StreamAny   StreamType = C.RS2_STREAM_ANY
	StreamDepth StreamType = C.RS2_STREAM_DEPTH
	StreamColor StreamType = C.RS2_STREAM_COLOR
	StreamInfra StreamType = C.RS2_STREAM_INFRARED
	StreamFish  StreamType = C.RS2_STREAM_FISHEYE
	StreamGiro  StreamType = C.RS2_STREAM_GYRO
	StreamAccel StreamType = C.RS2_STREAM_ACCEL
)

// checkError 是内部通用的错误检查函数
// 它是解决内存泄漏的第一道防线
func checkError(err *C.rs2_error) error {
	if err == nil {
		return nil
	}

	// 转换 C 错误消息为 Go error
	errMsg := C.GoString(C.rs2_get_error_message(err))
	goErr := fmt.Errorf("realsense error: %s", errMsg)

	// 必须释放 C 分配的 error 对象
	C.rs2_free_error(err)
	return goErr
}

// GetVersion 获取底层驱动版本，用于验证链接是否成功
func GetVersion() string {
	var err *C.rs2_error
	version := int(C.rs2_get_api_version(&err))
	if checkError(err) != nil {
		return ""
	}
	major := version / 10000
	minor := (version / 100) % 100
	patch := version % 100
	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}
