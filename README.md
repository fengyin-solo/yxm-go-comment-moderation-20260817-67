# Comment Moderation — 评论审核中心

纯 Go 标准库实现的评论审核后端服务，零第三方依赖，开箱即跑。

## 业务说明

管理评论从发布到处置的完整生命周期：**内容发布 → 评论提交 → 规则引擎自动审核 → 人工审核 → 举报处置 → 统计报表**。

- **内容**：被评论的对象（文章/视频），可开关评论区。
- **评论**：状态机 `pending → approved / rejected`，`approved → deleted`。
- **规则**：可配置审核规则（关键词拒绝、长度下限、单用户限流），支持全局或按内容生效；命中规则的评论自动拒绝。
- **审核记录**：自动/人工审核动作的不可变留痕。
- **举报**：用户举报评论，处理成立时自动处置被举报评论。

## 运行

```bash
cd origin
go run ./cmd/server
# 默认监听 :8080，可通过 PORT / ADDR 环境变量修改
```

环境变量：

| 变量 | 默认值 | 说明 |
|------|--------|------|
| PORT | 8080 | 监听端口 |
| ADDR | :PORT | 完整监听地址（优先于 PORT） |
| MAX_PAGE_SIZE | 100 | 分页最大条数 |
| LOG_LEVEL | info | 日志级别：debug/info/warn/error |

## API 一览

### 内容

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/contents | 创建内容 |
| GET | /api/contents | 列表（支持 category/status/keyword 筛选 + 分页） |
| GET | /api/contents/{id} | 详情 |
| PUT | /api/contents/{id} | 更新（含开关评论区） |
| DELETE | /api/contents/{id} | 删除（有未删除评论则拒绝） |
| POST | /api/contents/{id}/batch-approve | 批量通过该内容下全部待审核评论 |

### 评论

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/comments | 发表评论（自动过规则引擎） |
| GET | /api/comments | 列表（支持 content_id/user_id/status 筛选 + 分页） |
| GET | /api/comments/{id} | 详情 |
| POST | /api/comments/{id}/approve | 人工审核通过 |
| POST | /api/comments/{id}/reject | 人工审核拒绝（必须填理由） |
| DELETE | /api/comments/{id} | 删除已发布评论 |

### 审核记录

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/moderations | 列表（支持 comment_id/action/source 筛选 + 分页） |
| GET | /api/moderations/{id} | 详情 |

### 规则

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/rules | 创建规则（keyword/min_length/max_per_user） |
| GET | /api/rules | 列表（支持 type/status/content_id 筛选 + 分页） |
| GET | /api/rules/{id} | 详情 |
| PUT | /api/rules/{id} | 更新 |
| DELETE | /api/rules/{id} | 删除 |

### 举报

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/reports | 举报评论（同一人重复未处理举报拒绝） |
| GET | /api/reports | 列表（支持 comment_id/status 筛选 + 分页） |
| GET | /api/reports/{id} | 详情 |
| POST | /api/reports/{id}/resolve | 举报成立（自动处置被举报评论） |
| POST | /api/reports/{id}/dismiss | 举报驳回 |

### 统计

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/stats/overview | 全局概览（各状态计数/自动拒绝率） |
| GET | /api/stats/by-content | 按内容分组统计 |
| GET | /api/stats/top-moderators | 审核员工作量排行（?n=10） |

### 健康检查

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /healthz | 健康检查 |

## 统一响应格式

```json
{"code": 0, "message": "ok", "data": ...}
```

错误码映射：400 参数校验失败 / 404 记录不存在 / 409 状态冲突或唯一性冲突 / 500 内部错误。

## 工程结构

```
origin/
├── go.mod
├── README.md
├── cmd/server/main.go          # 入口：配置加载、依赖装配、优雅关闭
├── internal/
│   ├── app/app.go              # 依赖装配 store -> service -> handler
│   ├── config/config.go        # 环境变量配置
│   ├── model/                  # 领域模型 + 状态机 + 校验
│   ├── store/                  # Store 接口 + 内存实现
│   ├── service/                # 业务逻辑 + 规则引擎 + 统计
│   └── handler/                # HTTP 路由 + 处理器
└── pkg/
    ├── httpx/httpx.go          # 统一响应、分页、JSON 解析
    ├── idgen/idgen.go          # Hex ID + base62 短码
    └── logger/logger.go        # 分级日志
```

## 测试

```bash
go test ./...
```
