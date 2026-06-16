# RAG 知识库扩展设计文档

> 版本: v1.0 | 状态: 设计中 | 更新时间: 2026-04-13

## 一、背景与目标

### 1.1 为什么需要 RAG

当前 AI 助手依赖大模型的**通用知识**，存在以下局限：
- 不了解公司内部运维手册和故障处理流程
- 无法回答特定业务架构相关问题
- K8s 最佳实践知识可能过时
- 无法基于历史故障案例给出精准建议

### 1.2 RAG 解决的问题

| 问题 | 当前 | RAG 增强后 |
|------|------|-----------|
| "我们的 nginx 配置规范是什么？" | 回答通用规范 | 检索公司 nginx 配置手册精准回答 |
| "上次数据库连接池溢出怎么处理的？" | 无法回答 | 检索历史故障案例回答 |
| "K8s 1.29 有什么 Breaking Change？" | 可能过时 | 检索最新文档回答 |
| "生产环境部署审批流程是什么？" | 不知道 | 检索内部流程文档回答 |

### 1.3 目标

1. 支持上传运维文档（Markdown/PDF/TXT）自动索引
2. AI 对话自动检索相关知识片段增强回答
3. 支持知识库分类管理（K8s/运维/故障案例/业务文档）
4. 支持手动和自动索引（集群事件/告警 自动入库）

## 二、RAG 架构设计

```mermaid
graph TB
    subgraph Ingestion["知识入库流程"]
        Upload[文档上传]
        AutoIndex[自动索引]
        Splitter[文本分片 Chunking]
        Embedder[向量化 Embedding]
        Store[存储]
    end

    subgraph Retrieval["检索增强流程"]
        UserQuery[用户提问]
        QueryEmbed[问题向量化]
        VectorSearch[向量相似搜索]
        Rerank[重排序 可选]
        ContextBuild[上下文构建]
        LLMCall[LLM 增强回答]
    end

    subgraph Storage["存储层"]
        MySQL[(MySQL 元数据)]
        VectorDB[(向量数据库)]
    end

    Upload --> Splitter
    AutoIndex --> Splitter
    Splitter --> Embedder
    Embedder --> Store
    Store --> MySQL
    Store --> VectorDB

    UserQuery --> QueryEmbed
    QueryEmbed --> VectorSearch
    VectorSearch --> VectorDB
    VectorSearch --> Rerank
    Rerank --> ContextBuild
    ContextBuild --> LLMCall
```

## 三、向量数据库选型

### 3.1 方案对比

| 方案 | 优点 | 缺点 | 适用场景 |
|------|------|------|---------|
| **pgvector (PostgreSQL)** | 零额外部署、SQL 友好、运维简单 | 性能中等、百万级以上吃力 | 中小规模推荐 |
| **Milvus** | 高性能、分布式、功能全 | 部署复杂、资源占用大 | 大规模生产 |
| **ChromaDB** | 超轻量、嵌入式、API 简单 | 单机、无高可用 | 开发/POC |
| **Weaviate** | GraphQL API、混合搜索 | 部署复杂 | 混合搜索场景 |
| **Qdrant** | Rust 高性能、API 简洁 | 社区较小 | 高性能需求 |

### 3.2 推荐方案

**初期**：pgvector（复用 MySQL 替换为 PostgreSQL，或独立一个 PG 实例）
**或者**：ChromaDB（零部署，嵌入 Go 进程，适合 POC 快速验证）
**生产**：Milvus（性能好，Kubernetes 原生部署）

### 3.3 替代方案：MySQL 全文检索 + BM25

如果暂不想引入向量数据库，可以先用 MySQL 全文索引 + BM25 关键词检索做简易版：

```sql
ALTER TABLE knowledge_chunks ADD FULLTEXT INDEX ft_content (content) WITH PARSER ngram;
SELECT *, MATCH(content) AGAINST('nginx 配置' IN NATURAL LANGUAGE MODE) AS score 
FROM knowledge_chunks WHERE MATCH(content) AGAINST('nginx 配置') ORDER BY score DESC LIMIT 5;
```

## 四、Embedding 模型选型

| 模型 | 提供商 | 维度 | 特点 | 推荐 |
|------|--------|------|------|------|
| text-embedding-v3 | 通义千问 | 1024 | 中文优化、API 稳定 | 推荐 |
| text-embedding-3-small | OpenAI | 1536 | 效果好、成本低 | 备选 |
| bge-large-zh | BAAI | 1024 | 开源、中文最佳 | 自部署推荐 |
| embedding-3 | 智谱 | 2048 | 中文优化 | 备选 |

> 推荐直接复用已有 AI 提供商配置，在 `config.yaml` 的 Providers 中新增 Embedding 模型即可。

## 五、配置设计

```yaml
# config.yaml 新增段
RAG:
  Enabled: true
  
  # 向量数据库配置
  VectorDB:
    Type: "milvus"                 # milvus / pgvector / chroma / mysql_fulltext
    # Milvus 配置
    Milvus:
      Address: "localhost:19530"
      CollectionName: "k8s_knowledge"
      Dimension: 1024              # 向量维度（与 Embedding 模型一致）
    # pgvector 配置
    PGVector:
      DSN: "postgres://user:pass@localhost:5432/knowledge?sslmode=disable"
    # ChromaDB 配置
    Chroma:
      Address: "http://localhost:8000"
      CollectionName: "k8s_knowledge"
  
  # Embedding 配置（复用 AI 提供商）
  Embedding:
    ProviderID: "qwen"             # 复用已配置的 AI 提供商
    Model: "text-embedding-v3"     # Embedding 模型名
    BatchSize: 32                  # 批量 Embedding 大小
    
  # 文本分片配置
  Chunking:
    ChunkSize: 512                 # 分片大小（token 数）
    ChunkOverlap: 64               # 分片重叠（token 数）
    Separators:                    # 分隔符优先级
      - "\n## "                    # Markdown H2
      - "\n### "                   # Markdown H3
      - "\n\n"                     # 段落
      - "\n"                       # 换行
      - ". "                       # 句号
  
  # 检索配置
  Retrieval:
    TopK: 5                        # 返回最相关的 K 个片段
    ScoreThreshold: 0.7            # 最低相似度阈值
    MaxContextTokens: 2000         # 注入 LLM 的最大上下文 token
    EnableRerank: false            # 是否启用重排序
    
  # 自动索引（定期抓取集群信息入库）
  AutoIndex:
    Enabled: false
    Interval: 3600                 # 自动索引间隔（秒）
    Sources:
      - "cluster_events"           # 集群事件
      - "alert_history"            # 告警历史
      - "deployment_changes"       # 部署变更记录
```

## 六、后端模块设计

### 6.1 新增文件结构

```
pkg/
  ├── vectordb/                    # 向量数据库抽象层（可插拔）
  │   ├── vectordb.go              # 接口定义
  │   ├── milvus.go                # Milvus 实现
  │   ├── pgvector.go              # pgvector 实现
  │   ├── chroma.go                # ChromaDB 实现
  │   └── mysql_fulltext.go        # MySQL 全文检索实现
  │
  └── embedding/                   # Embedding 客户端
      └── embedding.go             # 调用 AI 提供商 Embedding API

internal/app/
  ├── services/
  │   ├── rag_knowledge.go         # 知识库管理 Service
  │   │   - UploadDocument()       # 上传文档
  │   │   - DeleteDocument()       # 删除文档及其分片
  │   │   - ListDocuments()        # 文档列表
  │   │   - ReindexDocument()      # 重新索引
  │   │   - ProcessDocument()      # 分片 + Embedding + 入库
  │   │
  │   └── rag_retrieval.go         # RAG 检索 Service
  │       - Retrieve(query, topK)  # 向量检索
  │       - BuildRAGContext()      # 构建增强上下文
  │       - EnhancePrompt()        # 增强 System Prompt
  │
  ├── models/
  │   ├── knowledge_doc.go         # 知识文档模型
  │   └── knowledge_chunk.go       # 知识分片模型
  │
  ├── dao/
  │   └── knowledge_dao.go         # 知识库 DAO
  │
  └── routers/knowledge/
      └── router.go                # 知识库 API 路由
```

### 6.2 核心接口定义

```go
// pkg/vectordb/vectordb.go

// VectorDB 向量数据库接口（可插拔设计）
type VectorDB interface {
    // Insert 插入向量
    Insert(ctx context.Context, docs []Document) error
    // Search 相似搜索
    Search(ctx context.Context, vector []float32, topK int, filter map[string]string) ([]SearchResult, error)
    // Delete 按 ID 删除
    Delete(ctx context.Context, ids []string) error
    // Close 关闭连接
    Close() error
}

// Document 待入库文档
type Document struct {
    ID       string            // 唯一标识
    Content  string            // 文本内容
    Vector   []float32         // Embedding 向量
    Metadata map[string]string // 元数据（来源/分类/文档ID 等）
}

// SearchResult 搜索结果
type SearchResult struct {
    ID       string
    Content  string
    Score    float32           // 相似度分数
    Metadata map[string]string
}
```

```go
// pkg/embedding/embedding.go

// Embedder Embedding 接口
type Embedder interface {
    // Embed 单条文本 → 向量
    Embed(ctx context.Context, text string) ([]float32, error)
    // EmbedBatch 批量文本 → 向量
    EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
}
```

### 6.3 数据库表设计

```sql
-- 知识文档（元数据）
CREATE TABLE knowledge_documents (
    id           BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    title        VARCHAR(256)  NOT NULL COMMENT '文档标题',
    category     VARCHAR(64)   NOT NULL DEFAULT 'general' COMMENT '分类: k8s/ops/troubleshoot/business',
    source_type  VARCHAR(32)   NOT NULL DEFAULT 'upload' COMMENT '来源: upload/auto_index/manual',
    file_type    VARCHAR(16)   NOT NULL DEFAULT 'md' COMMENT '文件类型: md/pdf/txt',
    file_size    INT UNSIGNED  DEFAULT 0 COMMENT '文件大小(bytes)',
    chunk_count  INT UNSIGNED  DEFAULT 0 COMMENT '分片数量',
    status       VARCHAR(16)   NOT NULL DEFAULT 'pending' COMMENT 'pending/indexing/ready/error',
    error_msg    TEXT           COMMENT '错误信息',
    tags         JSON           COMMENT '标签: ["nginx","配置"]',
    created_by   INT UNSIGNED  NOT NULL COMMENT '上传者',
    created_at   DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_category (category),
    INDEX idx_status (status)
) ENGINE=InnoDB COMMENT='知识库文档';

-- 知识分片（MySQL 存文本，向量存向量库）
CREATE TABLE knowledge_chunks (
    id           BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    doc_id       BIGINT UNSIGNED NOT NULL COMMENT '所属文档ID',
    chunk_index  INT UNSIGNED   NOT NULL COMMENT '分片序号',
    content      TEXT           NOT NULL COMMENT '分片文本内容',
    token_count  INT UNSIGNED   DEFAULT 0 COMMENT 'token 数量',
    vector_id    VARCHAR(128)   COMMENT '向量数据库中的 ID',
    metadata     JSON           COMMENT '元数据',
    created_at   DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_doc (doc_id),
    FULLTEXT INDEX ft_content (content) WITH PARSER ngram
) ENGINE=InnoDB COMMENT='知识库分片';
```

### 6.4 API 设计

```
# 知识库管理
GET    /api/v1/knowledge/documents              # 文档列表
POST   /api/v1/knowledge/documents              # 上传文档
GET    /api/v1/knowledge/documents/:id           # 文档详情
DELETE /api/v1/knowledge/documents/:id           # 删除文档
POST   /api/v1/knowledge/documents/:id/reindex   # 重新索引
GET    /api/v1/knowledge/documents/:id/chunks    # 查看分片

# 知识检索
POST   /api/v1/knowledge/search                  # 知识检索（测试用）

# 分类管理
GET    /api/v1/knowledge/categories              # 分类列表
GET    /api/v1/knowledge/stats                   # 知识库统计
```

## 七、AI 助手集成（核心改动）

### 7.1 改动点：ai_assistant.go

在对话核心流程中只需新增一步 **RAG 检索增强**：

```go
// AIChatSend 中的修改（伪代码）

// 现有步骤 1-3 不变...

// === 新增: 步骤 3.5 - RAG 检索增强 ===
var ragContext string
if global.RAGEnabled && ragService != nil {
    // 对用户消息做向量检索
    results, err := ragService.Retrieve(ctx, req.Message, 5)
    if err == nil && len(results) > 0 {
        ragContext = ragService.BuildRAGContext(results)
    }
}

// 步骤 4: 构建 System Prompt 时注入 RAG 上下文
systemPrompt := defaultSystemPrompt + toolUsageInstruction
if ragContext != "" {
    systemPrompt += fmt.Sprintf(`

【参考知识库 - 以下内容来自平台知识库，请优先参考】
%s

注意：如果知识库内容与你的通用知识矛盾，以知识库内容为准（因为这是用户环境的实际情况）。
`, ragContext)
}
```

### 7.2 新增 AI 工具

```go
// ai_tools.go 新增
"search_knowledge": {Name: "search_knowledge", RiskLevel: RiskRead, Description: "搜索知识库"},
```

这样 AI 可以主动调用知识库搜索，例如用户问「我们的部署规范是什么」时，AI 会调用 `search_knowledge` 获取相关文档。

### 7.3 needToolCalling 扩展

```go
// 新增知识库相关关键词
"知识库", "文档", "手册", "规范", "标准",
"knowledge", "document", "manual",
```

## 八、前端页面设计

### 8.1 知识库管理页布局

```
知识库管理页:
┌──────────────────────────────────────────────┐
│  知识库    文档数: 42    分片数: 1,286          │
│  [+ 上传文档]  [重新索引全部]                   │
├──────────────────────────────────────────────┤
│  分类筛选: [全部] [K8s] [运维] [故障案例] [业务] │
├──────────────────────────────────────────────┤
│  📄 Nginx 部署最佳实践.md                      │
│     分类: K8s | 分片: 12 | 状态: ✅ 已索引      │
│     标签: nginx, deployment, 最佳实践           │
│     上传者: admin | 2026-04-10                  │
│     [查看分片] [重新索引] [删除]                 │
├──────────────────────────────────────────────┤
│  📄 2026-03 数据库连接池故障复盘.md             │
│     分类: 故障案例 | 分片: 8 | 状态: ✅ 已索引   │
│     标签: MySQL, 连接池, OOM                    │
│     [查看分片] [重新索引] [删除]                 │
├──────────────────────────────────────────────┤
│  📄 生产环境发布审批流程.pdf                    │
│     分类: 业务 | 分片: 15 | 状态: ⏳ 索引中     │
│     [查看分片] [重新索引] [删除]                 │
└──────────────────────────────────────────────┘
```

## 九、知识库内容推荐

| 类别 | 内容示例 | 来源 |
|------|---------|------|
| K8s 最佳实践 | Resource Limits 设置、HPA 配置、NetworkPolicy | 官方文档 |
| 运维手册 | 集群升级流程、节点扩容 SOP、备份恢复 | 内部文档 |
| 故障案例 | OOM 排查、Pod Eviction 分析、网络不通排障 | 历史工单 |
| 业务文档 | 微服务架构说明、API 规范、部署审批流程 | 内部 Wiki |
| CI/CD 规范 | Dockerfile 编写规范、流水线模板说明 | 团队规范 |

## 十、实施步骤

```
Phase 1: 基础 RAG (3-4天)
  ├── 1. pkg/embedding 客户端（调用通义千问 Embedding API）
  ├── 2. pkg/vectordb 接口 + MySQL 全文检索实现（零依赖快速验证）
  ├── 3. 知识库 Service（上传 → 分片 → Embedding → 入库）
  ├── 4. ai_assistant.go 集成 RAG 检索增强
  └── 5. 前端知识库上传/管理页

Phase 2: 向量检索升级 (2-3天)
  ├── 1. 接入 Milvus/pgvector 真正向量检索
  ├── 2. 检索质量优化（重排序、混合检索）
  └── 3. 相似度阈值调优

Phase 3: 自动知识积累 (2天)
  ├── 1. 集群事件自动入库
  ├── 2. 告警记录自动入库
  ├── 3. 部署变更记录自动入库
  └── 4. AI 对话中的有价值 QA 自动入库

Phase 4: 高级能力 (3天)
  ├── 1. PDF 解析支持（go-fitz/unipdf）
  ├── 2. 知识图谱（文档间关系）
  └── 3. 多轮检索（追问时自动扩展检索）
```
