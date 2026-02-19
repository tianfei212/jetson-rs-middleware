# 项目基本信息
BINARY_NAME=jetson-rs-middleware
MODULE_NAME=github.com/jojo/go-realsense

# 编译参数
GO=go
CGO_ENABLED=1
# 针对 Jetson Orin 的优化：如果有特定库路径可以添加在此处
# PKG_CONFIG_PATH=/usr/local/lib/pkgconfig

.PHONY: all build clean test help

all: build

## build: 编译项目
build:
	@echo "正在为 Jetson Orin (ARM64) 编译项目..."
	$(GO) build -v -o bin/$(BINARY_NAME) ./cmd/test-camera
	@echo "编译完成！二进制文件位于 bin/$(BINARY_NAME)"

## rs-pkg: 仅编译 rs 核心包（用于语法检查）
rs-pkg:
	@echo "正在检查 rs 包编译..."
	$(GO) build -v ./rs/...

## clean: 清理编译产物
clean:
	@echo "清理中..."
	rm -rf bin/
	$(GO) clean
	@echo "清理完毕。"

## test: 运行单元测试
test:
	@echo "运行测试..."
	$(GO) test -v ./...

## deps: 安装必要的系统依赖 (Ubuntu/Jetson)
deps:
	@echo "正在检查系统依赖..."
	@sudo apt-get update && sudo apt-get install -y librealsense2-dev pkg-config

## help: 显示帮助信息
help:
	@echo "JOJO, 这里的 Makefile 可用指令如下:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' |  sed -e 's/^/ /'