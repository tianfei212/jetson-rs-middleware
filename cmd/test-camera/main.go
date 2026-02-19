package main

import (
	"fmt"
	"log"
	"time"

	"github.com/jojo/jetson-rs-middleware/rs"
)

// ROI 触发逻辑模拟
// 假设 ROI 是中心 100x100 的区域
func checkROITrigger(depthData []uint16, width, height int, scale float32) bool {
	roiW, roiH := 100, 100
	startX := (width - roiW) / 2
	startY := (height - roiH) / 2

	validPoints := 0
	minPoints := 500              // 阈值点数
	distThreshold := float32(1.5) // 1.5米

	for y := startY; y < startY+roiH; y++ {
		for x := startX; x < startX+roiW; x++ {
			idx := y*width + x
			if idx < len(depthData) {
				dist := float32(depthData[idx]) * scale
				if dist > 0.1 && dist < distThreshold {
					validPoints++
				}
			}
		}
	}

	return validPoints > minPoints
}

func main() {
	fmt.Println("Starting RealSense D455 Camera Comprehensive Test...")

	// 1. 创建上下文
	ctx, err := rs.NewContext()
	if err != nil {
		log.Fatalf("Failed to create context: %v", err)
	}
	defer ctx.Close()
	fmt.Println("Context created.")

	// 2. 创建 Pipeline
	pipeline, err := rs.NewPipeline(ctx)
	if err != nil {
		log.Fatalf("Failed to create pipeline: %v", err)
	}
	defer pipeline.Close()
	fmt.Println("Pipeline created.")

	// 3. 配置流
	cfg, err := rs.NewConfig()
	if err != nil {
		log.Fatalf("Failed to create config: %v", err)
	}
	defer cfg.Close()

	// 配置深度流 (640x480 @ 30fps, Z16)
	if err := cfg.EnableStream(rs.StreamDepth, 640, 480, 30, rs.FormatZ16); err != nil {
		log.Fatalf("Failed to enable depth stream: %v", err)
	}
	// 配置彩色流 (640x480 @ 30fps, RGB8)
	if err := cfg.EnableStream(rs.StreamColor, 640, 480, 30, rs.FormatRGB8); err != nil {
		log.Fatalf("Failed to enable color stream: %v", err)
	}
	fmt.Println("Streams configured.")

	// 4. 启动 Pipeline
	if err := pipeline.Start(cfg); err != nil {
		log.Fatalf("Failed to start pipeline: %v", err)
	}
	defer pipeline.Stop()
	fmt.Println("Pipeline started.")

	// 5. 传感器控制测试 (Sensor Control)
	// 获取设备和传感器
	dev, err := pipeline.GetDevice()
	if err != nil {
		log.Printf("Warning: Failed to get device: %v", err)
	} else {
		defer dev.Close()
		sensor, err := dev.GetDepthSensor()
		if err != nil {
			log.Printf("Warning: Failed to get depth sensor: %v", err)
		} else {
			defer sensor.Close()
			// 获取深度比例
			scale, err := sensor.GetDepthScale()
			if err == nil {
				fmt.Printf("Depth Scale: %.5f\n", scale)
			}

			// 尝试设置手动曝光 (仅测试接口调用，不保证实际效果)
			// 注意：通常需要先关闭自动曝光
			if err := sensor.SetOption(rs.OptionEnableAutoExposure, 0); err == nil {
				fmt.Println("Auto Exposure disabled.")
				// 设置曝光时间 (例如 8000us = 8ms)
				if err := sensor.SetOption(rs.OptionExposure, 8000); err == nil {
					fmt.Println("Manual Exposure set to 8000us.")
				}
			} else {
				log.Printf("Failed to disable auto exposure: %v", err)
			}
		}
	}

	// 6. 初始化过滤器 (Filters)
	decimation, _ := rs.NewDecimationFilter()
	defer decimation.Close()
	spatial, _ := rs.NewSpatialFilter()
	defer spatial.Close()
	temporal, _ := rs.NewTemporalFilter()
	defer temporal.Close()
	holeFilling, _ := rs.NewHoleFillingFilter()
	defer holeFilling.Close()

	// 初始化 Colorizer (深度伪彩色)
	colorizer, _ := rs.NewColorizer()
	defer colorizer.Close()

	// 初始化对齐 (Align Depth to Color)
	align, err := rs.NewAlign(rs.StreamColor)
	if err != nil {
		log.Fatalf("Failed to create align: %v", err)
	}
	// Align 也是一个 processing block，需要释放
	// 注意：rs.Align 结构体没有 Close 方法? 检查 rs/align.go
	// rs/align.go 中的 Align 有 ptr 和 queue，需要手动释放吗？
	// Align 结构体本身应该有 Close 方法。如果没有，我们应该添加。
	// 这里假设有 Close (如果没有，稍后添加)
	// defer align.Close() // TODO: Check if Align has Close method

	fmt.Println("Filters and Colorizer initialized.")

	// 7. 循环获取帧
	start := time.Now()
	frameCount := 0

	fmt.Println("Capturing frames for 10 seconds...")

	// 获取一次 scale 用于 ROI 计算
	var depthScale float32 = 0.001 // 默认值
	if dev != nil {
		s, err := dev.GetDepthSensor()
		if err == nil {
			depthScale, _ = s.GetDepthScale()
			s.Close()
		}
	}

	for time.Since(start) < 10*time.Second {
		// 等待一组帧
		frames, err := pipeline.WaitForFrames(5000)
		if err != nil {
			log.Printf("Error waiting for frames: %v", err)
			continue
		}

		// 8. 对齐 (Align)
		alignedFrames, err := align.Process(frames)
		frames.Close() // 原始帧集释放
		if err != nil {
			log.Printf("Error aligning frames: %v", err)
			continue
		}

		// 9. 获取深度帧并进行滤波处理
		depthFrame, err := alignedFrames.GetFrame(rs.StreamDepth)
		if err == nil {
			// 链式滤波处理
			// Decimation -> Spatial -> Temporal -> HoleFilling
			f1, _ := decimation.Process(depthFrame)
			depthFrame.Close() // 释放原始深度帧

			f2, _ := spatial.Process(f1)
			f1.Close()

			f3, _ := temporal.Process(f2)
			f2.Close()

			finalDepth, _ := holeFilling.Process(f3)
			f3.Close()

			// 10. 生成伪彩色深度图 (用于预览)
			heatmapFrame, _ := colorizer.Process(finalDepth)

			// 获取数据
			depthData := finalDepth.GetDepthData()
			heatmapData := heatmapFrame.GetRawData() // 重命名后的方法

			// 获取时间戳 (Hardware Timestamp)
			ts, _ := finalDepth.GetTimestamp()
			domain, _ := finalDepth.GetTimestampDomain()

			// 11. 模拟 ROI 触发
			triggered := checkROITrigger(depthData, 640, 480, depthScale)

			if frameCount%30 == 0 {
				fmt.Printf("Frame #%d: TS=%.2f (Domain: %d) | Depth Size: %d | Heatmap Size: %d | Trigger: %v\n",
					frameCount, ts, domain, len(depthData), len(heatmapData), triggered)
			}

			finalDepth.Close()
			heatmapFrame.Close()
		}

		// 12. 获取彩色帧
		colorFrame, err := alignedFrames.GetFrame(rs.StreamColor)
		if err == nil {
			// colorData := colorFrame.GetRawData()
			// ... 处理彩色数据 ...
			colorFrame.Close()
		}

		alignedFrames.Close()
		frameCount++
	}

	fmt.Printf("Test completed. Total frames captured: %d\n", frameCount)
}
