GO ?= go
BIN_DIR ?= bin

.DEFAULT_GOAL := help

.PHONY: help build build-server build-client run-server run-client fmt fmt-check vet test test-race tidy tidy-check check clean

help: ## 显示可用命令
	@awk 'BEGIN {FS = ":.*## "; printf "Usage: make <target>\n\nTargets:\n"} /^[a-zA-Z_-]+:.*## / {printf "  %-12s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: build-server build-client ## 构建服务端和客户端

build-server: ## 构建服务端到 bin/server
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/server ./cmd/server

build-client: ## 构建客户端到 bin/client
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/client ./cmd/client

run-server: ## 运行服务端
	$(GO) run ./cmd/server

run-client: ## 运行客户端
	$(GO) run ./cmd/client

fmt: ## 格式化 Go 源文件
	@find . -type f -name '*.go' -not -path './vendor/*' -exec gofmt -w {} +

fmt-check: ## 检查 Go 源文件格式
	@files="$$(find . -type f -name '*.go' -not -path './vendor/*' -exec gofmt -l {} +)"; \
	if [ -n "$$files" ]; then \
		printf '%s\n' "以下文件需要执行 gofmt:" "$$files"; \
		exit 1; \
	fi

vet: ## 运行 Go 静态检查
	@if [ -z "$$($(GO) list ./... 2>/dev/null)" ]; then \
		echo "尚无 Go 包，跳过 go vet"; \
	else \
		$(GO) vet ./...; \
	fi

test: ## 运行全部测试
	@if [ -z "$$($(GO) list ./... 2>/dev/null)" ]; then \
		echo "尚无 Go 包，跳过 go test"; \
	else \
		$(GO) test ./...; \
	fi

test-race: ## 启用竞态检测运行全部测试
	@if [ -z "$$($(GO) list ./... 2>/dev/null)" ]; then \
		echo "尚无 Go 包，跳过 go test -race"; \
	else \
		$(GO) test -race ./...; \
	fi

tidy: ## 整理 Go 模块依赖
	$(GO) mod tidy

tidy-check: ## 检查 Go 模块依赖是否整洁
	$(GO) mod tidy -diff

check: fmt-check tidy-check vet test ## 运行提交前检查

clean: ## 删除本地构建和测试产物
	$(GO) clean
	@find . -type f \( -name 'coverage.out' -o -name '*.test' \) -delete
	@rm -rf ./bin ./dist ./coverage
