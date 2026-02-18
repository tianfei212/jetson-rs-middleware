这份 Skill 文件将作为你开发 **go-realsense** 中间件的蓝图。它针对 Jetson Orin 的硬件特性（ARM64、统一内存）以及 Wails 的应用场景（高性能 CGO 交互）进行了深度优化。

---

# **Skill 文件：go-realsense 中间件开发指南**

## **1\. 项目愿景**

构建一个高性能、低延迟的 Go 语言驱动层，封装 librealsense2 C API，为上层业务（如 Wails）提供对齐的视觉与深度数据流，支撑“兴趣区域（ROI）触发拍摄”功能。

## **2\. 准确的项目目录结构**

Plaintext

go-realsense/  
├── go.mod                  \# 模块定义 (github.com/jojo/go-realsense)  
├── Makefile                \# 编译指令（处理 ARM64 交叉编译与 CGO 链接）  
├── lib/                    \# 连接库
│   └── librealsense2.so    \# 动态链接库 (librealsense2)  
├── scripts/                \# 脚本工具  
│   └── install-deps.sh     \# Jetson 环境依赖安装脚本  
├── rs/                     \# 核心封装包 (package rs)  
│   ├── rs.go               \# CGO 基础配置、头文件引入、全局类型  
│   ├── errors.go           \# rs2\_error 解析与 Go error 转换  
│   ├── context.go          \# 相机上下文管理  
│   ├── device.go           \# 设备枚举与硬件信息获取  
│   ├── config.go           \# 流配置（分辨率、FPS、格式设置）  
│   ├── pipeline.go         \# 数据流管道控制（Start/Stop/Wait）  
│   ├── frame.go            \# 帧数据提取（RGB、Depth、Raw Data）  
│   ├── align.go            \# 核心算法：RGB 与 Depth 空间对齐  
│   ├── filter.go           \# 图像增强：时间/空间滤波、孔洞填充  
│   ├── colorizer.go        \# 视觉辅助：深度图转伪彩色 (Heatmap)  
│   └── sensor.go           \# 传感器级控制（Depth Scale 获取）  
└── examples/               \# 验证用例  
    └── roi\_trigger/        \# ROI 触发逻辑演示模拟

---

## **3\. 核心函数定义与接口规范**

### **A. 基础流控 (Pipeline & Config)**

**文件：** config.go, pipeline.go

* **func NewConfig() (\*Config, error)**  
  * **功能**：初始化配置容器。  
* **func (c \*Config) EnableStream(stype StreamType, w, h, fps int) error**  
  * **输入**：stype (Color/Depth), w/h (像素), fps (帧率)。  
  * **输出**：error。  
* **func (p \*Pipeline) Start(cfg \*Config) error**  
  * **功能**：启动 D455 硬件流。

### **B. 空间对齐 (Alignment) —— ROI 核心**

**文件：** align.go

* **func NewAlign(alignTo StreamType) (\*Align, error)**  
  * **输入**：RS2\_STREAM\_COLOR (将深度对齐到彩色)。  
* **func (a \*Align) Process(fs \*FrameSet) (\*FrameSet, error)**  
  * **输入**：原始 FrameSet。  
  * **输出**：对齐后的 FrameSet。  
  * **说明**：确保彩色图的像素 $(x, y)$ 与深度图的 $(x, y)$ 在物理空间上完全重合。

### **C. 数据提取 (Frame Extraction)**

**文件：** frame.go

* **func (f \*Frame) GetRawData() \[\]byte**  
  * **输出**：\[\]byte (RGB 字节流，用于 JPEG 编码或 Wails 展示)。  
* **func (f \*Frame) GetDepthData() \[\]uint16**  
  * **输出**：\[\]uint16 (原始深度值，单位：步长)。  
* **func (s \*Sensor) GetDepthScale() float32**  
  * **输出**：float32 (每个步长代表的米数，D455 通常为 0.001)。

---

## **4\. 输入输出字段详解 (Payload 定义)**

为了方便 Wails 调用，中间件应输出一个结构化的 **FramePayload**：

| 字段名 | 类型 | 说明 | 来源 |
| :---- | :---- | :---- | :---- |
| **ColorBuffer** | \[\]byte | RGB888 原始图像数据 | rs2\_get\_frame\_data |
| **DepthBuffer** | \[\]uint16 | 16-bit 原始深度数值矩阵 | rs2\_get\_frame\_data |
| **HeatmapBuffer** | \[\]byte | 经过 Colorizer 处理的伪彩色预览图 | rs2\_process\_frame |
| **DepthScale** | float32 | 深度转换比例 (如 0.001) | rs2\_get\_depth\_scale |
| **Width/Height** | int | 当前流的实际分辨率 | rs2\_get\_frame\_width |
| **Timestamp** | float64 | 硬件时间戳，用于多机同步 | rs2\_get\_frame\_timestamp |

---

## **5\. ROI 触发逻辑的功能定义 (外部实现指导)**

中间件不直接决定是否“拍摄”，但提供判定所需的字段。建议的外部逻辑：

1. **输入接口**：  
   * ROI\_Rect: {x, y, w, h} (前端传来的框选坐标)。  
   * Dist\_Threshold: float32 (触发距离，如 1.5m)。  
   * Min\_Points: int (区域内达到距离要求的最小像素点数，过滤噪点)。  
2. **判定流程**：  
   * 从 FramePayload 获取 DepthBuffer。  
   * 遍历 ROI\_Rect 范围内的像素。  
   * 计算 dist \= DepthBuffer\[idx\] \* DepthScale。  
   * 如果 dist \> 0 && dist \< Dist\_Threshold，计数值 \+1。  
3. **输出动作**：  
   * 计数值 \> Min\_Points $\\rightarrow$ 触发拍摄 (保存 ColorBuffer)。

---

## **6\. 开发注意事项 (Jetson 特别版)**

* **CGO 内存管理**：rs2\_frame 是由 C 分配的。在 Go 中必须手动调用 C.rs2\_release\_frame，否则在 Jetson Orin 上运行半小时就会因为内存泄漏导致系统 OOM 崩溃。  
* **统一内存优化**：利用 Jetson 的物理内存特性，尽量减少 C.GoBytes 的大拷贝。在做 ROI 计算时，直接通过 unsafe.Pointer 访问 C 内存块。  
* **指令集优化**：在编译时，确保开启针对 Cortex-A78AE (Orin CPU) 的优化标志。

 

