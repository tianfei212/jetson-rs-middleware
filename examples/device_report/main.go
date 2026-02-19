package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/jojo/jetson-rs-middleware/rs"
)

func main() {
	fmt.Println("开始生成设备报告...")

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

	// 准备输出文件
	outputDir := "examples/output"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}
	reportPath := filepath.Join(outputDir, "device_report.md")
	f, err := os.Create(reportPath)
	if err != nil {
		log.Fatalf("创建报告文件失败: %v", err)
	}
	defer f.Close()

	// 辅助函数：写入 Markdown
	writeLine := func(format string, a ...interface{}) {
		fmt.Fprintf(f, format+"\n", a...)
	}

	// 开始写入报告
	writeLine("# RealSense 设备诊断报告")
	writeLine("生成时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))

	// 3. Device Information
	writeLine("## 1. 设备基本信息")
	name, _ := dev.GetInfo(rs.CameraInfoName)
	serial, _ := dev.GetInfo(rs.CameraInfoSerialNumber)
	fw, _ := dev.GetInfo(rs.CameraInfoFirmwareVersion)
	
	writeLine("- **设备名称**: %s", name)
	writeLine("- **序列号**: %s", serial)
	writeLine("- **固件版本**: %s", fw)

	usbType, err := dev.GetUSBTypeDescriptor()
	if err != nil {
		writeLine("- **USB 类型**: 错误 (%v)", err)
	} else {
		writeLine("- **USB 类型**: %s", usbType)
	}
	
	phyPort, err := dev.GetPhysicalPort()
	if err != nil {
		writeLine("- **物理端口**: 错误 (%v)", err)
	} else {
		writeLine("- **物理端口**: %s", phyPort)
	}

	// 4. Telemetry
	writeLine("\n## 2. 硬件遥测数据")
	telemetry, err := dev.GetTelemetry()
	if err != nil {
		writeLine("> 获取遥测数据失败: %v", err)
	} else {
		writeLine("| 指标 | 值 | 单位 |")
		writeLine("|---|---|---|")
		writeLine("| ASIC 温度 | %.2f | °C |", telemetry.AsicTemperature)
		writeLine("| 投影模组温度 | %.2f | °C |", telemetry.ProjectorTemperature)
	}

	// 5. Sync Status
	writeLine("\n## 3. 同步状态")
	syncMode, err := dev.GetSyncMode()
	if err != nil {
		writeLine("> 获取同步模式失败: %v", err)
	} else {
		modeStr := "未知"
		switch int(syncMode) {
		case 0:
			modeStr = "默认 (0) - 独立工作"
		case 1:
			modeStr = "主模式 (1) - Master"
		case 2:
			modeStr = "从模式 (2) - Slave"
		case 3: 
			modeStr = "完全从模式 (3) - Full Slave"
		default:
			modeStr = fmt.Sprintf("自定义 (%v)", syncMode)
		}
		writeLine("- **当前同步模式**: %s", modeStr)
	}

	// 6. Capabilities
	writeLine("\n## 4. 能力矩阵 (Stream Profiles)")
	profiles, err := dev.GetCapabilities()
	if err != nil {
		writeLine("> 获取能力矩阵失败: %v", err)
	} else {
		writeLine("共发现 **%d** 个流配置组合。", len(profiles))

		// Group by Stream Type
		profileMap := make(map[string][]rs.StreamProfile)
		for _, p := range profiles {
			profileMap[p.Stream] = append(profileMap[p.Stream], p)
		}

		// Sort keys for consistent output
		var streams []string
		for k := range profileMap {
			streams = append(streams, k)
		}
		sort.Strings(streams)

		for _, streamName := range streams {
			writeLine("\n### %s Stream", streamName)
			writeLine("| 格式 (Format) | 分辨率 (WxH) | 帧率 (FPS) | 默认配置? |")
			writeLine("|---|---|---|---|")
			
			// Sort profiles within stream: Format -> Width -> Height -> FPS
			pList := profileMap[streamName]
			sort.Slice(pList, func(i, j int) bool {
				if pList[i].Format != pList[j].Format {
					return pList[i].Format < pList[j].Format
				}
				if pList[i].Width != pList[j].Width {
					return pList[i].Width > pList[j].Width // Higher res first
				}
				if pList[i].Height != pList[j].Height {
					return pList[i].Height > pList[j].Height
				}
				return pList[i].FPS > pList[j].FPS
			})

			for _, p := range pList {
				isDefault := ""
				if p.IsDefault {
					isDefault = "✅"
				}
				writeLine("| %s | %dx%d | %d | %s |", p.Format, p.Width, p.Height, p.FPS, isDefault)
			}
		}
	}
	
	fmt.Printf("\n报告已生成: %s\n", reportPath)
}
