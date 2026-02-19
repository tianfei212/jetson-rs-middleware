# Jetson RealSense Middleware 开发指南

本文档旨在指导开发者如何使用 `jetson-rs-middleware` 进行二次开发，涵盖了从初始化、数据流获取、图像处理到硬件控制的全流程。

---

## 1. 核心概念

在使用本中间件前，请先理解以下核心对象：

*   **Context (`rs.Context`)**: 全局单例，管理所有连接的 RealSense 设备。
*   **Pipeline (`rs.Pipeline`)**: 高级接口，封装了设备配置、数据流启动和帧同步逻辑，是大多数应用场景的入口。
*   **Device (`rs.Device`)**: 代表物理相机，用于查询硬件信息（序列号、USB类型）、设置传感器参数（曝光、增益）和获取遥测数据。
*   **Config (`rs.Config`)**: 用于配置 Pipeline，指定需要开启的流类型（Color/Depth/Infrared）、分辨率和帧率。
*   **Frame (`rs.Frame`)**: 图像数据容器，包含像素数据、元数据（时间戳、帧号）等。
*   **ProcessingBlock (`rs.Colorizer`, `rs.Filter`)**: 图像处理单元，用于深度图着色、滤波降噪等。

---

## 2. 基础开发流程

一个标准的 RealSense 应用开发流程如下：

### 2.1 初始化与配置

```go
package main

import (
    "log"
    "github.com/tianfei212/jetson-rs-middleware/rs"
)

func main() {
    // 1. 创建上下文
    ctx, err := rs.NewContext()
    if err != nil {
        log.Fatal(err)
    }
    defer ctx.Close() // 务必记得关闭释放资源

    // 2. 创建管道
    pipeline, err := rs.NewPipeline(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer pipeline.Close()

    // 3. 配置流参数
    cfg, err := rs.NewConfig()
    if err != nil {
        log.Fatal(err)
    }
    defer cfg.Close()

    // 启用深度流：640x480, 30fps, Z16格式
    cfg.EnableStream(rs.StreamDepth, 640, 480, 30, rs.FormatZ16)
    // 启用彩色流：640x480, 30fps, RGB8格式
    cfg.EnableStream(rs.StreamColor, 640, 480, 30, rs.FormatRGB8)
}
```

### 2.2 启动与数据循环

```go
    // 4. 启动管道
    if err := pipeline.Start(cfg); err != nil {
        log.Fatal(err)
    }
    defer pipeline.Stop()

    // 5. 数据处理循环
    for {
        // 等待新的一组帧（包含深度和彩色）
        // 超时时间设置为 1000ms
        frames, err := pipeline.WaitForFrames(1000)
        if err != nil {
            log.Printf("Wait for frames failed: %v", err)
            continue
        }

        // 处理帧数据...
        processFrames(frames)

        // 必须手动释放 FrameSet，否则会内存泄漏！
        frames.Close()
    }
```

### 2.3 获取帧数据

```go
func processFrames(fs *rs.FrameSet) {
    // 获取深度帧
    depthFrame, err := fs.GetDepthFrame()
    if err == nil {
        defer depthFrame.Close() // 释放单个帧引用
        
        // 获取原始深度数据 (16-bit)
        depthData := depthFrame.GetDepthData()
        width := depthFrame.GetWidth()
        height := depthFrame.GetHeight()
        timestamp := depthFrame.GetTimestamp() // 硬件时间戳
        
        // ... 业务逻辑
    }

    // 获取彩色帧
    colorFrame, err := fs.GetColorFrame()
    if err == nil {
        defer colorFrame.Close()
        
        // 获取 RGB 数据
        rgbData := colorFrame.GetData()
        // ...
    }
}
```

---

## 3. 进阶功能

### 3.1 图像对齐 (Alignment)

D455 的 RGB 摄像头和深度摄像头位置不同，导致图像视野不重合。使用 `rs.Align` 可以将深度图对齐到 RGB 视角。

```go
    // 初始化对齐器 (对齐到彩色流)
    align, _ := rs.NewAlign(rs.StreamColor)
    defer align.Close()

    // 在循环中：
    alignedFrames, err := align.Process(frames)
    if err != nil {
        return
    }
    defer alignedFrames.Close() // 释放对齐后的帧集

    // 此时获取的深度帧已与彩色帧像素对齐
    alignedDepth, _ := alignedFrames.GetDepthFrame()
    // ...
```

### 3.2 深度图着色 (Colorizer)

将 16-bit 的灰度深度图转换为可视化的伪彩色热力图。

```go
    // 初始化着色器
    colorizer, _ := rs.NewColorizer()
    defer colorizer.Close()

    // 处理深度帧
    colorizedFrame, err := colorizer.Process(depthFrame)
    if err == nil {
        defer colorizedFrame.Close()
        // 获取渲染后的 RGB 数据
        visualData := colorizedFrame.GetData() 
    }
```

### 3.3 滤波器 (Filters)

使用滤波器提升深度图质量。

```go
    // 1. 降采样 (减少计算量)
    decimation, _ := rs.NewDecimationFilter()
    // 2. 空间滤波 (平滑噪点，填充孔洞)
    spatial, _ := rs.NewSpatialFilter()
    // 3. 时间滤波 (利用历史帧平滑数据)
    temporal, _ := rs.NewTemporalFilter()
    
    defer decimation.Close()
    defer spatial.Close()
    defer temporal.Close()

    // 链式处理
    f1, _ := decimation.Process(depthFrame)
    defer f1.Close()
    
    f2, _ := spatial.Process(f1)
    defer f2.Close()
    
    resultFrame, _ := temporal.Process(f2)
    defer resultFrame.Close()
```

### 3.4 硬件控制与遥测

```go
    // 获取设备句柄
    dev, _ := pipeline.GetDevice()
    defer dev.Close()

    // 1. 获取传感器并控制参数
    sensors, _ := dev.GetSensors()
    for _, s := range sensors {
        defer s.Close()
        // 检查是否支持曝光控制
        if s.SupportsOption(rs.OptionExposure) {
            // 设置曝光值
            s.SetOption(rs.OptionExposure, 1500)
        }
    }

    // 2. 获取遥测数据
    telemetry, _ := dev.GetTelemetry()
    fmt.Printf("ASIC 温度: %.2f\n", telemetry.AsicTemperature)

    // 3. 检查 USB 连接
    usbType, _ := dev.GetUSBTypeDescriptor()
    fmt.Println("USB 连接:", usbType) // 应为 "3.2"
```

---

## 4. Jetson 平台注意事项

1.  **内存管理**: 
    *   Go 的 GC 无法自动回收 CGO 分配的 `rs2_frame` 对象。
    *   **必须** 对每个 `New...` 创建的对象和 `Get...` 返回的 Frame/FrameSet 调用 `Close()`。
    *   建议使用 `defer` 确保资源释放，避免内存泄漏导致 Jetson OOM。

2.  **性能优化**:
    *   尽量复用 `rs.Align` 和 `rs.Filter` 对象，不要在循环内重复创建。
    *   获取数据 (`GetData`, `GetDepthData`) 会发生一次内存拷贝（C -> Go），这是为了 Go 内存安全。在高性能场景下，确保处理逻辑尽量快，避免阻塞数据流。

3.  **USB 带宽**:
    *   Jetson 的 USB 控制器带宽有限。如果开启高分辨率（如 1280x720 @ 30fps）双流，请确保使用高质量 USB 3.0 线缆，并连接到原生 USB 3.0 接口。

---

## 5. 常见问题排查

*   **"Frame didn't arrive within 5000"**: 
    *   检查 USB 线是否松动。
    *   检查是否有其他进程占用了相机。
    *   尝试调高 `WaitForFrames` 的超时时间。

*   **"uvc_video: Failed to query (GET_CUR) UVC control"**:
    *   Linux 内核 UVC 驱动问题，通常不影响使用。如果频繁出现，建议升级 librealsense2 或内核补丁。

*   **编译报错 "fatal error: librealsense2/rs.h: No such file or directory"**:
    *   确保 `CGO_CFLAGS` 和 `CGO_LDFLAGS` 配置正确。
    *   检查 `lib/librealsense2.so` 是否存在，且 `Makefile` 中是否设置了正确的 RPATH。
