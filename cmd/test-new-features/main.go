package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jojo/jetson-rs-middleware/rs"
)

func main() {
	fmt.Println("开始测试新功能...")

	// 1. Initialize RealSense
	ctx, err := rs.NewContext()
	if err != nil {
		log.Fatalf("创建上下文失败: %v", err)
	}
	defer ctx.Close()

	pipeline, err := rs.NewPipeline(ctx)
	if err != nil {
		log.Fatalf("创建管道失败: %v", err)
	}
	defer pipeline.Close()

	// Start pipeline with default config to get the device
	// We need an active profile to get the device handle
	fmt.Println("正在启动管道...")
	if err := pipeline.Start(nil); err != nil {
		log.Fatalf("启动管道失败: %v", err)
	}
	defer pipeline.Stop()

	// Allow some time for the device to settle
	time.Sleep(1 * time.Second)

	// 2. Get Device
	dev, err := pipeline.GetDevice()
	if err != nil {
		log.Fatalf("获取设备失败: %v", err)
	}
	defer dev.Close()

	// 3. Test USB Info & Physical Port
	fmt.Println("\n=== 设备信息 ===")
	name, _ := dev.GetInfo(rs.CameraInfoName)
	serial, _ := dev.GetInfo(rs.CameraInfoSerialNumber)
	fw, _ := dev.GetInfo(rs.CameraInfoFirmwareVersion)
	usbType, err := dev.GetUSBTypeDescriptor()
	if err != nil {
		fmt.Printf("USB 类型: 错误 (%v)\n", err)
	} else {
		fmt.Printf("USB 类型: %s\n", usbType)
	}

	phyPort, err := dev.GetPhysicalPort()
	if err != nil {
		fmt.Printf("物理端口: 错误 (%v)\n", err)
	} else {
		fmt.Printf("物理端口: %s\n", phyPort)
	}

	fmt.Printf("设备名称: %s\n", name)
	fmt.Printf("序列号: %s\n", serial)
	fmt.Printf("固件版本: %s\n", fw)

	// 4. Test Telemetry
	fmt.Println("\n=== 遥测数据 ===")
	telemetry, err := dev.GetTelemetry()
	if err != nil {
		fmt.Printf("获取遥测数据失败: %v\n", err)
	} else {
		fmt.Printf("ASIC 温度: %.2f C\n", telemetry.AsicTemperature)
		fmt.Printf("投影模组温度: %.2f C\n", telemetry.ProjectorTemperature)
	}

	// 5. Test Sync Status
	fmt.Println("\n=== 同步状态 ===")
	syncMode, err := dev.GetSyncMode()
	if err != nil {
		fmt.Printf("获取同步模式失败: %v\n", err)
	} else {
		modeStr := "未知"
		switch int(syncMode) {
		case 0:
			modeStr = "默认 (0)"
		case 1:
			modeStr = "主模式 (1)"
		case 2:
			modeStr = "从模式 (2)"
		case 3:
			modeStr = "完全从模式 (3)"
		default:
			modeStr = fmt.Sprintf("自定义 (%v)", syncMode)
		}
		fmt.Printf("同步模式: %s\n", modeStr)
	}

	// 6. Test Capabilities
	fmt.Println("\n=== 能力矩阵 (流配置) ===")
	profiles, err := dev.GetCapabilities()
	if err != nil {
		fmt.Printf("获取能力矩阵失败: %v\n", err)
	} else {
		fmt.Printf("共发现 %d 个流配置\n", len(profiles))

		// Print summary of profiles (grouped by Stream type)
		summary := make(map[string]int)
		for _, p := range profiles {
			summary[p.Stream]++
		}

		fmt.Println("各流类型配置统计:")
		for stream, count := range summary {
			fmt.Printf("  - %s: %d 个配置\n", stream, count)
		}

		// Print first 5 profiles as sample JSON
		fmt.Println("\n配置示例 JSON):")
		sampleCount := len(profiles)
		if len(profiles) < sampleCount {
			sampleCount = len(profiles)
		}

		jsonData, _ := json.MarshalIndent(profiles[:sampleCount], "", "  ")
		fmt.Println(string(jsonData))
	}

	fmt.Println("\n测试成功完成。")
}
