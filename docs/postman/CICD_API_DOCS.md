# K8s CICD 接口文档

## 概述

本系统提供完整的 CICD 发布能力，支持多集群部署、回滚、取消等操作，并可与 Jenkins 实现闭环。

### 基础信息
| 项目 | 值 |
|------|-----|
| Base URL | `http://localhost:8080` |
| 认证方式 | JWT Bearer Token |
| Content-Type | `application/json` |

### 状态流转
```
Pending → Queued → Running → Succeeded
                          ↘ Failed
                          ↘ Canceled
                          ↘ Rollback
```

---

## 接口清单

### 1. 发布单管理

#### 1.1 创建发布单
**POST** `/api/v1/k8s/cicd/release/create`

**Request Body:**
```json
{
    "app_name": "my-app",              // 必填：应用名称
    "namespace": "default",            // 必填：命名空间
    "workload_kind": "Deployment",     // 可选：工作负载类型（默认 Deployment）
    "workload_name": "nginx",          // 必填：工作负载名称
    "container_name": "nginx",         // 必填：容器名称
    "strategy": "rolling",             // 可选：发布策略（默认 rolling）
    "timeout_sec": 300,                // 可选：超时秒数（默认 300）
    "concurrency": 3,                  // 可选：并发数（默认 3）
    "image_repo": "nginx",             // 必填：镜像仓库
    "image_tag": "1.25.3",             // 必填：镜像标签
    "image_digest": "",                // 可选：镜像摘要
    "cluster_ids": [1, 2],             // 必填：目标集群ID列表
    "request_id": "uuid-xxx"           // 可选：幂等请求ID
}
```

**Response:**
```json
{
    "code": 0,
    "msg": "success",
    "data": {
        "release_id": 1
    }
}
```

**验证点:**
- `release_id` 非空
- 重复 `request_id` 返回相同 `release_id`（幂等性）

---

#### 1.2 发布单详情
**GET** `/api/v1/k8s/cicd/release/detail`

**Query Params:**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int64 | 是 | 发布单ID |

**Response:**
```json
{
    "code": 0,
    "msg": "success",
    "data": {
        "release": {
            "id": 1,
            "app_name": "my-app",
            "namespace": "default",
            "workload_kind": "Deployment",
            "workload_name": "nginx",
            "container_name": "nginx",
            "status": "Running",
            "message": "deploying",
            "image_repo": "nginx",
            "image_tag": "1.25.3",
            "created_at": 1708934400
        },
        "tasks": [
            {
                "id": 1,
                "release_id": 1,
                "cluster_id": 1,
                "status": "Succeeded",
                "message": "部署成功",
                "prev_image": "nginx:1.24",
                "target_image": "nginx:1.25.3"
            }
        ]
    }
}
```

---

#### 1.3 发布单列表
**GET** `/api/v1/k8s/cicd/release/list`

**Query Params:**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码（默认 1）|
| page_size | int | 否 | 每页数量（默认 20）|
| app_name | string | 否 | 应用名称模糊搜索 |
| status | string | 否 | 状态筛选 |

**状态枚举:**
- `Pending` - 等待中
- `Queued` - 已入队
- `Running` - 执行中
- `Succeeded` - 成功
- `Failed` - 失败
- `Canceled` - 已取消
- `Rollback` - 已回滚

**Response:**
```json
{
    "code": 0,
    "msg": "success",
    "data": {
        "list": [...],
        "total": 100
    }
}
```

---

#### 1.4 任务列表
**GET** `/api/v1/k8s/cicd/release/tasks`

**Query Params:**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| release_id | int64 | 是 | 发布单ID |

**Response:**
```json
{
    "code": 0,
    "msg": "success",
    "data": {
        "tasks": [
            {
                "id": 1,
                "cluster_id": 1,
                "status": "Succeeded",
                "prev_image": "nginx:1.24",
                "target_image": "nginx:1.25.3",
                "started_at": 1708934401,
                "finished_at": 1708934450
            }
        ]
    }
}
```

---

### 2. 发布单操作

#### 2.1 取消发布单
**POST** `/api/v1/k8s/cicd/release/cancel`

**智能判断逻辑:**
- 状态为 `Pending`/`Queued` → 直接取消
- 状态为 `Running`/`Succeeded` → 自动触发回滚

**Request Body:**
```json
{
    "id": 1
}
```

**Response (取消):**
```json
{
    "code": 0,
    "data": {
        "message": "取消成功",
        "action": "canceled"
    }
}
```

**Response (回滚):**
```json
{
    "code": 0,
    "data": {
        "message": "发布单已触发回滚",
        "action": "rollback",
        "rollback_release_id": 2
    }
}
```

---

#### 2.2 回滚发布单
**POST** `/api/v1/k8s/cicd/release/rollback`

**前置条件:**
- 状态必须为 `Succeeded` 或 `Running`
- 必须有成功执行的任务（有 `prev_image`）

**Request Body:**
```json
{
    "id": 1
}
```

**Response:**
```json
{
    "code": 0,
    "data": {
        "message": "回滚成功",
        "rollback_release_id": 2
    }
}
```

---

#### 2.3 重试发布单
**POST** `/api/v1/k8s/cicd/release/retry`

**功能:** 基于原发布单配置创建新的发布单

**Request Body:**
```json
{
    "id": 1
}
```

**Response:**
```json
{
    "code": 0,
    "data": {
        "release_id": 3
    }
}
```

---

### 3. Jenkins 回调

#### 3.1 构建回调
**POST** `/api/v1/k8s/cicd/callback/build`

**调用时机:** Jenkins 构建完成后

**Request Body:**
```json
{
    "build_id": 123,           // Jenkins 构建号
    "status": "SUCCESS",       // SUCCESS / FAILURE / ABORTED
    "image_repo": "nginx",     // 构建出的镜像仓库
    "image_tag": "1.25.4",     // 构建出的镜像标签
    "image_digest": "sha256:abc123",  // 可选：镜像摘要
    "message": "Build completed"      // 构建消息
}
```

**Response:**
```json
{
    "code": 0,
    "data": {
        "message": "回调处理成功"
    }
}
```

**Jenkins Pipeline 示例:**
```groovy
pipeline {
    agent any
    stages {
        stage('Build') {
            steps {
                sh 'docker build -t ${IMAGE_REPO}:${IMAGE_TAG} .'
                sh 'docker push ${IMAGE_REPO}:${IMAGE_TAG}'
            }
        }
    }
    post {
        success {
            httpRequest(
                url: "${API_BASE}/api/v1/k8s/cicd/callback/build",
                httpMode: 'POST',
                contentType: 'APPLICATION_JSON',
                customHeaders: [[name: 'Authorization', value: "Bearer ${TOKEN}"]],
                requestBody: """{
                    "build_id": ${BUILD_NUMBER},
                    "status": "SUCCESS",
                    "image_repo": "${IMAGE_REPO}",
                    "image_tag": "${IMAGE_TAG}"
                }"""
            )
        }
        failure {
            httpRequest(
                url: "${API_BASE}/api/v1/k8s/cicd/callback/build",
                httpMode: 'POST',
                contentType: 'APPLICATION_JSON',
                customHeaders: [[name: 'Authorization', value: "Bearer ${TOKEN}"]],
                requestBody: """{
                    "build_id": ${BUILD_NUMBER},
                    "status": "FAILURE",
                    "message": "Build failed"
                }"""
            )
        }
    }
}
```

---

## Postman 使用指南

### 导入方式
1. 打开 Postman
2. 点击 **Import** 按钮
3. 选择文件: `docs/postman/K8s_CICD_API_Collection.json`
4. 导入完成

### 环境配置
导入后会自动创建以下变量:
| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `baseUrl` | `http://localhost:8080` | API 地址 |
| `token` | 空 | JWT Token（登录后自动填充）|
| `releaseId` | 空 | 发布单ID（创建后自动填充）|
| `clusterId` | `1` | 集群ID |

### 执行顺序
```
00-前置准备
  ├── 00-1 健康检查       → 验证服务可用
  └── 00-2 登录获取Token  → 自动保存 token

01-发布单管理
  ├── 01-1 创建发布单     → 保存 releaseId
  ├── 01-2 发布单详情
  ├── 01-3 发布单列表
  └── 01-4 获取任务列表

02-发布单操作
  ├── 02-1 取消发布单
  ├── 02-2 回滚发布单
  └── 02-3 重试发布单

03-Jenkins回调
  ├── 03-1 构建成功回调
  └── 03-2 构建失败回调
```

### 批量运行
1. 右键 Collection → **Run collection**
2. 勾选要执行的请求
3. 点击 **Run**

---

## 错误码说明

| 错误码 | 含义 | 处理建议 |
|--------|------|----------|
| 40001 | 参数校验失败 | 检查必填字段 |
| 50001 | 创建发布单失败 | 检查数据库连接 |
| 50002 | 查询发布单失败 | 检查发布单ID是否存在 |
| 50003 | 取消发布单失败 | 检查发布单状态是否允许取消 |
| 50004 | 回滚失败 | 检查是否有可回滚的任务 |
| 50005 | 重试失败 | 检查原发布单是否存在 |
| 50006 | 回调处理失败 | 检查 build_id 是否关联发布单 |

---

## 数据库依赖

确保已执行以下 SQL:
```bash
mysql -u root -p k8s-platform < docs/sql/k8s-platform.sql
```

涉及表:
- `cicd_release` - 发布单主表
- `cicd_release_task` - 发布任务表
