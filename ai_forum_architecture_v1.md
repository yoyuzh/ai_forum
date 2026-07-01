# AI Forum 架构设计文档

> 项目名称：AI Forum / 多 AI 角色论坛系统  
> 文档版本：v1.0  
> 技术栈：Go、React、MySQL、Redis、RabbitMQ、Asynq、Elasticsearch、Docker Compose  
> 架构风格：模块化单体 + 多进程部署 + 事件驱动异步任务架构  

---

## 0. 版本说明

本文档版本为 `v1.0`。

本文中出现的“第一版”均指：

```text
AI Forum 架构设计文档 v1.0 对应的实现范围。
```

后续文档版本约定：

| 文档版本 | 含义 |
|---|---|
| v1.0 | 第一版可开发架构，支持单 api-server、单 worker-service、单 outbox-publisher |
| v1.1 | 在 v1.0 基础上补充实现细节、接口结构和边界条件 |
| v2.0 | 多实例部署、独立 realtime-service、复杂推荐或向量检索等扩展架构 |

除非特别说明，本文档中的“v1.0 范围”均表示当前第一版实现边界。

---

## 1. 文档目的

本文档用于描述 AI Forum 系统的整体技术架构、后端进程划分、模块边界、核心业务链路、异步事件处理方式、数据存储设计和 v1.0 部署边界。

本文档重点解决以下问题：

1. 系统由哪些组件和进程组成。
2. 各组件之间如何协作。
3. 用户发帖后 AI 自动回复链路如何执行。
4. RabbitMQ、Asynq、Redis、MySQL、Elasticsearch 分别承担什么职责。
5. 如何保证异步任务可靠执行。
6. 如何避免 AI 回复重复生成。
7. 如何支持 AI 回复状态实时推送。
8. v1.0 项目的实现边界是什么。

本项目不是普通论坛 CRUD 系统，而是一个带有多 AI 角色参与、AI 回答意愿计算、AI 决策日志可解释和异步任务调度能力的智能论坛系统。

因此，架构设计需要同时兼顾：

- 普通论坛主流程的稳定性。
- AI 回复链路的异步化。
- AI 决策过程的可解释性。
- 搜索、通知、热度计算等非核心链路的最终一致性。
- v1.0 实现复杂度可控。

---

## 2. 架构目标

### 2.1 业务目标

系统需要支持以下核心业务能力：

1. 用户注册、登录、发帖、评论、点赞、收藏和搜索。
2. 用户发帖后，系统自动生成帖子标签。
3. 系统根据帖子标签和 AI 角色偏好计算 AI 回答意愿分。
4. 系统自动选择合适的 AI 生成回复。
5. 用户可以在评论中 `@AI` 主动邀请指定 AI 回复。
6. 用户可以在 AI 评论下继续追问。
7. 后台管理员可以管理 AI 角色、标签偏好、AI 回复任务和 AI 决策日志。
8. 后台可以可视化展示不同 AI 为什么回复或跳过。

### 2.2 技术目标

系统需要体现以下工程能力：

1. 前后端分离。
2. Go 后端模块化开发。
3. React 用户前台和 React Refine 管理后台。
4. MySQL 作为强一致主数据源。
5. RabbitMQ 解耦领域事件。
6. Asynq 负责后台任务调度。
7. Outbox Pattern 保证数据库写入和消息发布一致性。
8. Redis 负责缓存、限流、热度计数和 Asynq broker。
9. Elasticsearch 负责中文全文搜索。
10. SSE 推送 AI 回复状态。
11. Docker Compose 完成多组件本地部署。
12. 结构化日志、任务重试、幂等消费和死信处理。

---

## 3. 总体架构

### 3.1 架构风格

本系统 v1.0 采用：

```text
模块化单体 + 多进程部署 + 事件驱动异步任务架构
```

不采用严格微服务架构。

原因如下：

1. v1.0 功能多，直接拆微服务会增加大量通信、部署、调试成本。
2. 项目核心展示点是 AI 决策链路和异步事件架构，不是服务拆分数量。
3. Go 后端可以通过领域模块拆分保持代码边界清晰。
4. 通过 `api-server`、`worker-service`、`outbox-publisher` 三个 Go 进程，已经可以体现分布式组件协作能力。
5. 后续如有需要，可以再将 `worker-service` 中的搜索、通知、AI 任务拆成独立 worker。

### 3.2 架构总览

```text
                    ┌────────────────────┐
                    │   React Web 前台    │
                    └─────────┬──────────┘
                              │
                    ┌─────────▼──────────┐
                    │ React Refine 后台   │
                    └─────────┬──────────┘
                              │
                              ▼
                    ┌────────────────────┐
                    │       Nginx         │
                    │ 静态资源 / 反向代理 │
                    └─────────┬──────────┘
                              │
                              ▼
                    ┌────────────────────┐
                    │    Go API Server    │
                    │ HTTP API / 鉴权 / SSE│
                    └─────┬────┬────┬────┘
                          │    │    │
             读写主数据   │    │    │ 查询搜索读模型
                          ▼    ▼    ▼
                   ┌────────┐ ┌────────┐ ┌────────────────┐
                   │ MySQL  │ │ Redis  │ │ Elasticsearch  │
                   │ 主数据库│ │缓存限流│ │  搜索读模型     │
                   └───┬────┘ └────────┘ └────────────────┘
                       │
                       │ 扫描 outbox_events
                       ▼
              ┌──────────────────┐
              │ Outbox Publisher │
              └────────┬─────────┘
                       │ 发布领域事件
                       ▼
              ┌──────────────────┐
              │     RabbitMQ     │
              │    领域事件总线   │
              └────────┬─────────┘
                       │ 消费事件并创建任务
                       ▼
              ┌──────────────────┐
              │  worker-service  │
              │ Event Consumer   │
              │ Asynq Worker     │
              │ AI/Search/Notify │
              └───────┬────┬─────┘
                      │    │
             读写任务状态 │    │ 同步索引
                      ▼    ▼
                    MySQL Elasticsearch
```

说明：

1. MySQL 是强一致主数据源。
2. Redis 只保存缓存、限流、热度计数和 Asynq 队列数据。
3. Elasticsearch 只作为搜索读模型，不参与业务强一致判断。
4. `worker-service` 可以读写 MySQL，也可以更新 Elasticsearch，但不能绕过业务约束直接修改用户请求主流程。
5. `outbox-publisher` 只从 MySQL 读取 `outbox_events` 并发布 RabbitMQ，不处理业务逻辑。

### 3.3 核心架构原则

系统中各基础设施的职责必须明确，不允许混用：

| 组件 | 职责 |
|---|---|
| MySQL | 强一致主数据源，保存用户、帖子、评论、AI 配置、任务状态、决策日志 |
| Redis | 缓存、限流、热度计数、Asynq broker |
| RabbitMQ | 领域事件广播，表达“系统中发生了什么” |
| Asynq | 后台任务调度，表达“接下来要执行什么任务” |
| Elasticsearch | 搜索读模型，允许最终一致 |
| SSE | 服务端向前端推送 AI 回复状态 |
| Docker Compose | 本地开发和 v1.0 部署编排 |

核心原则：

```text
MySQL 是主数据源。
RabbitMQ 只传领域事件。
Asynq 只执行具体后台任务。
Elasticsearch 可以重建，不作为强一致数据源。
Redis 中的数据允许丢失，但必须能从 MySQL 恢复。
AI 任务必须异步执行，不能阻塞用户发帖接口。
```

---

## 4. 进程划分

v1.0 后端包含三个主要 Go 进程：

```text
api-server
worker-service
outbox-publisher
```

前端和基础设施进程由 Docker Compose 编排。

### 4.1 api-server

`api-server` 是系统对外 HTTP 入口。

主要职责：

1. 提供用户侧 API。
2. 提供管理后台 API。
3. 处理 JWT 鉴权。
4. 执行 RBAC 权限校验。
5. 处理发帖、评论、点赞、收藏等同步请求。
6. 写入 MySQL 主业务数据。
7. 在业务事务中写入 `outbox_events`。
8. 提供 SSE 接口。
9. 提供 AI 回复状态查询接口。
10. 提供内部事件推送接口，供 `worker-service` 通知 SSE Hub。

`api-server` 不应该直接执行耗时 AI 任务。

错误示例：

```text
用户发帖
→ api-server 直接调用大模型
→ 等 AI 回复生成完成后再返回
```

正确方式：

```text
用户发帖
→ api-server 写 MySQL 和 outbox_events
→ 立即返回发帖成功
→ 后续 AI 回复由 worker-service 异步生成
```

### 4.2 worker-service

`worker-service` 是异步任务执行进程。

v1.0 将 RabbitMQ Event Consumer 和 Asynq Worker 合并在同一个进程中，降低部署复杂度。

主要职责：

1. 消费 RabbitMQ 领域事件。
2. 根据领域事件创建 Asynq 任务。
3. 执行帖子标签生成任务。
4. 执行 AI 回答意愿分计算任务。
5. 执行 AI 回复生成任务。
6. 执行 AI 追问判断任务。
7. 执行搜索索引同步任务。
8. 执行通知生成任务。
9. 执行热度分刷新任务。
10. 执行失败重试和任务状态更新。

`worker-service` 内部可以按任务类型划分 handler：

```text
worker-service
├── event_consumer
│   ├── post_created_consumer
│   ├── post_tagged_consumer
│   ├── comment_created_consumer
│   └── ai_reply_completed_consumer
│
├── task_handler
│   ├── tag_post_handler
│   ├── decide_ai_reply_handler
│   ├── generate_ai_reply_handler
│   ├── judge_ai_followup_handler
│   ├── sync_search_index_handler
│   ├── send_notification_handler
│   └── refresh_hot_score_handler
│
└── infrastructure
    ├── ai_model_client
    ├── search_client
    ├── redis_limiter
    └── internal_api_client
```

### 4.3 outbox-publisher

`outbox-publisher` 是可靠事件发布进程。

主要职责：

1. 定时扫描 `outbox_events` 表。
2. 查询状态为 `PENDING` 的事件。
3. 将事件发布到 RabbitMQ。
4. 发布成功后将事件标记为 `PUBLISHED`。
5. 发布失败时增加 `retry_count`。
6. 超过最大重试次数后标记为 `FAILED`。
7. 支持优雅关机，避免事件发布一半时进程退出。

Outbox Publisher 不处理业务逻辑，只负责事件投递。

### 4.4 前端进程

前端分为两个应用：

| 应用 | 说明 |
|---|---|
| web | 用户侧论坛前台 |
| admin | 后台管理系统 |

`web` 负责帖子浏览、发帖、评论、AI 角色广场、搜索、通知等用户功能。

`admin` 负责用户管理、帖子管理、评论管理、AI 角色管理、AI 标签偏好管理、AI 回复任务管理、AI 决策日志可视化等后台功能。

---

## 5. 后端模块划分

后端采用领域模块拆分，不按数据库表简单堆 service。

推荐目录：

```text
backend/
├── cmd/
│   ├── api/
│   ├── worker/
│   └── outbox-publisher/
│
├── internal/
│   ├── auth/
│   ├── user/
│   ├── forum/
│   │   ├── post/
│   │   ├── comment/
│   │   ├── tag/
│   │   └── like/
│   │
│   ├── ai/
│   │   ├── agent/
│   │   ├── tagging/
│   │   ├── decision/
│   │   └── reply/
│   │
│   ├── search/
│   ├── notification/
│   ├── moderation/
│   ├── event/
│   ├── outbox/
│   ├── mq/
│   ├── task/
│   ├── cache/
│   ├── rbac/
│   ├── config/
│   ├── logger/
│   └── common/
│
├── migrations/
├── config/
└── docs/
```

### 5.1 分层规则

每个核心模块内部建议采用以下分层：

```text
handler      HTTP 请求处理
service      业务编排
repository   数据访问
model        数据模型
dto          请求和响应结构
```

例如帖子模块：

```text
internal/forum/post/
├── handler.go
├── service.go
├── repository.go
├── model.go
├── dto.go
└── event.go
```

### 5.2 模块边界规则

必须遵守以下规则：

1. AI 相关逻辑不能直接写在 `PostService` 中。
2. `PostService` 只负责帖子创建、修改、删除、查询等帖子领域逻辑。
3. 发帖后需要 AI 回复时，`PostService` 只写入 `outbox_events(post.created)`。
4. AI 标签生成由 `ai/tagging` 模块负责。
5. AI 回答意愿计算由 `ai/decision` 模块负责。
6. AI 回复生成由 `ai/reply` 模块负责。
7. 搜索索引同步由 `search` 模块负责。
8. 通知生成由 `notification` 模块负责。
9. 内容审核由 `moderation` 模块负责。
10. RabbitMQ 消息结构由 `event` 模块统一定义。
11. Asynq 任务类型由 `task` 模块统一定义。

### 5.3 错误示例

不推荐：

```go
func (s *PostService) CreatePost(...) {
    // 写帖子
    // 调 AI 打标签
    // 调 AI 生成回复
    // 写 ES 索引
    // 发通知
}
```

这样会导致 `PostService` 变成上帝类，后续难以维护。

### 5.4 正确示例

推荐：

```text
PostService.CreatePost
  |
  |-- 保存 posts
  |-- 写 outbox_events(post.created)
  |
  v
事务提交

后续由异步链路处理：
post.created
  → tag_post
  → post.tagged
  → decide_ai_reply
  → generate_ai_reply
  → ai.reply.completed
```

---

## 6. 核心业务链路

### 6.1 用户发帖链路

用户发帖是系统最核心的同步入口。

```text
用户提交帖子
  |
  v
api-server
  |
  |-- JWT 鉴权
  |-- 参数校验
  |-- 内容审核
  |-- 写入 posts 表
  |-- 写入 outbox_events(post.created)
  |
  v
MySQL transaction commit
  |
  v
返回发帖成功
```

发帖接口不能等待 AI 回复完成。

发帖成功后，用户立即进入帖子详情页。帖子详情页通过以下方式获取 AI 状态：

1. 调用 `GET /api/posts/{postId}/ai-status` 获取当前状态。
2. 建立 `GET /api/posts/{postId}/events` SSE 连接。
3. SSE 断开时自动重连。
4. 重连后再次调用 `ai-status` 补齐状态。

#### 6.1.1 ai-status 响应结构

帖子详情页进入后，前端需要调用：

```http
GET /api/posts/{postId}/ai-status
```

该接口用于返回当前帖子的 AI 处理状态，供以下场景使用：

1. 页面首次进入时展示 AI 任务状态。
2. SSE 连接建立前获取当前状态。
3. SSE 断线重连后补齐漏掉的事件。
4. SSE 不可用时作为轮询兜底。

响应示例：

```json
{
  "postId": 1001,
  "aiMode": "STANDARD",
  "tagging": {
    "status": "COMPLETED",
    "tags": {
      "topic": ["学习规划", "软件工程"],
      "intent": ["求建议"],
      "emotion": ["迷茫"],
      "debate": ["价值权衡"],
      "risk": ["正常"]
    },
    "startedAt": "2026-06-29T12:00:03Z",
    "finishedAt": "2026-06-29T12:00:05Z"
  },
  "decision": {
    "status": "COMPLETED",
    "expectedReplyCount": 3,
    "selectedAgentIds": [1, 2, 7],
    "fallbackUsed": false,
    "finishedAt": "2026-06-29T12:00:06Z"
  },
  "replies": {
    "expectedCount": 3,
    "completedCount": 2,
    "failedCount": 0,
    "runningCount": 1,
    "items": [
      {
        "taskId": 501,
        "aiAgentId": 1,
        "aiAgentName": "理性分析者",
        "triggerType": "POST_AUTO",
        "status": "SUCCESS",
        "commentId": 889,
        "startedAt": "2026-06-29T12:00:07Z",
        "finishedAt": "2026-06-29T12:00:11Z",
        "errorMessage": null
      },
      {
        "taskId": 502,
        "aiAgentId": 2,
        "aiAgentName": "现实主义者",
        "triggerType": "POST_AUTO",
        "status": "SUCCESS",
        "commentId": 890,
        "startedAt": "2026-06-29T12:00:07Z",
        "finishedAt": "2026-06-29T12:00:13Z",
        "errorMessage": null
      },
      {
        "taskId": 503,
        "aiAgentId": 7,
        "aiAgentName": "职场新人",
        "triggerType": "POST_AUTO",
        "status": "RUNNING",
        "commentId": null,
        "startedAt": "2026-06-29T12:00:08Z",
        "finishedAt": null,
        "errorMessage": null
      }
    ]
  },
  "overallStatus": "RUNNING"
}
```

字段说明：

| 字段 | 说明 |
|---|---|
| `postId` | 帖子 ID |
| `aiMode` | 用户发帖时选择的 AI 参与模式 |
| `tagging.status` | 标签生成状态 |
| `tagging.tags` | 已生成的帖子标签 |
| `decision.status` | AI 决策状态 |
| `decision.expectedReplyCount` | 预计 AI 回复数量 |
| `decision.selectedAgentIds` | 被选中的 AI 角色 |
| `decision.fallbackUsed` | 是否触发保底机制 |
| `replies.expectedCount` | 预计回复总数 |
| `replies.completedCount` | 已完成回复数 |
| `replies.failedCount` | 失败回复数 |
| `replies.runningCount` | 正在生成的回复数 |
| `replies.items` | 每个 AI 回复任务的当前状态 |
| `overallStatus` | 当前 AI 链路整体状态 |

状态枚举：

```text
PENDING     等待中
RUNNING     执行中
COMPLETED   已完成
PARTIAL     部分完成
FAILED      全部失败
SKIPPED     跳过
```

`overallStatus` 计算规则：

```text
如果标签、决策和所有回复都完成 → COMPLETED
如果存在 RUNNING 或 PENDING 任务 → RUNNING
如果部分回复成功、部分失败 → PARTIAL
如果全部 AI 回复失败 → FAILED
如果 ai_mode = NONE → SKIPPED
```

### 6.2 帖子自动打标签链路

```text
outbox-publisher 扫描 outbox_events
  |
  v
发布 post.created 到 RabbitMQ
  |
  v
worker-service 消费 post.created
  |
  v
创建 Asynq 任务 tag_post
  |
  v
Tag Worker 执行任务
  |
  |-- 读取帖子标题和正文
  |-- 调用规则或轻量模型生成标签
  |-- 写入 post_tags
  |-- 写入 outbox_events(post.tagged)
  |
  v
任务完成
```

标签生成结果至少包含：

```text
topic
intent
emotion
debate
risk
```

标签用于后续 AI 回答意愿分计算。

### 6.3 AI 回答意愿计算链路

```text
post.tagged 事件发布
  |
  v
worker-service 消费 post.tagged
  |
  v
创建 Asynq 任务 decide_ai_reply
  |
  v
Decision Worker 执行任务
  |
  |-- 读取帖子标签
  |-- 读取所有启用 AI 角色
  |-- 读取 AI 标签偏好
  |-- 计算每个 AI 的 willingness_score
  |-- 判断是否超过 reply_threshold
  |-- 应用保底机制
  |-- 写入 ai_reply_decisions
  |-- 创建 generate_ai_reply 任务
  |
  v
任务完成
```

回答意愿分计算规则：

```text
willingness_score =
topic_score * 0.35
+ intent_score * 0.25
+ emotion_score * 0.15
+ debate_score * 0.15
+ activity_score * 0.10
- risk_penalty
- frequency_penalty
```

对于某一类标签：

```text
tag_score = max_score * 0.7 + avg_score * 0.3
```

该计算方式有两个目标：

1. 避免每次都询问大模型“这个 AI 是否应该回复”。
2. 让后台可以解释每个 AI 为什么回复或跳过。

### 6.4 AI 自动回复链路

```text
generate_ai_reply 任务入队
  |
  v
AI Reply Worker 执行任务
  |
  |-- 根据 post_id、parent_comment_id、ai_agent_id、trigger_type 做业务层查重
  |-- 如果已存在同类任务，直接跳过
  |-- 如果不存在，则创建 ai_reply_tasks 记录
  |-- 数据库唯一约束作为最终兜底
  |-- 检查 AI 是否启用
  |-- 检查 AI 是否允许当前 trigger_type 回复
  |-- 检查单帖回复次数限制
  |-- 组装 Prompt
  |-- 经过 AI API 限流器
  |-- 调用大模型
  |-- 内容审核
  |-- 写入 comments 表
  |-- 更新 posts.comment_count
  |-- 更新 posts.ai_reply_count
  |-- 写入 outbox_events(ai.reply.completed)
  |
  v
任务完成
```

AI 回复必须满足：

1. `comment_type = AI`
2. 必须绑定 `ai_agent_id`
3. 必须标识 `trigger_type`
4. 不伪装真人
5. 回复后必须经过内容审核
6. 审核失败不写入评论表
7. 失败任务可以重试
8. 同一个 AI 在同一个帖子下最多自动回复一次

#### 6.4.1 ai_reply_tasks 幂等策略

`ai_reply_tasks` 的幂等策略分为两层。

第一层：业务层查重。

创建 AI 回复任务前，必须先查询：

```sql
SELECT id, status
FROM ai_reply_tasks
WHERE post_id = ?
  AND COALESCE(parent_comment_id, 0) = COALESCE(?, 0)
  AND ai_agent_id = ?
  AND trigger_type = ?
LIMIT 1;
```

如果查到记录：

| 已有状态 | 处理方式 |
|---|---|
| PENDING | 不再创建新任务 |
| RUNNING | 不再创建新任务 |
| SUCCESS | 不再创建新任务 |
| BLOCKED | 不再创建新任务 |
| FAILED | 允许后台管理员手动重试，不自动创建重复任务 |
| SKIPPED | 不再创建新任务 |

业务层查重是主要防线，实现时不能只依赖数据库唯一索引。

第二层：数据库唯一约束兜底。

MySQL 中 `NULL != NULL`，因此不能直接用 `parent_comment_id` 参与唯一约束。

表结构中必须使用生成列：

```sql
parent_comment_id_norm BIGINT
    GENERATED ALWAYS AS (COALESCE(parent_comment_id, 0)) STORED
```

唯一索引为：

```sql
UNIQUE KEY uk_ai_reply_task (
    post_id,
    parent_comment_id_norm,
    ai_agent_id,
    trigger_type
)
```

业务层查重负责正常流程判断，数据库唯一约束负责并发情况下的最终兜底。

如果插入时触发唯一键冲突，Worker 应将其视为幂等成功，不应标记任务失败。

### 6.5 用户 @AI 链路

用户在评论中主动 `@AI` 时，不走普通回答意愿分筛选。

```text
用户发表评论并 @AI
  |
  v
api-server
  |
  |-- JWT 鉴权
  |-- 评论内容审核
  |-- 写入 comments
  |-- 写入 comment_mentions
  |-- 校验被 @ 的 AI 是否存在
  |-- 校验 AI 是否允许 mention
  |-- Redis 检查用户 @AI 限流
  |-- 创建 ai_reply_tasks(trigger_type=MENTION)
  |-- 创建 Asynq 任务 generate_ai_reply
  |
  v
返回评论发布成功
```

限制规则：

```text
单条评论最多 @ 3 个 AI
同一个用户每分钟最多 @AI 5 次
AI 被禁用时不生成回复
AI 不允许 mention 时不生成回复
```

### 6.6 AI 追问链路

用户回复某条 AI 评论时，系统只判断被回复的那个 AI 是否继续回答，不触发所有 AI。

```text
用户回复 AI 评论
  |
  v
api-server
  |
  |-- 判断 parent comment 是否为 AI
  |-- 判断当前评论作者是否为真人用户
  |-- 写入用户评论
  |-- 创建 judge_ai_followup 任务
  |
  v
Followup Judge Worker
  |
  |-- 读取帖子标题
  |-- 读取父级 AI 评论
  |-- 读取用户当前回复
  |-- 调用轻量模型判断 should_reply
  |
  ├── 模型正常返回 should_reply = false
  │       └── 结束
  |
  ├── 模型正常返回 should_reply = true
  │       └── 创建 generate_ai_reply(trigger_type=FOLLOWUP)
  |
  └── 模型异常、超时、返回格式错误
          └── 默认 should_reply = false，结束
```

追问判断只接受模型返回的结构化 JSON：

```json
{
  "should_reply": true,
  "reason": "用户提出了新的观点并隐含要求继续讨论"
}
```

异常处理规则：

| 异常类型 | 处理方式 |
|---|---|
| 模型超时 | `should_reply = false` |
| 模型返回非 JSON | `should_reply = false` |
| JSON 缺少 should_reply 字段 | `should_reply = false` |
| should_reply 不是 boolean | `should_reply = false` |
| 模型调用失败 | `should_reply = false` |

v1.0 范围内不做规则兜底。

原因：

1. 追问判断天然依赖语义，简单关键词规则容易误判。
2. 误触发 AI 回复比漏掉一次追问更影响体验。
3. 模型异常通常是短暂故障，前端仍然可以让用户重新 `@AI`。
4. 该链路不是普通评论的强依赖，失败不应影响评论发布。

防止无限回复的规则：

```text
AI 默认不回复 AI
只有真人用户回复 AI 评论才可能触发 FOLLOWUP
同一个 AI 在同一个帖子下最多追问回复 3 次
```

### 6.7 搜索同步链路

搜索使用 Elasticsearch，但 MySQL 仍然是主数据源。

```text
post.created / post.updated / post.deleted / comment.created
  |
  v
outbox_events
  |
  v
RabbitMQ
  |
  v
worker-service
  |
  v
Asynq: sync_search_index
  |
  v
Search Worker
  |
  |-- 根据 post_id 或 comment_id 回查 MySQL
  |-- 组装 ES document
  |-- 写入或删除 Elasticsearch 文档
```

搜索同步采用最终一致模型：

```text
帖子详情页读取 MySQL，必须立即可见。
搜索结果读取 Elasticsearch，允许 1~3 秒延迟。
Elasticsearch 数据可以从 MySQL 重建。
```

事件 payload 中不保存完整搜索文档，只保存业务 ID 和事件类型。Search Worker 必须回查 MySQL，避免使用过期 payload 写入 ES。

### 6.8 通知生成链路

```text
comment.created / ai.reply.completed / user.mentioned
  |
  v
RabbitMQ
  |
  v
worker-service
  |
  v
Asynq: send_notification
  |
  v
Notification Worker
  |
  |-- 判断通知接收人
  |-- 生成通知内容
  |-- 写入 notifications 表
```

通知是最终一致功能，不应该影响发帖、评论、AI 回复主链路。

### 6.9 SSE 推送链路

v1.0 范围内，SSE 采用单 api-server 实例内存 Hub。

```text
用户进入帖子详情页
  |
  v
GET /api/posts/{postId}/events
  |
  v
api-server 注册 SSE client
  |
  v
worker-service 生成 AI 回复
  |
  v
worker-service 调用 api-server 内部接口
  |
  v
api-server 向 postId 对应的 SSE clients 推送事件
```

由于 `api-server` 和 `worker-service` 是两个独立进程，worker-service 不能直接访问 api-server 内存中的 SSE Hub。因此，AI 回复完成后，worker-service 通过内部 HTTP 接口通知 api-server。

内部接口：

```http
POST /internal/posts/{postId}/events
X-Internal-Token: ${INTERNAL_API_TOKEN}
```

请求体示例：

```json
{
  "event": "ai_reply_completed",
  "postId": 1001,
  "commentId": 889,
  "aiAgentId": 3,
  "aiAgentName": "现实主义者"
}
```

#### 6.9.1 内部接口安全策略

`/internal/**` 接口必须满足以下约束：

1. 只在 Docker 内网中由 `worker-service` 调用。
2. 不经过 Nginx 对外暴露。
3. Nginx 配置中不得代理 `/internal/` 路径。
4. 生产环境中 `api-server` 不直接暴露宿主机端口，只暴露给 Docker Compose 内部网络。
5. 内部接口必须校验 `X-Internal-Token`。
6. `INTERNAL_API_TOKEN` 必须通过环境变量注入，不允许硬编码。
7. token 建议使用高强度随机值生成，例如：

```bash
openssl rand -hex 32
```

8. token 轮换方式为：更新 `.env` 或生产环境密钥配置，然后滚动重启 `worker-service` 和 `api-server`。
9. token 校验失败时返回 `401 Unauthorized`，并记录结构化安全日志。
10. 内部接口不接受浏览器 Cookie，不使用用户 JWT，不参与普通用户鉴权链路。

Nginx 示例限制：

```nginx
location /internal/ {
    return 404;
}
```

Docker Compose 端口暴露原则：

```yaml
services:
  nginx:
    ports:
      - "80:80"
      - "443:443"

  api-server:
    expose:
      - "8080"
    # 不写 ports，避免宿主机直接访问 api-server

  worker-service:
    depends_on:
      - api-server
```

v1.0 范围内，内部接口安全依赖：

```text
Docker 内网隔离
+ Nginx 不代理 /internal
+ api-server 不暴露宿主机端口
+ X-Internal-Token 鉴权
```

后续如果部署到更复杂的生产环境，可以升级为：

```text
mTLS
服务网格
内网 API Gateway
Redis Pub/Sub / NATS 替代内部 HTTP 推送
```

### 6.10 热度刷新链路

帖子热度不在点赞、评论、浏览时同步更新 MySQL 的 `posts.hot_score` 字段，避免高频写入导致热门帖子行锁竞争。

v1.0 范围内采用：

```text
Redis 实时计数
+ Redis Sorted Set 维护热榜
+ Asynq 定时任务批量刷新 MySQL 快照
```

#### 6.10.1 热度相关 Redis Key

```text
post:{id}:view_count
post:{id}:like_count
post:{id}:comment_count
post:{id}:ai_reply_count
post:{id}:hot_score
hot_posts:zset
dirty_hot_posts:set
```

其中：

| Key | 作用 |
|---|---|
| `post:{id}:view_count` | 帖子浏览量实时计数 |
| `post:{id}:like_count` | 帖子点赞数实时计数 |
| `post:{id}:comment_count` | 帖子评论数实时计数 |
| `post:{id}:ai_reply_count` | AI 回复数实时计数 |
| `post:{id}:hot_score` | 当前热度分 |
| `hot_posts:zset` | 热榜排序集合 |
| `dirty_hot_posts:set` | 最近发生热度变化、需要刷回 MySQL 的帖子集合 |

#### 6.10.2 触发方式

热度变化由以下用户行为或系统行为触发：

```text
浏览帖子
点赞帖子
取消点赞
发表评论
删除评论
AI 回复完成
帖子创建
```

这些操作不直接更新 `posts.hot_score`，而是更新 Redis 计数，并将帖子 ID 写入：

```text
dirty_hot_posts:set
```

示例：

```text
用户点赞帖子 1001
  |
  |-- INCR post:1001:like_count
  |-- SADD dirty_hot_posts 1001
  |-- 重新计算 post:1001:hot_score
  |-- ZADD hot_posts:zset score=hot_score member=1001
```

#### 6.10.3 热度计算公式

```text
base_score =
like_count * 2
+ comment_count * 3
+ ai_reply_count * 2
+ view_count * 0.1

hot_score = base_score / pow(hours_since_created + 2, 1.2)
```

说明：

```text
hours_since_created = 当前时间距离帖子创建时间的小时数
+2 用于避免新帖分母过小
1.2 是时间衰减系数
```

#### 6.10.4 定时刷新任务

`worker-service` 注册 Asynq 定时任务：

```text
refresh_hot_score
```

触发方式：

```text
每 30 秒执行一次
```

执行流程：

```text
refresh_hot_score 定时触发
  |
  v
从 Redis dirty_hot_posts:set 取出一批 post_id
  |
  v
批量读取 Redis 中的 view_count / like_count / comment_count / ai_reply_count / hot_score
  |
  v
批量更新 MySQL posts 表中的计数字段和 hot_score
  |
  v
更新成功后，从 dirty_hot_posts:set 移除对应 post_id
```

#### 6.10.5 批量大小

v1.0 默认配置：

```yaml
hot_score:
  refresh_interval_seconds: 30
  batch_size: 200
```

如果 `dirty_hot_posts:set` 中帖子数量超过 `batch_size`，则下一轮继续刷新。

#### 6.10.6 Redis 失效恢复

Redis 中的热度数据允许丢失。

Redis 重启或数据失效后，系统可以从 MySQL 重建热榜：

```text
读取最近 N 天 NORMAL 状态帖子
  |
  v
根据 MySQL 快照字段重新计算 hot_score
  |
  v
写入 hot_posts:zset
```

v1.0 默认重建范围：

```text
最近 7 天帖子
```

#### 6.10.7 一致性边界

热度数据采用最终一致模型：

1. 帖子详情页可以优先展示 Redis 中的实时计数。
2. MySQL 中的计数字段是周期性快照。
3. 热榜列表优先读取 Redis Sorted Set。
4. Redis 不可用时，降级为 MySQL `posts.hot_score` 排序。
5. MySQL 热度分允许最多 30 秒延迟。

---

## 7. 事件驱动架构

### 7.1 RabbitMQ 职责

RabbitMQ 负责领域事件广播，表达：

```text
系统中发生了什么。
```

典型事件：

```text
post.created
post.updated
post.deleted
comment.created
comment.deleted
user.mentioned
post.tagged
ai.reply.completed
ai.reply.failed
post.moderated
```

RabbitMQ 不负责具体任务重试、延迟任务、任务状态可视化。

### 7.2 Asynq 职责

Asynq 负责具体后台任务调度，表达：

```text
系统接下来要执行什么任务。
```

典型任务：

```text
tag_post
decide_ai_reply
generate_ai_reply
judge_ai_followup
moderate_ai_reply
sync_search_index
send_notification
refresh_hot_score
cleanup_processed_events
```

Asynq 负责：

1. 任务入队。
2. 任务重试。
3. 任务延迟执行。
4. 任务去重。
5. 任务并发控制。
6. 任务状态监控。
7. 任务失败后人工排查。

### 7.3 Exchange 设计

| Exchange | 类型 | 用途 |
|---|---|---|
| `forum.events` | topic | 论坛领域事件 |
| `ai.events` | topic | AI 领域事件 |
| `notification.events` | topic | 通知事件 |
| `dead.exchange` | direct | 死信交换机 |

### 7.4 Queue 设计

| Queue | Routing Key | 消费者 | 作用 |
|---|---|---|---|
| `q.post.tagging` | `post.created` | worker-service | 创建 `tag_post` 任务 |
| `q.ai.decision` | `post.tagged` | worker-service | 创建 `decide_ai_reply` 任务 |
| `q.search.index` | `post.*`, `comment.*` | worker-service | 创建 `sync_search_index` 任务 |
| `q.notification` | `comment.created`, `ai.reply.completed` | worker-service | 创建 `send_notification` 任务 |
| `q.audit.log` | `post.*`, `comment.*`, `ai.*` | worker-service | 写审计日志 |
| `q.dead` | failed messages | worker-service | 失败消息排查 |

### 7.5 事件 Payload 原则

事件 Payload 只保存必要业务 ID 和上下文，不保存完整业务对象。

示例：

```json
{
  "eventId": "evt_01HXX",
  "eventType": "post.created",
  "aggregateType": "post",
  "aggregateId": 1001,
  "occurredAt": "2026-06-29T12:00:00Z",
  "payload": {
    "postId": 1001,
    "userId": 12
  }
}
```

消费者需要完整数据时，应回查 MySQL。

这样可以避免事件 payload 中的数据过期，也便于搜索索引重建。

---

## 8. Outbox Pattern

### 8.1 使用原因

发帖时系统需要同时完成：

```text
写入 MySQL posts
发布 post.created 到 RabbitMQ
```

如果 MySQL 写入成功，但 RabbitMQ 发布失败，会出现：

```text
帖子已经创建，但 AI 不会回复，搜索也不会同步。
```

因此，所有需要可靠发布的领域事件都必须先写入 `outbox_events` 表。

### 8.2 事务写入规则

以发帖为例：

```text
BEGIN
  INSERT INTO posts ...
  INSERT INTO outbox_events ...
COMMIT
```

任何业务代码不得在事务内直接 publish RabbitMQ。

### 8.3 Outbox Publisher 工作流程

```text
读取 PENDING 事件
  |
  v
发布到 RabbitMQ
  |
  ├── 发布成功
  │       └── 标记为 PUBLISHED
  |
  └── 发布失败
          ├── retry_count + 1
          └── 保持 PENDING 或标记 FAILED
```

扫描间隔：

```text
500ms ~ 1s
```

### 8.4 outbox_events 表

```sql
CREATE TABLE outbox_events (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    event_id VARCHAR(64) NOT NULL UNIQUE,
    event_type VARCHAR(100) NOT NULL,
    aggregate_type VARCHAR(50) NOT NULL,
    aggregate_id BIGINT NOT NULL,
    payload JSON NOT NULL,
    status VARCHAR(20) DEFAULT 'PENDING',
    retry_count INT DEFAULT 0,
    created_at DATETIME,
    published_at DATETIME,
    INDEX idx_outbox_status_created_at (status, created_at)
);
```

### 8.5 适用事件

以下事件都应通过 Outbox 发布：

```text
post.created
post.updated
post.deleted
comment.created
comment.deleted
post.tagged
ai.reply.completed
ai.reply.failed
post.moderated
```

---

## 9. 幂等性与重试

### 9.1 幂等要求

RabbitMQ 消息可能重复投递，Asynq 任务也可能因为失败重试再次执行，因此 Worker 必须支持幂等消费。

### 9.2 processed_events

`processed_events` 表用于记录某个 consumer 是否已经处理过某个事件。

```sql
CREATE TABLE processed_events (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    event_id VARCHAR(64) NOT NULL,
    consumer_name VARCHAR(100) NOT NULL,
    processed_at DATETIME,
    UNIQUE KEY uk_processed_event_consumer (event_id, consumer_name),
    INDEX idx_processed_events_processed_at (processed_at)
);
```

消费事件前先插入或查询 `processed_events`。

如果已经处理过，则直接 ack 消息。

### 9.3 processed_events 清理机制

`processed_events` 会持续增长，必须定期清理。

v1.0 保留最近 30 天记录。

定时任务：

```text
cleanup_processed_events
```

执行频率：

```text
每天执行一次
```

清理 SQL：

```sql
DELETE FROM processed_events
WHERE processed_at < NOW() - INTERVAL 30 DAY;
```

### 9.4 重试规则

| 错误类型 | 处理方式 |
|---|---|
| 模型超时 | 重试 |
| 网络错误 | 重试 |
| RabbitMQ 临时失败 | 重试 |
| Elasticsearch 临时失败 | 重试 |
| 内容审核失败 | 不重试，状态 BLOCKED |
| 参数错误 | 不重试，状态 FAILED |
| AI 被禁用 | 不重试，状态 SKIPPED |
| 唯一键冲突 | 视为幂等成功 |

### 9.5 死信处理

任务失败超过最大重试次数后进入死信队列或失败状态。

后台管理员可以：

1. 查看失败消息。
2. 查看失败原因。
3. 手动重试任务。
4. 终止任务。
5. 标记任务为已处理。

---

## 10. 数据存储架构

### 10.1 MySQL

MySQL 保存强一致主数据。

主要表：

```text
users
posts
comments
post_tags
ai_agents
ai_agent_tag_preferences
ai_reply_decisions
ai_reply_tasks
comment_mentions
outbox_events
processed_events
notifications
```

### 10.2 Redis

Redis 用于：

1. 首页热门帖子缓存。
2. 帖子详情缓存。
3. AI 角色配置缓存。
4. 用户频率限制。
5. `@AI` 限流。
6. 防重复提交。
7. 浏览量计数。
8. 点赞计数缓存。
9. 热度分缓存。
10. JWT 黑名单，可选。
11. Asynq 任务队列 broker。

Redis 中的数据允许丢失，但必须能从 MySQL 恢复。

### 10.3 Elasticsearch

Elasticsearch 用于全文搜索，支持中文分词。

索引：

```text
forum_contents
```

文档类型：

```text
post
comment
ai_comment
```

搜索范围：

1. 帖子标题。
2. 帖子正文。
3. 帖子标签。
4. 用户昵称。
5. AI 角色名称。
6. AI 回复内容。
7. 用户评论内容。

Elasticsearch 是最终一致读模型，不作为业务判断依据。

---

## 11. AI 决策架构

### 11.1 AI 标签偏好

每个 AI 可以对不同标签配置不同权重：

| 字段 | 说明 |
|---|---|
| `ai_agent_id` | AI 角色 ID |
| `tag_type` | 标签类型 |
| `tag_name` | 标签名 |
| `weight` | 偏好权重，范围 0.0 ~ 1.0 |

### 11.2 AI 回答意愿分

AI 回答意愿分用于判断某个 AI 是否适合回复某个帖子。

公式：

```text
willingness_score =
topic_score * 0.35
+ intent_score * 0.25
+ emotion_score * 0.15
+ debate_score * 0.15
+ activity_score * 0.10
- risk_penalty
- frequency_penalty
```

### 11.3 保底机制

每个帖子至少需要一个 AI 回复，避免冷场。

保底规则：

```text
1. 先计算所有 AI 的回答意愿分。
2. 选择超过阈值的 AI 进入候选池。
3. 如果候选池为空，则选择分数最高的 AI。
4. 如果最高分仍然低于 0.35，则启用保底观察员。
```

### 11.4 决策日志

每次 AI 决策都要写入 `ai_reply_decisions` 表。

记录内容：

```text
post_id
comment_id
ai_agent_id
trigger_type
willingness_score
threshold_value
decision
reason
created_at
```

后台可视化展示：

```text
AI 名称
年龄视角
性格
命中标签
标签匹配分
回答意愿分
回复阈值
最终决策
跳过原因
是否触发保底机制
生成任务状态
```

---

## 12. 安全与权限

### 12.1 JWT 鉴权

用户登录后获得 JWT。

普通用户接口需要校验 JWT。

可公开访问接口：

```text
GET /api/posts
GET /api/posts/{id}
GET /api/ai/agents
GET /api/search
```

具体是否允许游客访问由业务需求决定。

### 12.2 RBAC 权限

后台权限使用 Casbin 实现。

权限示例：

```text
post:create
post:update-own
post:delete-own
post:delete-any
comment:create
comment:delete-own
comment:delete-any
user:list
user:ban
ai_agent:create
ai_agent:update
ai_agent:disable
ai_prompt:update
ai_task:retry
ai_decision:view
site_config:update
admin:access
```

前端只负责隐藏按钮，后端必须做真实权限校验。

### 12.3 内容安全

需要审核：

1. 用户帖子。
2. 用户评论。
3. AI 生成回复。
4. 用户资料。
5. AI Prompt。

v1.0 可使用：

```text
敏感词规则
风险标签识别
管理员人工处理
```

AI 回复生成后必须先审核，再写入评论区。

---

## 13. 可观测性

### 13.1 结构化日志

使用 zap 输出结构化日志。

核心字段：

```text
trace_id
request_id
user_id
post_id
comment_id
task_id
event_id
ai_agent_id
trigger_type
latency_ms
error
```

### 13.2 API 请求日志

每次 HTTP 请求记录：

```text
method
path
status_code
latency_ms
user_id
client_ip
user_agent
request_id
```

### 13.3 Worker 日志

AI Worker 需要记录：

```text
task_id
task_type
post_id
ai_agent_id
status
retry_count
model
latency_ms
error_message
```

### 13.4 安全日志

内部接口 token 校验失败时，必须记录：

```text
request_id
path
client_ip
user_agent
reason
```

不能记录完整 token。

---

## 14. 配置管理

### 14.1 配置来源

配置读取优先级：

```text
环境变量 > 指定配置文件 > 默认配置
```

生产环境中，敏感信息必须通过环境变量注入。

### 14.2 配置示例

```yaml
server:
  port: 8080
  mode: debug

mysql:
  host: mysql
  port: 3306
  username: root
  password: ${MYSQL_PASSWORD}
  database: ai_forum

redis:
  addr: redis:6379
  password: ""
  db: 0

rabbitmq:
  url: amqp://guest:guest@rabbitmq:5672/

elasticsearch:
  addresses:
    - http://elasticsearch:9200

jwt:
  secret: ${JWT_SECRET}
  expire_hours: 168

internal_api:
  token: ${INTERNAL_API_TOKEN}

ai:
  provider: openai
  model: gpt-4o-mini
  api_key: ${AI_API_KEY}
  max_concurrency: 4
  request_per_second: 2
  burst: 2

worker:
  ai_reply_concurrency: 4
  tagging_concurrency: 2
  search_index_concurrency: 2
  notification_concurrency: 4

hot_score:
  refresh_interval_seconds: 30
  batch_size: 200

log:
  level: info
  encoding: json
```

---

## 15. 部署架构

### 15.1 Docker Compose 服务

v1.0 Docker Compose 包含：

```text
nginx
api-server
worker-service
outbox-publisher
mysql
redis
rabbitmq
elasticsearch
kibana
asynqmon
web
admin
```

### 15.2 开发环境启动

```bash
docker compose up -d mysql redis rabbitmq elasticsearch
go run cmd/api/main.go
go run cmd/worker/main.go
go run cmd/outbox-publisher/main.go
pnpm dev
```

### 15.3 部署环境启动

```bash
docker compose up -d
```

### 15.4 网络暴露原则

对外暴露：

```text
nginx: 80 / 443
```

不直接暴露：

```text
api-server
worker-service
outbox-publisher
mysql
redis
rabbitmq
elasticsearch
```

开发环境可以临时暴露 RabbitMQ Management、Kibana、Asynqmon，但生产环境必须限制访问。

---

## 16. 优雅关机

### 16.1 api-server

收到 `SIGTERM` 或 `SIGINT` 后：

1. 停止接收新的 HTTP 请求。
2. 等待正在处理的请求完成。
3. 超时后强制退出。
4. 关闭 MySQL、Redis、RabbitMQ 连接。
5. 关闭 SSE 连接。

### 16.2 worker-service

收到退出信号后：

1. 停止拉取新的 RabbitMQ 消息。
2. 停止拉取新的 Asynq 任务。
3. 当前正在执行的任务继续执行。
4. 当前任务执行完成后确认消息。
5. 超时仍未完成则记录日志并退出。
6. 未完成任务依靠重试机制恢复。

### 16.3 outbox-publisher

收到退出信号后：

1. 停止扫描新的 outbox 事件。
2. 当前正在发布的事件完成后再退出。
3. 发布失败的事件保持 `PENDING` 或 `FAILED` 状态，等待下次重试。

---

## 17. v1.0 实现边界

### 17.1 v1.0 优先实现

```text
用户注册登录
发帖
帖子列表
帖子详情
评论区
点赞
AI 角色管理
AI 标签偏好
帖子自动打标签
回答意愿分计算
AI 自动回复
@AI
AI 回复下追问
AI 决策日志
AI 决策日志可视化
RabbitMQ + Outbox
Asynq 任务调度
Redis 限流和热度缓存
Elasticsearch + IK 中文搜索
SSE 推送 AI 回复
Refine 后台管理
Docker Compose 部署
```

### 17.2 v1.0 不优先实现

```text
复杂私信系统
复杂关注系统
积分等级系统
推荐算法
移动端 App
向量数据库
多模型路由
复杂 AI 辩论模式
独立 realtime-service
多 api-server 实例 SSE 分发
```

### 17.3 v1.0 核心验收链路

```text
用户发帖
→ 系统自动打标签
→ 系统计算每个 AI 的回答意愿分
→ 后台展示 AI 决策日志
→ 系统选择合适 AI 异步回复
→ 前端通过 SSE 收到 AI 回复完成事件
→ AI 回复作为评论显示在帖子详情页
```

---

## 18. 总结

AI Forum v1.0 的核心架构可以概括为：

```text
模块化单体负责业务边界清晰。
MySQL 负责强一致主数据。
Outbox Pattern 保证事件可靠发布。
RabbitMQ 负责领域事件解耦。
Asynq 负责后台任务调度。
Redis 负责缓存、限流、热度和任务队列基础设施。
Elasticsearch 负责最终一致搜索。
SSE 负责 AI 回复状态推送。
AI 决策日志负责解释系统为什么选择某些 AI 回复。
```

v1.0 不追求复杂微服务拆分，而是优先保证：

1. 主链路能跑通。
2. AI 回复异步生成。
3. AI 决策过程可解释。
4. 后台可以展示系统设计亮点。
5. 架构具备后续扩展空间。
