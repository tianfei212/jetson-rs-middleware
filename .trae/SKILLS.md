这份 Skill 文件将作为你开发 **go-realsense** 中间件的蓝图。它针对 Jetson Orin 的硬件特性（ARM64、统一内存）以及 Wails 的应用场景（高性能 CGO 交互）进行了深度优化。

---

# **Skill 文件：go-realsense 中间件开发指南**

## **1. 项目愿景**

构建一个高性能、低延迟的 Go 语言驱动层，封装 librealsense2 C API，为上层业务（如 Wails）提供对齐的视觉与深度数据流，支撑“兴趣区域（ROI）触发拍摄”功能，并提供完善的设备诊断与遥测能力。

## **2. 准确的项目目录结构**

```text
go-realsense/  
├── go.mod                  # 模块定义 (github.com/jojo/go-realsense)  
├── Makefile                # 编译指令（处理 ARM64 交叉编译与 CGO 链接）  
├── lib/                    # 连接库
│   └── librealsense2.so    # 动态链接库 (librealsense2)  
├── scripts/                # 脚本工具  
│   └── install-deps.sh     # Jetson 环境依赖安装脚本  
├── rs/                     # 核心封装包 (package rs)  
│   ├── rs.go               # CGO 基础配置、头文件引入、全局类型  
│   ├── errors.go           # rs2_error 解析与 Go error 转换  
│   ├── context.go          # 相机上下文管理  
│   ├── device.go           # 设备枚举、USB信息、物理端口与固件信息  
│   ├── config.go           # 流配置（分辨率、FPS、格式设置）  
│   ├── pipeline.go         # 数据流管道控制（Start/Stop/Wait）  
│   ├── frame.go            # 帧数据提取（RGB、Depth、Raw Data、Hardware Timestamp）  
│   ├── align.go            # 核心算法：RGB 与 Depth 空间对齐  
│   ├── filter.go           # 图像增强：Decimation, Spatial, Temporal Filters  
│   ├── colorizer.go        # 视觉辅助：深度图转伪彩色 (Heatmap)  
│   ├── sensor.go           # 传感器级控制（曝光、增益、激光功率、ROI）  
│   ├── telemetry.go        # 硬件遥测（温度、电压、同步状态）  
│   └── capabilities.go     # 能力矩阵（流配置遍历）  
└── examples/               # 验证用例  
    ├── roi_trigger/        # ROI 触发逻辑演示模拟
    ├── hud_video_record/   # HUD叠加与视频录制（RGB+Depth）
    └── device_report/      # 设备诊断报告生成（Markdown）
```

---

## **3. 核心函数定义与接口规范**

### **A. 基础流控 (Pipeline & Config)**

**文件：** config.go, pipeline.go

* **func NewConfig() (*Config, error)**  
  * **功能**：初始化配置容器。  
* **func (c *Config) EnableStream(stype StreamType, w, h, fps int) error**  
  * **输入**：stype (Color/Depth), w/h (像素), fps (帧率)。  
  * **输出**：error。  
* **func (p *Pipeline) Start(cfg *Config) error**  
  * **功能**：启动 D455 硬件流。
* **func (p *Pipeline) WaitForFrames(timeout uint) (*FrameSet, error)**
  * **输入**：timeout (毫秒)。
  * **输出**：FrameSet (包含彩色和深度帧的集合)。
* **func (p *Pipeline) Stop()**
  * **功能**：停止相机流。

### **B. 空间对齐 (Alignment) —— ROI 核心**

**文件：** align.go

* **func NewAlign(alignTo StreamType) (*Align, error)**  
  * **输入**：RS2_STREAM_COLOR (将深度对齐到彩色)。  
* **func (a *Align) Process(fs *FrameSet) (*FrameSet, error)**  
  * **输入**：原始 FrameSet。  
  * **输出**：对齐后的 FrameSet。  
  * **说明**：确保彩色图的像素 $(x, y)$ 与深度图的 $(x, y)$ 在物理空间上完全重合。

### **C. 数据提取 (Frame Extraction)**

**文件：** frame.go

* **func (f *Frame) GetRawData() []byte**  
  * **输出**：[]byte (RGB 字节流，用于 JPEG 编码或 Wails 展示)。  
* **func (f *Frame) GetDepthData() []uint16**  
  * **输出**：[]uint16 (原始深度值，单位：步长)。  
* **func (s *Sensor) GetDepthScale() float32**  
  * **输出**：float32 (每个步长代表的米数，D455 通常为 0.001)。
* **func (f *Frame) GetTimestamp() float64**
  * **输出**：float64 (硬件时间戳，用于多机同步与高精度时序分析)。

### **D. 高级处理 (Processing Blocks)**

**文件：** colorizer.go, filter.go

* **func NewColorizer() (*Colorizer, error)**
  * **功能**：创建深度伪彩色处理器。
* **func (c *Colorizer) Process(f *Frame) (*Frame, error)**
  * **输入**：深度帧 (Z16)。
  * **输出**：彩色帧 (RGB8/BGR8)，用于可视化。
* **func NewDecimationFilter() (*Filter, error)**
  * **功能**：降采样过滤器，减少计算量。
* **func NewSpatialFilter() (*Filter, error)**
  * **功能**：空间滤波器，平滑边缘，填充孔洞。
* **func NewTemporalFilter() (*Filter, error)**
  * **功能**：时间滤波器，利用历史帧平滑深度数据，减少闪烁。

### **E. 设备诊断与遥测 (Diagnostics & Telemetry)**

**文件：** device.go, telemetry.go, capabilities.go

* **func (d *Device) GetUSBTypeDescriptor() (string, error)**
  * **输出**：USB 连接类型 (如 "3.2", "2.1")，用于判断带宽瓶颈。
* **func (d *Device) GetPhysicalPort() (string, error)**
  * **输出**：物理 USB 端口路径 (如 "2-1.3-4")，用于多机定位。
* **func (d *Device) GetTelemetry() (*TelemetryData, error)**
  * **输出**：包含 ASIC 温度、投影模组温度等硬件状态。
* **func (d *Device) GetSyncMode() (float32, error)**
  * **输出**：相机同步模式 (0=Default, 1=Master, 2=Slave)。
* **func (d *Device) GetCapabilities() ([]StreamProfile, error)**
  * **输出**：设备支持的所有流配置列表 (分辨率、帧率、格式)。

---

## **4. 输入输出字段详解 (Payload 定义)**

为了方便 Wails 调用，中间件应输出一个结构化的 **FramePayload**：

| 字段名 | 类型 | 说明 | 来源 |
| :---- | :---- | :---- | :---- |
| **ColorBuffer** | []byte | RGB888 原始图像数据 | rs2_get_frame_data |
| **DepthBuffer** | []uint16 | 16-bit 原始深度数值矩阵 | rs2_get_frame_data |
| **HeatmapBuffer** | []byte | 经过 Colorizer 处理的伪彩色预览图 | rs2_process_frame |
| **DepthScale** | float32 | 深度转换比例 (如 0.001) | rs2_get_depth_scale |
| **Width/Height** | int | 当前流的实际分辨率 | rs2_get_frame_width |
| **Timestamp** | float64 | 硬件时间戳，用于多机同步 | rs2_get_frame_timestamp |

---

## **5. ROI 触发逻辑的功能定义 (外部实现指导)**

中间件不直接决定是否“拍摄”，但提供判定所需的字段。建议的外部逻辑：

1. **输入接口**：  
   * ROI_Rect: {x, y, w, h} (前端传来的框选坐标)。  
   * Dist_Threshold: float32 (触发距离，如 1.5m)。  
   * Min_Points: int (区域内达到距离要求的最小像素点数，过滤噪点)。  
2. **判定流程**：  
   * 从 FramePayload 获取 DepthBuffer。  
   * 遍历 ROI_Rect 范围内的像素。  
   * 计算 dist = DepthBuffer[idx] * DepthScale。  
   * 如果 dist > 0 && dist < Dist_Threshold，计数值 +1。  
3. **输出动作**：  
   * 计数值 > Min_Points $\rightarrow$ 触发拍摄 (保存 ColorBuffer)。

---

## **6. 开发注意事项 (Jetson 特别版)**

* **CGO 内存管理**：rs2_frame 是由 C 分配的。在 Go 中必须手动调用 C.rs2_release_frame，否则在 Jetson Orin 上运行半小时就会因为内存泄漏导致系统 OOM 崩溃。  
* **统一内存优化**：利用 Jetson 的物理内存特性，尽量减少 C.GoBytes 的大拷贝。在做 ROI 计算时，直接通过 unsafe.Pointer 访问 C 内存块。  
* **指令集优化**：在编译时，确保开启针对 Cortex-A78AE (Orin CPU) 的优化标志。

---

## **7. 新增功能模块详解 (New Features)**

### **USB 信息查询**
*   **功能**：使用 `RS2_CAMERA_INFO_USB_TYPE_DESCRIPTOR` 查询是 USB 3.2 还是 2.1。
*   **价值**：决定了带宽上限，USB 2.1 下某些高分辨率高帧率模式不可用。
*   **物理定位**：使用 `RS2_CAMERA_INFO_PHYSICAL_PORT` 获取具体的 USB 挂载路径，方便在多 D455 接入时定位物理位置。

### **硬件遥测 (Telemetry)**
*   **功能**：获取 ASIC 温度和投影模组温度。
*   **价值**：在工业现场或封闭机箱内，监控温度防止过热降频或损坏。

### **能力矩阵 (Capabilities)**
*   **功能**：遍历 `rs2_get_stream_profiles`。
*   **输出**：JSON 友好的 Go 结构体，列出该相机支持的所有 Resolution 和 FPS 组合，便于前端动态展示可选配置。

### **同步状态检查**
*   **功能**：查询 `RS2_OPTION_INTER_CAM_SYNC_MODE`。
*   **价值**：验证多相机同步设置是否生效（Master/Slave），结合硬件时间戳计算同步误差。

### **HUD 视频录制与数据叠加**
*   **功能**：支持在 RGB 和深度图上叠加实时数据（时间戳、分辨率、模式）。
*   **录制**：支持分段录制视频（如前2秒 RGB，后2秒深度），用于故障回溯或数据采集。
