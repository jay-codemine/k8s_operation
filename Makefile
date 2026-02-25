# ====== 基本信息 ======
APP_NAME    ?= k8soperation
PKG         ?= ./cmd/k8soperation
BIN_DIR     ?= ./bin

# 根据系统自动加可执行后缀（Windows 加 .exe，其他为空）
GOOS        ?= $(shell go env GOOS)
ifeq ($(GOOS),windows)
EXE := .exe
else
EXE :=
endif
BIN_FILE    ?= $(BIN_DIR)/$(APP_NAME)$(EXE)

GO          ?= go
GOFLAGS     ?=
LDFLAGS     ?= -s -w

# ====== 运行时配置 ======
PORT        ?= 8080
GIN_MODE    ?= release

# ====== Docker / nerdctl ======
DOCKER      ?= docker                       # 切换为 nerdctl： DOCKER=nerdctl make docker-build
IMAGE       ?= $(APP_NAME):latest
DOCKERFILE  ?= build/docker/Dockerfile      # BuildKit 版：build/containerd/Dockerfile
CONTEXT     ?= .                            # 项目根作为 build context

# ====== Swagger 配置 ======
SWAG        ?= swag
SWAG_MAIN   ?= cmd/k8soperation/main.go           # 入口（main.go）路径
SWAG_OUT    ?= docs                         # 生成目录（默认 docs/）

# ====== 跨平台（Git Bash / Linux）路径与前缀处理 ======
UNAME_S     := $(shell uname -s)
PWD_POSIX   := $(shell pwd)

# 识别 Git Bash / MSYS / MINGW / CYGWIN 作为 Windows 环境
IS_WIN      :=
ifneq (,$(findstring MINGW,$(UNAME_S)))
  IS_WIN := 1
endif
ifneq (,$(findstring MSYS,$(UNAME_S)))
  IS_WIN := 1
endif
ifneq (,$(findstring CYGWIN,$(UNAME_S)))
  IS_WIN := 1
endif

# 配置目录卷挂载路径与前缀
ifeq ($(IS_WIN),1)
  # 转为 C:/... 形式（Docker Desktop 习惯用法）
  VOL_CONFIGS := $(shell cygpath -m "$(PWD_POSIX)/configs")
  VOL_DOCS    := $(shell cygpath -m "$(PWD_POSIX)/$(SWAG_OUT)")
  # 防止 Git Bash 对 -v 参数做路径自动转换
  DOCKER_RUN_PREFIX := MSYS_NO_PATHCONV=1 MSYS2_ARG_CONV_EXCL="*"
else
  VOL_CONFIGS := $(PWD_POSIX)/configs
  VOL_DOCS    := $(PWD_POSIX)/$(SWAG_OUT)
  DOCKER_RUN_PREFIX :=
endif

# ---- 统一去掉可能的尾随空格（防止重定向等语法错误）----
APP_NAME     := $(strip $(APP_NAME))
PKG          := $(strip $(PKG))
BIN_DIR      := $(strip $(BIN_DIR))
BIN_FILE     := $(strip $(BIN_FILE))
DOCKER       := $(strip $(DOCKER))
IMAGE        := $(strip $(IMAGE))
DOCKERFILE   := $(strip $(DOCKERFILE))
CONTEXT      := $(strip $(CONTEXT))
SWAG         := $(strip $(SWAG))
SWAG_MAIN    := $(strip $(SWAG_MAIN))
SWAG_OUT     := $(strip $(SWAG_OUT))
VOL_CONFIGS  := $(strip $(VOL_CONFIGS))
VOL_DOCS     := $(strip $(VOL_DOCS))

.PHONY: all build run run-local test fmt lint clean \
        swag swag-clean swagger-ui swagger-ui-stop \
        docker-build bk-build docker-run docker-logs docker-stop docker-rm docker-push \
        help

# ====== Go 基本命令 ======
all: build

# 在 build 前自动生成 swagger 文档
build: swag
	@echo ">> Building $(BIN_FILE) ($(GOOS))"
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -trimpath -ldflags "$(LDFLAGS)" -o $(BIN_FILE) $(PKG)

# 用已构建的二进制运行（接近生产）
run: build
	@echo ">> Running $(BIN_FILE)"
	APP_CONFIG="$(VOL_CONFIGS)/config.yaml" GIN_MODE=$(GIN_MODE) "$(BIN_FILE)"

# 直接 go run（开发期）——先生成 swagger
run-local: swag
	@echo ">> go run $(PKG)"
	APP_CONFIG="$(VOL_CONFIGS)/config.yaml" GIN_MODE=debug $(GO) run $(PKG)

test:
	@echo ">> Running tests"
	$(GO) test ./... -v

fmt:
	$(GO) fmt ./...

lint:
	$(GO) vet ./...

clean:
	@echo ">> Cleaning"
	@rm -rf $(BIN_DIR)

# ====== Swagger ======
swag:
	@echo ">> Generating Swagger docs"
	@command -v $(SWAG) >/dev/null 2>&1 || { \
		echo ">> swag not found, installing..."; \
		$(GO) install github.com/swaggo/swag/cmd/swag@latest; \
	}
	$(SWAG) init -g $(SWAG_MAIN) -o $(SWAG_OUT) -d ./ --parseInternal

swag-clean:
	@echo ">> Cleaning Swagger artifacts"
	@rm -f $(SWAG_OUT)/swagger.json $(SWAG_OUT)/swagger.yaml $(SWAG_OUT)/docs.go

# 用 Docker 跑官方 swagger-ui（http://localhost:8081）
swagger-ui: swag
	@echo ">> Running swagger-ui on http://localhost:8081"
	$(DOCKER_RUN_PREFIX) $(DOCKER) run -d --name $(APP_NAME)-swagger \
	  -p 8081:8080 \
	  -e SWAGGER_JSON=/spec/swagger.json \
	  -v "$(VOL_DOCS)/swagger.json:/spec/swagger.json:ro" \
	  swaggerapi/swagger-ui:latest

swagger-ui-stop:
	- $(DOCKER) rm -f $(APP_NAME)-swagger >/dev/null 2>&1 || true

# ====== Docker 镜像 ======
docker-build: swag
	@echo ">> Building image $(IMAGE) with $(DOCKER) using $(DOCKERFILE)"
	$(DOCKER) build -f $(DOCKERFILE) -t $(IMAGE) $(CONTEXT)

# 使用 BuildKit 版 Dockerfile（更快依赖/构建缓存）
bk-build: swag
	@echo ">> Building (BuildKit) image $(IMAGE) with $(DOCKER)"
	DOCKER_BUILDKIT=1 $(DOCKER) build -f build/containerd/Dockerfile -t $(IMAGE) $(CONTEXT)

docker-run:
	@echo ">> Running container $(APP_NAME)  (configs: $(VOL_CONFIGS))"
	$(DOCKER_RUN_PREFIX) $(DOCKER) run -d --name $(APP_NAME) \
		-p $(PORT):8080 \
		-v "$(VOL_CONFIGS):/app/configs:ro" \
		-e APP_CONFIG=/app/configs/config.yaml \
		-e GIN_MODE=$(GIN_MODE) \
		--restart=always \
		$(IMAGE)

docker-logs:
	$(DOCKER) logs -f $(APP_NAME)

docker-stop:
	-$(DOCKER) stop $(APP_NAME) || true

docker-rm: docker-stop
	-$(DOCKER) rm $(APP_NAME) || true

docker-push:
	@test "$(REGISTRY)" != "" || (echo "REGISTRY not set, e.g. REGISTRY=registry.example.com/ns"; exit 1)
	$(DOCKER) tag $(IMAGE) $(REGISTRY)/$(IMAGE)
	$(DOCKER) push $(REGISTRY)/$(IMAGE)

help:
	@echo "Targets:"
	@echo "  build / run / run-local / test / fmt / lint / clean"
	@echo "  swag / swag-clean / swagger-ui / swagger-ui-stop"
	@echo "  docker-build / bk-build / docker-run / docker-logs / docker-stop / docker-rm / docker-push"
	@echo ""
	@echo "Hints:"
	@echo "  * build / run / docker-build 会自动先生成 Swagger 文档"
	@echo "  * swagger-ui 在 8081 端口起官方 UI（Windows Git Bash 路径已处理）"
	@echo "  * 若未安装 swag，会自动 go install github.com/swaggo/swag/cmd/swag@latest"
	@echo "  DOCKER=nerdctl make docker-build   # 使用 nerdctl"
