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

# ====== Docker / nerdctl（自动检测） ======
# 优先使用 docker，不可用时回退到 nerdctl
DOCKER ?= $(shell command -v docker >/dev/null 2>&1 && echo docker || (command -v nerdctl >/dev/null 2>&1 && echo nerdctl || echo docker))
IMAGE_BE ?= $(APP_NAME)-be:latest           # 后端镜像
IMAGE_FE ?= $(APP_NAME)-fe:latest           # 前端镜像
DOCKERFILE_BE ?= docker/backend/Dockerfile  # 后端 Dockerfile
DOCKERFILE_FE ?= docker/frontend/Dockerfile # 前端 Dockerfile
CONTEXT     ?= .

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

.PHONY: all build run run-quick run-local test fmt lint clean \
        swag swag-clean swagger-ui swagger-ui-stop \
        docker-build-be docker-build-fe docker-build-all docker-push-be docker-push-fe \
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

# 每次全新编译运行（强制 -a 重编所有包，适合验证 DDD 改动）
run-new:
	@echo ">> Force rebuild & run $(BIN_FILE)"
	@mkdir -p $(BIN_DIR)
	$(GO) build -a -trimpath -ldflags "$(LDFLAGS)" -o $(BIN_FILE) $(PKG)
	@echo ">> Running $(BIN_FILE)"
	APP_CONFIG="configs/config.yaml" GIN_MODE=$(GIN_MODE) "$(BIN_FILE)"

# 快速启动（跳过 swagger 重新生成，适合日常开发）
run-quick:
	@echo ">> Building $(BIN_FILE) ($(GOOS)) [skip swag]"
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -trimpath -ldflags "$(LDFLAGS)" -o $(BIN_FILE) $(PKG)
	@echo ">> Running $(BIN_FILE)"
	APP_CONFIG="configs/config.yaml" GIN_MODE=$(GIN_MODE) "$(BIN_FILE)"

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
	$(SWAG) init -g $(SWAG_MAIN) -o $(SWAG_OUT) -d ./ --parseInternal --parseDependency --parseDepth 3

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

# ====== 镜像构建（Docker / nerdctl 通用） ======
# 使用方式：
#   make docker-build-be             使用自动检测的运行时构建后端镜像
#   make docker-build-fe             构建前端镜像
#   make DOCKER=nerdctl docker-build-be  强制使用 nerdctl
#   make docker-push REGISTRY=harbor.example.com/k8s

# 构建后端镜像（多阶段，无需本地 Go 环境）
docker-build-be: swag
	@echo ">> Building backend image $(IMAGE_BE) with $(DOCKER)"
	$(DOCKER) build -f $(DOCKERFILE_BE) -t $(IMAGE_BE) $(CONTEXT)

# 构建前端镜像（多阶段，Node build → Nginx serve）
docker-build-fe:
	@echo ">> Building frontend image $(IMAGE_FE) with $(DOCKER)"
	$(DOCKER) build -f $(DOCKERFILE_FE) -t $(IMAGE_FE) k8s-web/

# 同时构建前后端
docker-build-all: docker-build-be docker-build-fe

# 多架构构建（amd64 + arm64）
docker-buildx-be:
	@echo ">> Building multi-arch backend $(IMAGE_BE)"
	$(DOCKER) buildx build --platform linux/amd64,linux/arm64 \
		-f $(DOCKERFILE_BE) -t $(IMAGE_BE) $(CONTEXT) --push

docker-buildx-fe:
	@echo ">> Building multi-arch frontend $(IMAGE_FE)"
	$(DOCKER) buildx build --platform linux/amd64,linux/arm64 \
		-f $(DOCKERFILE_FE) -t $(IMAGE_FE) k8s-web/ --push

# 运行后端容器
docker-run-be:
	@echo ">> Running backend $(IMAGE_BE)"
	$(DOCKER_RUN_PREFIX) $(DOCKER) run -d --name $(APP_NAME)-be \
		-p $(PORT):8080 \
		-v "$(VOL_CONFIGS):/app/configs:ro" \
		-e APP_CONFIG=/app/configs/config.yaml \
		-e GIN_MODE=$(GIN_MODE) \
		--restart=always \
		$(IMAGE_BE)

# 运行前端容器
docker-run-fe:
	@echo ">> Running frontend $(IMAGE_FE)"
	$(DOCKER_RUN_PREFIX) $(DOCKER) run -d --name $(APP_NAME)-fe \
		-p 80:80 \
		-e API_BACKEND_URL=http://$(APP_NAME)-be:8080 \
		--restart=always \
		$(IMAGE_FE)

# nerdctl 别名（与 docker 完全相同，只是显式指定 runtime）
nerdctl-build-be: DOCKER=nerdctl
nerdctl-build-be: docker-build-be
nerdctl-build-fe: DOCKER=nerdctl
nerdctl-build-fe: docker-build-fe
nerdctl-build-all: DOCKER=nerdctl
nerdctl-build-all: docker-build-all
nerdctl-run-be: DOCKER=nerdctl
nerdctl-run-be: docker-run-be
nerdctl-run-fe: DOCKER=nerdctl
nerdctl-run-fe: docker-run-fe

# 日志
docker-logs-be:
	$(DOCKER) logs -f $(APP_NAME)-be
docker-logs-fe:
	$(DOCKER) logs -f $(APP_NAME)-fe

# 停止/删除
docker-stop-be:
	-$(DOCKER) stop $(APP_NAME)-be || true
docker-stop-fe:
	-$(DOCKER) stop $(APP_NAME)-fe || true
docker-rm-be: docker-stop-be
	-$(DOCKER) rm $(APP_NAME)-be || true
docker-rm-fe: docker-stop-fe
	-$(DOCKER) rm $(APP_NAME)-fe || true

# 推送（Docker Registry）
docker-push-be:
	@test "$(REGISTRY)" != "" || (echo "REGISTRY not set, e.g. REGISTRY=harbor.example.com/k8s"; exit 1)
	$(DOCKER) tag $(IMAGE_BE) $(REGISTRY)/$(IMAGE_BE)
	$(DOCKER) push $(REGISTRY)/$(IMAGE_BE)
docker-push-fe:
	@test "$(REGISTRY)" != "" || (echo "REGISTRY not set, e.g. REGISTRY=harbor.example.com/k8s"; exit 1)
	$(DOCKER) tag $(IMAGE_FE) $(REGISTRY)/$(IMAGE_FE)
	$(DOCKER) push $(REGISTRY)/$(IMAGE_FE)

help:
	@echo "  build / run / run-local / test / fmt / lint / clean"
	@echo "  swag / swag-clean / swagger-ui / swagger-ui-stop"
	@echo "  docker-build-be / docker-build-fe / docker-build-all / docker-run-be / docker-run-fe / docker-push-be / docker-push-fe"
	@echo ""
	@echo "Hints:"
	@echo "  * docker-build           平台编译模式（先 go build 再 docker build，纯运行时镜像 < 20MB）"
	@echo "  * docker-build-standalone 多阶段构建（Docker 内部编译，无需本地 Go 环境）"
	@echo "  * docker-buildx          多架构构建 (amd64 + arm64)，需 docker buildx + push"
	@echo "  * swagger-ui             在 8081 端口起官方 UI（Windows Git Bash 路径已处理）"
	@echo "  * 若未安装 swag，会自动 go install github.com/swaggo/swag/cmd/swag@latest"
	@echo "  DOCKER=nerdctl-build-be / nerdctl-build-fe / nerdctl-build-all   # 使用 nerdctl"
