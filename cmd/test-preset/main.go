package main

import (
	"fmt"
	"log"
	"time"

	"github.com/tianfei212/jetson-rs-middleware/rs"
)

func main() {
	// 1. 创建上下文
	ctx, err := rs.NewContext()
	if err != nil {
		log.Fatalf("Failed to create context: %v", err)
	}
	defer ctx.Close()

	// 2. 获取设备
	// 由于 Context.QueryDevices 在 Go 封装中未实现（rs/context.go 只有 NewContext 和 Close），
	// 我们改用 Pipeline 来获取设备，这在其他示例中是标准做法。
	pipeline, err := rs.NewPipeline(ctx)
	if err != nil {
		log.Fatalf("Failed to create pipeline: %v", err)
	}
	defer pipeline.Close()

	// 启动管道以获取活动设备
	if err := pipeline.Start(nil); err != nil {
		log.Fatalf("Failed to start pipeline: %v", err)
	}
	defer pipeline.Stop()

	// 等待设备就绪
	time.Sleep(1 * time.Second)

	dev, err := pipeline.GetDevice()
	if err != nil {
		log.Fatalf("Failed to get device: %v", err)
	}
	defer dev.Close()

	// 3. 获取深度传感器
	sensor, err := dev.GetDepthSensor()
	if err != nil {
		log.Fatalf("Failed to get depth sensor: %v", err)
	}
	defer sensor.Close()

	// 4. 测试 Visual Preset
	fmt.Println("Testing Visual Preset...")

	// 获取当前预设
	currentPreset, err := sensor.GetVisualPreset()
	if err != nil {
		log.Printf("Warning: Failed to get current preset: %v", err)
	} else {
		fmt.Printf("Current Preset: %d\n", currentPreset)
	}

	// 设置为 High Accuracy
	fmt.Println("Setting to High Accuracy...")
	err = sensor.SetVisualPreset(rs.VisualPresetHighAccuracy)
	if err != nil {
		log.Fatalf("Failed to set High Accuracy preset: %v", err)
	}

	// 验证设置
	newPreset, err := sensor.GetVisualPreset()
	if err != nil {
		log.Fatalf("Failed to get new preset: %v", err)
	}
	fmt.Printf("New Preset: %d\n", newPreset)

	if newPreset != rs.VisualPresetHighAccuracy {
		log.Fatalf("Preset mismatch! Expected %d, got %d", rs.VisualPresetHighAccuracy, newPreset)
	}

	// 恢复默认 (可选)
	fmt.Println("Setting back to Default...")
	err = sensor.SetVisualPreset(rs.VisualPresetDefault)
	if err != nil {
		log.Printf("Failed to set Default preset: %v", err)
	}

	fmt.Println("Visual Preset test passed!")
	
	// 简单等待一下确保命令执行完成
	time.Sleep(1 * time.Second)
}
