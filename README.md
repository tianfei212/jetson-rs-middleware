# Jetson RealSense Middleware (Go)

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8.svg)](https://golang.org/)
[![Platform](https://img.shields.io/badge/Platform-Jetson%20Orin%20(ARM64)-green.svg)](https://developer.nvidia.com/embedded/jetson-orin)

针对 Jetson Orin 平台优化的高性能 Intel RealSense D455 相机 Go 语言中间件。专为 Wails 等上层业务应用设计，提供低延迟的 CGO 封装、硬件时间戳同步、深度图对齐与增强、以及完善的设备遥测能力。

---

## ✨ 核心特性

*   **高性能 CGO 封装**: 针对 ARM64 架构优化的 `librealsense2` 绑定，最小化内存拷贝。
*   **深度/彩色对齐**: 硬件级像素对齐 (Alignment)，支持 ROI (Region of Interest) 触发逻辑。
*   **图像增强管道**: 内置 Decimation (降采样)、Spatial (空间滤波)、Temporal (时间滤波) 和 Colorizer (伪彩色) 处理器。
*   **硬件遥测监控**: 实时获取 ASIC 温度、投影模组温度、USB 连接类型及物理端口路径。
*   **多机同步支持**: 提供硬件时间戳 (Hardware Timestamp) 和同步模式查询 (Master/Slave)。
*   **能力矩阵查询**: 自动遍历并返回设备支持的所有流配置 (分辨率/帧率/格式)。
*   **HUD 数据叠加**: 支持在视频流中实时叠加时间戳、分辨率等元数据，便于调试与记录。

---

## 🛠️ 目录结构

```text
jetson-rs-middleware/
├── rs/                     # 核心驱动包 (package rs)
│   ├── context.go          # 上下文管理
│   ├── device.go           # 设备枚举与信息查询
│   ├── config.go           # 流配置
│   ├── pipeline.go         # 数据流管道
│   ├── frame.go            # 帧数据与时间戳
│   ├── align.go            # 空间对齐
│   ├── filter.go           # 图像滤波器
│   ├── colorizer.go        # 深度着色器
│   ├── sensor.go           # 传感器控制 (曝光/增益)
│   ├── telemetry.go        # 硬件遥测
│   └── capabilities.go     # 能力矩阵
├── lib/                    # 依赖库
│   └── librealsense2.so    # ARM64 动态链接库
├── examples/               # 示例代码
│   ├── device_report/      # 生成设备诊断报告 (Markdown)
│   ├── hud_video_record/   # HUD 叠加与视频录制
│   └── roi_trigger/        # ROI 触发逻辑模拟
├── cmd/                    # 命令行工具
│   ├── test-camera/        # 基础功能测试
│   └── test-new-features/  # 新特性综合测试
├── scripts/                # 辅助脚本
└── Makefile                # 构建与测试指令
```

---

## 🚀 快速开始

### 1. 环境要求
*   **硬件**: NVIDIA Jetson Orin (或兼容的 ARM64/x86_64 Linux 环境)
*   **系统**: Ubuntu 20.04 / 22.04
*   **依赖**: `libusb-1.0`, `libgtk-3-dev` (可选，用于 GUI)
*   **Go**: 1.21+ (推荐 1.26)

### 2. 安装依赖
```bash
# 安装系统依赖
sudo apt-get update && sudo apt-get install -y libusb-1.0-0-dev libglfw3-dev libgtk-3-dev

# 运行依赖安装脚本 (可选)
./scripts/install-deps.sh
```

### 3. 编译与运行测试
```bash
# 验证 D455 连接与基础功能
make test

# 测试所有新特性 (遥测、同步、能力矩阵)
go run ./cmd/test-new-features/main.go
```

### 4. 运行示例
生成设备诊断报告：
```bash
go run examples/device_report/main.go
# 查看输出: examples/output/device_report.md
```

录制带有 HUD 的视频：
```bash
go run examples/hud_video_record/main.go
# 查看输出: examples/output/output.mp4
```

---

## 📖 开发指南

详细的 API 调用与开发说明请参考 [DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md)。

---

## 📄 开源协议 (License)

Copyright 2026 Jetson RealSense Middleware Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
