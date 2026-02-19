package main

// Version defines the current version of the shared library
const Version = "v1.0.1-20260219"

/*
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

// 简单的 C 结构体定义，方便调用者理解
typedef struct {
    int width;
    int height;
    int fps;
} StreamConfig;

typedef struct {
    float asic_temp;
    float projector_temp;
    float sync_mode;
    char usb_type[32];
} TelemetryInfo;

*/
import "C"
import (
	"sync"
	"unsafe"

	"github.com/tianfei212/jetson-rs-middleware/rs"
)

// 全局变量保持引用，防止 GC
var (
	ctx      *rs.Context
	pipeline *rs.Pipeline
	config   *rs.Config
	mu       sync.Mutex
)

//export JM_Init
func JM_Init() int {
	mu.Lock()
	defer mu.Unlock()

	var err error
	ctx, err = rs.NewContext()
	if err != nil {
		return -1
	}

	pipeline, err = rs.NewPipeline(ctx)
	if err != nil {
		ctx.Close()
		return -2
	}

	config, err = rs.NewConfig()
	if err != nil {
		pipeline.Close()
		ctx.Close()
		return -3
	}

	return 0 // Success
}

//export JM_StartStream
func JM_StartStream(width int, height int, fps int) int {
	mu.Lock()
	defer mu.Unlock()

	if pipeline == nil {
		return -1
	}

	// 启用深度和彩色流
	config.EnableStream(rs.StreamDepth, width, height, fps, rs.FormatZ16)
	config.EnableStream(rs.StreamColor, width, height, fps, rs.FormatRGB8)

	err := pipeline.Start(config)
	if err != nil {
		return -2
	}

	return 0
}

// 返回值: 0=成功, <0=失败
// rgbBuffer: 指向 RGB8 数据的指针 (大小需为 width*height*3)
// depthBuffer: 指向 Z16 数据的指针 (大小需为 width*height*2)
//
//export JM_WaitForFrames
func JM_WaitForFrames(rgbBuffer unsafe.Pointer, depthBuffer unsafe.Pointer, timeoutMs int) int {
	mu.Lock()
	defer mu.Unlock()

	if pipeline == nil {
		return -1
	}

	frames, err := pipeline.WaitForFrames(uint(timeoutMs))
	if err != nil {
		return -2
	}
	defer frames.Close()

	// 获取深度帧
	depthFrame, err := frames.GetFrame(rs.StreamDepth)
	if err == nil {
		defer depthFrame.Close()
		data := depthFrame.GetDepthData()
		size := depthFrame.GetWidth() * depthFrame.GetHeight() * 2 // 2 bytes per pixel
		// 拷贝数据到 C 缓冲区
		if depthBuffer != nil {
			C.memcpy(depthBuffer, unsafe.Pointer(&data[0]), C.size_t(size))
		}
	}

	// 获取彩色帧
	colorFrame, err := frames.GetFrame(rs.StreamColor)
	if err == nil {
		defer colorFrame.Close()
		data := colorFrame.GetRawData()
		size := colorFrame.GetWidth() * colorFrame.GetHeight() * 3 // 3 bytes per pixel
		if rgbBuffer != nil {
			C.memcpy(rgbBuffer, unsafe.Pointer(&data[0]), C.size_t(size))
		}
	}

	return 0
}

//export JM_GetTelemetry
func JM_GetTelemetry(info *C.TelemetryInfo) int {
	mu.Lock()
	defer mu.Unlock()

	if pipeline == nil {
		return -1
	}

	dev, err := pipeline.GetDevice()
	if err != nil {
		return -2
	}
	defer dev.Close()

	// 获取遥测数据
	telemetry, err := dev.GetTelemetry()
	if err == nil {
		info.asic_temp = C.float(telemetry.AsicTemperature)
		info.projector_temp = C.float(telemetry.ProjectorTemperature)
	}

	// 获取同步模式
	syncMode, err := dev.GetSyncMode()
	if err == nil {
		info.sync_mode = C.float(syncMode)
	}

	// 获取 USB 类型
	usbType, err := dev.GetUSBTypeDescriptor()
	if err == nil {
		// 简单的字符串拷贝
		cStr := C.CString(usbType)
		defer C.free(unsafe.Pointer(cStr))
		C.strcpy(&info.usb_type[0], cStr)
	}

	return 0
}

//export JM_Close
func JM_Close() {
	mu.Lock()
	defer mu.Unlock()

	if pipeline != nil {
		pipeline.Stop()
		pipeline.Close()
		pipeline = nil
	}
	if config != nil {
		config.Close()
		config = nil
	}
	if ctx != nil {
		ctx.Close()
		ctx = nil
	}
}

func main() {
	// 必须包含 main 函数，即使为空
}
