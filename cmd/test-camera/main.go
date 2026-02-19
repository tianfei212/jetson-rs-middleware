package main

import (
	"fmt"
	"log"
	"time"

	"github.com/jojo/jetson-rs-middleware/rs"
)

func main() {
	fmt.Println("Starting RealSense D455 Camera Test...")

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

	// 3. 启动 Pipeline (使用默认配置)
	if err := pipeline.Start(nil); err != nil {
		log.Fatalf("Failed to start pipeline: %v", err)
	}
	defer pipeline.Stop()
	fmt.Println("Pipeline started. Capturing frames for 5 seconds...")

	// 4. 循环获取帧
	start := time.Now()
	frameCount := 0

	for time.Since(start) < 5*time.Second {
		// 等待一组帧 (超时 5000ms)
		frames, err := pipeline.WaitForFrames(5000)
		if err != nil {
			log.Printf("Error waiting for frames: %v", err)
			continue
		}

		frameCount++
		
		// 尝试获取深度帧
		depthFrame, err := frames.GetFrame(rs.StreamDepth)
		if err == nil {
			// 简单的验证：获取一下中间点的深度值
			// 注意：这里只是为了验证数据通了，生产环境需要更严谨的处理
			data := depthFrame.GetDepthData()
			if len(data) > 0 {
				centerIndex := len(data) / 2
				dist := data[centerIndex]
				if frameCount%30 == 0 { // 每30帧打印一次，避免刷屏
					fmt.Printf("Frame #%d: Center Depth = %d mm\n", frameCount, dist)
				}
			}
			depthFrame.Close()
		}

		// 尝试获取彩色帧
		colorFrame, err := frames.GetFrame(rs.StreamColor)
		if err == nil {
			// 验证彩色数据
			data := colorFrame.GetData()
			if len(data) > 0 && frameCount%30 == 0 {
				fmt.Printf("Frame #%d: Color Data Size = %d bytes\n", frameCount, len(data))
			}
			colorFrame.Close()
		}

		// 记得释放 FrameSet
		frames.Close()
	}

	fmt.Printf("Test completed. Total frames captured: %d\n", frameCount)
}
