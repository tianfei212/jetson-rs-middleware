package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/tianfei212/jetson-rs-middleware/rs"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// Config
const (
	Width  = 640
	Height = 480
	FPS    = 30
)

func main() {
	fmt.Println("Starting HUD Video Record Example...")

	// 1. 初始化 RealSense
	ctx, err := rs.NewContext()
	if err != nil {
		log.Fatalf("Failed to create context: %v", err)
	}
	defer ctx.Close()

	pipeline, err := rs.NewPipeline(ctx)
	if err != nil {
		log.Fatalf("Failed to create pipeline: %v", err)
	}
	defer pipeline.Close()

	cfg, err := rs.NewConfig()
	if err != nil {
		log.Fatalf("Failed to create config: %v", err)
	}
	defer cfg.Close()

	// 启用 Color 和 Depth 流
	cfg.EnableStream(rs.StreamColor, Width, Height, FPS, rs.FormatRGB8)
	cfg.EnableStream(rs.StreamDepth, Width, Height, FPS, rs.FormatZ16)

	// 启动 Pipeline
	if err := pipeline.Start(cfg); err != nil {
		log.Fatalf("Failed to start pipeline: %v", err)
	}
	defer pipeline.Stop()

	// 初始化 Colorizer (深度 -> 伪彩色)
	colorizer, err := rs.NewColorizer()
	if err != nil {
		log.Fatalf("Failed to create colorizer: %v", err)
	}
	defer colorizer.Close()

	// 初始化 Align (深度对齐到彩色)
	align, err := rs.NewAlign(rs.StreamColor)
	if err != nil {
		log.Fatalf("Failed to create align: %v", err)
	}
	defer align.Close() // Assuming Align has Close method (checked previously)

	// 准备输出目录
	outputDir := "examples/output"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// 2. 准备 FFmpeg 录制
	// 录制 output.mp4, 30fps, RGB24 输入
	outputVideoPath := filepath.Join(outputDir, "output.mp4")
	ffmpegCmd := exec.Command("ffmpeg",
		"-y", // 覆盖输出文件
		"-f", "rawvideo",
		"-vcodec", "rawvideo",
		"-s", fmt.Sprintf("%dx%d", Width, Height),
		"-pix_fmt", "rgba", // 使用 RGBA 输入方便处理
		"-r", fmt.Sprintf("%d", FPS),
		"-i", "-", // 从 stdin 读取
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-pix_fmt", "yuv420p",
		outputVideoPath,
	)

	ffmpegIn, err := ffmpegCmd.StdinPipe()
	if err != nil {
		log.Fatalf("Failed to get ffmpeg stdin: %v", err)
	}

	// 启动 ffmpeg
	if err := ffmpegCmd.Start(); err != nil {
		log.Fatalf("Failed to start ffmpeg: %v", err)
	}
	defer func() {
		ffmpegIn.Close()
		ffmpegCmd.Wait()
	}()

	fmt.Println("Recording started... (4 seconds total: 2s RGB + 2s Depth)")

	// 3. 循环采集
	// 前 60 帧 (2s) 录制 RGB
	// 后 60 帧 (2s) 录制 Depth (Colorized)
	totalFrames := FPS * 4
	frameCount := 0

	// 预热几帧让自动曝光稳定
	for i := 0; i < 30; i++ {
		frames, err := pipeline.WaitForFrames(1000)
		if err == nil {
			frames.Close()
		}
	}

	for frameCount < totalFrames {
		frames, err := pipeline.WaitForFrames(1000)
		if err != nil {
			log.Printf("Error waiting for frames: %v", err)
			continue
		}

		// 对齐
		alignedFrames, err := align.Process(frames)
		frames.Close()
		if err != nil {
			log.Printf("Error aligning frames: %v", err)
			continue
		}

		var currentImg *image.RGBA
		var modeStr string
		var ts float64
		var tsDomain int

		// 前 2 秒 (60 帧) -> RGB
		if frameCount < FPS*2 {
			colorFrame, err := alignedFrames.GetFrame(rs.StreamColor)
			if err == nil {
				data := colorFrame.GetRawData()
				ts, _ = colorFrame.GetTimestamp()
				tsDomain, _ = colorFrame.GetTimestampDomain()

				// 转换为 image.RGBA
				currentImg = rgb8ToRGBA(data, Width, Height)
				modeStr = "RGB"
				colorFrame.Close()
			}
		} else {
			// 后 2 秒 (60 帧) -> Depth (Colorized)
			depthFrame, err := alignedFrames.GetFrame(rs.StreamDepth)
			if err == nil {
				// 生成伪彩色
				colorizedFrame, err := colorizer.Process(depthFrame)
				depthFrame.Close()
				if err == nil {
					data := colorizedFrame.GetRawData()
					ts, _ = colorizedFrame.GetTimestamp()
					tsDomain, _ = colorizedFrame.GetTimestampDomain()

					// Colorizer 输出通常是 RGB8
					currentImg = rgb8ToRGBA(data, Width, Height)
					modeStr = "Depth (Colorized)"
					colorizedFrame.Close()
				}
			}
		}
		alignedFrames.Close()

		if currentImg != nil {
			// 4. 叠加 HUD 信息
			drawHUD(currentImg, frameCount, ts, tsDomain, modeStr)

			// 5. 保存截图 (RGB 第一帧 和 Depth 第一帧)
			if frameCount == 0 {
				saveImage(filepath.Join(outputDir, "rgb_snapshot.png"), currentImg)
				fmt.Println("Saved rgb_snapshot.png")
			} else if frameCount == FPS*2 {
				saveImage(filepath.Join(outputDir, "depth_snapshot.png"), currentImg)
				fmt.Println("Saved depth_snapshot.png")
			}

			// 6. 写入 FFmpeg
			// image.RGBA 的 Pix 是 []uint8 (R, G, B, A)
			_, err := ffmpegIn.Write(currentImg.Pix)
			if err != nil {
				log.Printf("Error writing to ffmpeg: %v", err)
				break
			}

			frameCount++
			if frameCount%30 == 0 {
				fmt.Printf("Processed frame %d/%d\n", frameCount, totalFrames)
			}
		}
	}

	fmt.Println("Recording finished. Saved to output.mp4")
}

// rgb8ToRGBA 将 RGB8 数据转换为 image.RGBA
func rgb8ToRGBA(data []byte, w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	// 简单拷贝，注意 data 是 RGB (3 bytes)，img.Pix 是 RGBA (4 bytes)
	// 性能优化：直接操作切片
	pix := img.Pix
	// 确保 data 长度足够
	limit := w * h
	if len(data) < limit*3 {
		return img // 或者返回错误，这里简单处理返回黑图
	}

	for i := 0; i < limit; i++ {
		pix[i*4+0] = data[i*3+0] // R
		pix[i*4+1] = data[i*3+1] // G
		pix[i*4+2] = data[i*3+2] // B
		pix[i*4+3] = 255         // A
	}
	return img
}

// drawHUD 在图像上绘制 HUD 信息
func drawHUD(img *image.RGBA, frameIdx int, ts float64, domain int, mode string) {
	// 绘制半透明背景条
	bgRect := image.Rect(0, 0, Width, 80)
	// image.NewUniform 创建纯色背景
	bg := image.NewUniform(color.RGBA{0, 0, 0, 150})
	draw.Draw(img, bgRect, bg, image.Point{}, draw.Over)

	// 准备文字绘制
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.White),
		Face: basicfont.Face7x13,
	}

	lines := []string{
		fmt.Sprintf("Mode: %s", mode),
		fmt.Sprintf("Frame: %d | Time: %.2f ms (Domain: %d [1=HW, 2=Sys])", frameIdx, ts, domain),
		fmt.Sprintf("Res: %dx%d | Fmt: RGB8/Z16", Width, Height),
		fmt.Sprintf("System: %s", time.Now().Format("15:04:05.000")),
	}

	y := 20
	for _, line := range lines {
		d.Dot = fixed.Point26_6{
			X: fixed.I(10),
			Y: fixed.I(y),
		}
		d.DrawString(line)
		y += 18
	}
}

// saveImage 保存图片为 PNG
func saveImage(filename string, img image.Image) {
	f, err := os.Create(filename)
	if err != nil {
		log.Printf("Failed to create file %s: %v", filename, err)
		return
	}
	defer f.Close()
	png.Encode(f, img)
}
