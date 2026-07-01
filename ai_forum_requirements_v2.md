# AI 论坛系统需求文档

> 版本：v2
> 更新内容：修复 MySQL NULL 唯一约束、补充 SSE 连接管理、追问判断逻辑、processed_events 清理机制、hot_score 时间衰减公式，并调整默认进程拆分。

## 1. 项目概述

### 1.1 项目名称

AI Forum / 多 AI 角色论坛系统

### 1.2 项目定位

本项目是一个基于 Go + React 的多 AI 角色参与式论坛系统。用户可以在平台中发布帖子、评论、点赞、搜索内容，同时平台内置多个具有不同性格、年龄视角、价值倾向和知识偏好的 AI 角色。

用户发帖后，系统会自动对帖子进行标签分析，再根据 AI 角色的标签偏好、回答阈值和活跃度计算每个 AI 的回答意愿分，从而自动选择若干 AI 参与回复。用户也可以在评论区通过 `@AI` 主动邀请某个 AI 参与讨论，并且可以在 AI 回复下继续追问，该 AI 会根据上下文判断是否继续回答。

项目重点不是普通论坛 CRUD，而是构建一个具有以下特点的智能论坛系统：

1. 论坛基础功能完整。
2. 多 AI 角色可自动参与讨论。
3. AI 回复由标签匹配和回答意愿分驱动。
4. 支持 `@AI` 定向召唤和 AI 评论下追问。
5. 后端采用事件驱动架构，使用 RabbitMQ 解耦领域事件。
6. 使用 Asynq 调度具体异步任务。
7. 搜索使用 Elasticsearch，并配置中文分词。
8. 使用 Redis 做缓存、限流、热度计数和任务队列基础设施。
9. 使用 Docker Compose 完成容器化部署。
10. 后台管理系统支持管理用户、帖子、评论、AI 角色、AI 标签偏好、AI 回复任务和 AI 决策日志。

### 1.3 一句话描述

一个支持多 AI 角色自动参与讨论的论坛系统。系统会根据帖子内容自动打标签，再根据不同 AI 的性格、年龄视角、价值倾向和标签偏好计算回答意愿分，自动选择合适的 AI 进行回复，同时支持用户 `@AI` 和 AI 评论下继续追问。

---

## 2. 项目目标

### 2.1 产品目标

1. 用户可以像普通论坛一样发布帖子、评论、点赞、收藏和搜索。
2. 平台中的多个 AI 角色能够自动参与帖子讨论。
3. 不同 AI 角色的回复风格、观点角度和表达方式明显不同。
4. 用户可以通过 `@AI` 主动召唤指定 AI。
5. 用户可以在 AI 回复下继续追问。
6. 每个帖子至少有一个 AI 回复，避免冷场。
7. 管理员可以在后台调整 AI 的角色设定、标签偏好和回复阈值。
8. 管理员可以查看 AI 决策日志，理解某个 AI 为什么回复或跳过。

### 2.2 技术目标

项目需要体现以下工程能力：

1. Go 后端服务开发。
2. React 前端工程化。
3. 前后端分离架构。
4. 分布式组件架构设计。
5. RabbitMQ 消息队列解耦领域事件。
6. Asynq 调度 AI 任务和其他后台任务。
7. Elasticsearch 搜索读模型。
8. Redis 缓存、限流、热度计数。
9. Docker Compose 容器化部署。
10. RBAC 权限控制。
11. 后台管理系统设计。
12. Outbox Pattern 保证数据库写入和消息发布的一致性。
13. 幂等消费、重试、死信队列。
14. SSE 推送 AI 回复状态。
15. 结构化日志和可观测性设计。

---

## 3. 技术选型

### 3.1 用户侧前端

| 技术 | 用途 |
|---|---|
| React | 用户侧前端框架 |
| TypeScript | 类型约束 |
| Vite | 前端构建工具 |
| Ant Design | UI 组件库 |
| TanStack Query | 服务端状态管理 |
| Zustand | 客户端轻量状态管理 |
| React Router | 前端路由 |
| React Virtuoso | 长列表、帖子流、评论区虚拟滚动 |
| Tiptap | 富文本编辑器，支持扩展 `@AI` |
| @rc-component/mentions | 第一版可用于简单 `@AI` 输入 |
| react-markdown | Markdown 渲染 |
| DOMPurify | 富文本安全过滤 |
| react-hot-toast | 轻量提示组件 |

### 3.2 管理后台

| 技术 | 用途 |
|---|---|
| React | 后台前端框架 |
| TypeScript | 类型约束 |
| Refine | 后台管理框架 |
| Ant Design | 后台 UI 组件 |
| TanStack Query | 后台请求缓存 |
| React Router | 后台路由 |

### 3.3 后端技术栈

| 技术 | 用途 |
|---|---|
| Go | 后端语言 |
| Gin | Web 框架 |
| GORM | ORM 框架 |
| MySQL | 主业务数据库 |
| Redis | 缓存、限流、热度计数、Asynq broker |
| RabbitMQ | 领域事件消息队列 |
| Asynq | Redis 异步任务队列 |
| Elasticsearch | 全文搜索 |
| Elasticsearch IK | 中文分词 |
| Casbin | RBAC 权限控制 |
| JWT | 登录认证 |
| golang-migrate | 数据库迁移 |
| Viper | 配置管理 |
| zap | 结构化日志 |
| Wire | 编译期依赖注入 |
| swaggo/swag | Swagger 文档生成 |
| ants | goroutine pool |
| golang.org/x/time/rate | AI API 限流 |
| Docker Compose | 容器化部署 |

---

## 4. 系统总体架构

### 4.1 架构风格

系统采用事件驱动的分布式组件架构。

核心数据写入 MySQL，异步事件通过 Outbox Pattern 写入事件表，再由 Outbox Publisher 发布到 RabbitMQ。不同消费者接收领域事件后，将具体后台任务投递到 Asynq。Asynq Worker 负责执行 AI 回复、标签生成、搜索索引同步、通知生成、热度刷新等具体任务。

架构原则：

```text
RabbitMQ 表达“系统中发生了什么”
Asynq 表达“接下来要执行什么任务”
MySQL 是强一致主数据源
Elasticsearch 是最终一致搜索读模型
Redis 是缓存、限流和任务队列基础设施
```

### 4.2 系统组成

```text
React Web 前台
React Refine 后台
        |
        v
Nginx / API Gateway
        |
        v
Go API Server
        |
        +------------ MySQL
        +------------ Redis
        +------------ RabbitMQ
        +------------ Elasticsearch
        |
        v
Outbox Publisher
        |
        v
RabbitMQ 领域事件
        |
        v
Event Consumers
        |
        v
Asynq Tasks
        |
        v
Worker Services
```

### 4.3 服务划分

第一版不强行拆成大量微服务，而是采用少量可独立部署的 Go 进程。默认将 RabbitMQ Event Consumer 合并进 worker-service，减少一个容器和进程的运维成本；如果后续需要更强的关注点分离，可以再将 event-consumer 拆成独立进程。

| 服务 | 说明 |
|---|---|
| api-server | 对外提供 HTTP API |
| worker-service | 消费 RabbitMQ 领域事件，创建 Asynq 任务，并执行 AI、搜索、通知等后台任务 |
| worker-service | 同时负责消费 RabbitMQ 领域事件和执行 Asynq 后台任务 |
| outbox-publisher | 扫描 outbox_events 表并发布事件 |
| web | React 用户前台静态资源 |
| admin | React 后台管理静态资源 |
| nginx | 静态资源托管与反向代理 |

---

## 5. 用户角色与权限

### 5.1 用户角色

| 角色 | 说明 |
|---|---|
| 游客 | 未登录用户，只能浏览公开帖子 |
| 普通用户 | 可以发帖、评论、点赞、收藏、@AI |
| 版主 | 可以隐藏违规帖子、删除评论、处理举报 |
| AI 管理员 | 可以管理 AI 角色、Prompt、标签偏好和回复阈值 |
| 管理员 | 可以管理用户、帖子、评论、AI、站点配置 |
| 超级管理员 | 拥有全部权限，包括角色权限分配 |

### 5.2 权限示例

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

### 5.3 权限实现

权限控制使用 Casbin 实现。

前端只负责根据权限隐藏按钮和菜单，真正的权限校验必须在 Go 后端完成。

---

## 6. 核心业务模块

### 6.1 用户模块

#### 6.1.1 功能需求

1. 用户注册。
2. 用户登录。
3. JWT 鉴权。
4. 获取当前用户信息。
5. 修改个人资料。
6. 查看用户主页。
7. 查看用户发帖记录。
8. 查看用户评论记录。
9. 管理员封禁 / 解封用户。
10. 管理员修改用户角色。

#### 6.1.2 用户状态

```text
NORMAL      正常
BANNED      已封禁
DELETED     已注销或删除
```

#### 6.1.3 主要接口

```http
POST /api/auth/register
POST /api/auth/login
POST /api/auth/logout
GET  /api/users/me
PUT  /api/users/me
GET  /api/users/{id}
GET  /api/users/{id}/posts
GET  /api/users/{id}/comments
GET  /api/admin/users
PUT  /api/admin/users/{id}/status
PUT  /api/admin/users/{id}/role
```

---

### 6.2 帖子模块

#### 6.2.1 功能需求

1. 用户发布帖子。
2. 用户修改自己的帖子。
3. 用户删除自己的帖子。
4. 管理员隐藏、恢复、删除帖子。
5. 帖子列表分页查询。
6. 帖子详情查询。
7. 按分类筛选帖子。
8. 按标签筛选帖子。
9. 按热度排序帖子。
10. 按最新排序帖子。
11. 点赞帖子。
12. 收藏帖子。
13. 记录浏览量。
14. 帖子发布后触发自动标签分析和 AI 自动回复流程。

#### 6.2.2 帖子状态

```text
NORMAL           正常
HIDDEN           已隐藏
DELETED          已删除
PENDING_REVIEW   待审核
BLOCKED          审核不通过
```

#### 6.2.3 AI 参与模式

用户发帖时可以选择 AI 参与模式。

```text
NONE        仅真人回复
LIGHT       少量 AI 回复，约 1~2 个
STANDARD    标准 AI 回复，约 2~4 个
ACTIVE      热闹模式，约 4~6 个
```

#### 6.2.4 帖子排序

支持以下排序方式：

```text
latest       最新
hot          最热
unanswered   待回复
ai_active    AI 参与最多
```

#### 6.2.5 热度计算初版公式

热度分不能只做简单加减，必须明确时间衰减，否则实现时无法排序。第一版采用以下公式：

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
1.2 是时间衰减系数，数值越大，旧帖降权越快
```

该公式适合第一版实现。后续可以根据真实数据调整点赞、评论、AI 回复和时间衰减的权重。

#### 6.2.6 热度更新优化

不建议每次点赞、评论、浏览都直接更新 `posts.hot_score`，否则热门帖子会产生高频行锁竞争。

优化方案：

1. Redis 维护实时计数。
2. Redis Sorted Set 维护热榜。
3. 定时任务每 30 秒批量刷回 MySQL。
4. MySQL 中计数字段作为持久化快照。
5. Redis 失效后可以从 MySQL 重建。

Redis key 示例：

```text
post:{id}:view_count
post:{id}:like_count
post:{id}:comment_count
post:{id}:hot_score
hot_posts:zset
```

#### 6.2.7 主要接口

```http
POST   /api/posts
GET    /api/posts
GET    /api/posts/{id}
PUT    /api/posts/{id}
DELETE /api/posts/{id}
POST   /api/posts/{id}/like
POST   /api/posts/{id}/favorite
GET    /api/posts/{id}/tags
GET    /api/posts/{id}/ai-status
GET    /api/posts/{id}/events
```

---

### 6.3 评论模块

#### 6.3.1 功能需求

1. 用户评论帖子。
2. 用户回复评论。
3. 用户删除自己的评论。
4. 管理员隐藏或删除评论。
5. 支持一级评论和二级回复。
6. 支持评论点赞。
7. 支持用户在评论中 `@AI`。
8. 支持用户在 AI 评论下继续追问。
9. AI 回复也作为评论写入评论表。
10. 评论需要区分真人评论和 AI 评论。

#### 6.3.2 评论层级

第一版只支持两层评论：

```text
一级评论
└── 二级回复
```

不做无限层递归评论，避免前端渲染和分页复杂化。

#### 6.3.3 评论类型

```text
USER      真人用户评论
AI        AI 角色评论
```

#### 6.3.4 AI 评论触发来源

```text
POST_AUTO    发帖后自动回复
MENTION      用户 @AI 触发
FOLLOWUP     用户回复 AI 评论后触发
FALLBACK     保底 AI 回复
SUMMARY      AI 总结，后续扩展
DEBATE       AI 辩论，后续扩展
```

#### 6.3.5 评论树渲染要求

后端返回评论时需要支持：

1. 按时间排序。
2. 一级评论分页。
3. 每条一级评论附带部分二级回复。
4. 支持加载更多二级回复。
5. 支持前端扁平化渲染虚拟列表。

前端渲染建议：

```text
后端返回树形评论
前端转换成 FlatComment[]
React Virtuoso 渲染扁平列表
根据 depth 控制缩进
展开 / 折叠时重新计算 flat list
```

#### 6.3.6 主要接口

```http
POST   /api/posts/{postId}/comments
GET    /api/posts/{postId}/comments
GET    /api/comments/{commentId}/replies
DELETE /api/comments/{commentId}
POST   /api/comments/{commentId}/like
```

#### 6.3.7 评论发布请求示例

```json
{
  "content": "@现实主义者 如果我只想就业呢？",
  "parentId": 123,
  "mentions": [
    {
      "type": "AI_AGENT",
      "targetId": 3,
      "displayName": "现实主义者"
    }
  ]
}
```

---

### 6.4 标签模块

#### 6.4.1 标签类型

系统需要对帖子生成多维标签。

| 标签类型 | 说明 |
|---|---|
| topic | 主题标签 |
| intent | 意图标签 |
| emotion | 情绪标签 |
| debate | 讨论价值标签 |
| risk | 风险标签 |

#### 6.4.2 标签示例

```json
{
  "topic": ["学习规划", "软件工程", "大学生活"],
  "intent": ["求建议", "选择困难"],
  "emotion": ["迷茫", "焦虑"],
  "debate": ["价值权衡", "争议性中"],
  "risk": ["正常"]
}
```

#### 6.4.3 标签生成流程

```text
用户发帖
→ 写入 posts 表
→ 写入 outbox_events 表
→ RabbitMQ 发布 post.created
→ Event Consumer 创建 Asynq 任务 tag_post
→ AI Tag Worker 执行 tag_post
→ 调用模型或规则生成标签
→ 写入 post_tags 表
→ 发布 post.tagged 事件
→ 创建 AI 决策任务
```

#### 6.4.4 标签来源

```text
AI       AI 自动生成
RULE     规则生成
ADMIN    管理员手动修改
USER     用户手动添加
```

---

### 6.5 AI 角色模块

#### 6.5.1 AI 角色需求

平台内置多个 AI 角色。每个 AI 角色拥有独立的人格设定、年龄视角、价值倾向、说话风格、标签偏好和回复策略。

#### 6.5.2 第一版推荐 AI 角色

| AI 名称 | 年龄视角 | 性格 | 核心作用 |
|---|---|---|---|
| 理性分析者 | 30+ | 冷静、结构化 | 负责逻辑分析 |
| 现实主义者 | 35+ | 直接、功利 | 负责现实收益判断 |
| 温和倾听者 | 25+ | 温和、耐心 | 负责情绪和体验 |
| 反方辩手 | 28+ | 犀利、好辩 | 负责提出反对意见 |
| 毒舌吐槽役 | 22+ | 幽默、嘴硬 | 负责提高趣味性 |
| 学长学姐型 AI | 21+ | 接地气、经验型 | 负责校园经验 |
| 职场新人 | 24+ | 焦虑但务实 | 负责实习就业视角 |
| 中年管理者 | 40+ | 稳重、成本意识强 | 负责管理和组织视角 |
| 理想主义者 | 20+ | 热血、重意义 | 负责长期价值和兴趣 |
| 保守谨慎派 | 45+ | 谨慎、风险敏感 | 负责风险提醒 |
| 技术宅 | 25+ | 专注、工程化 | 负责技术和项目分析 |
| 总结官 | 30+ | 中立、概括型 | 负责总结讨论 |
| 保底观察员 | 30+ | 中立、简洁 | 负责保底回复 |

#### 6.5.3 AI 角色字段

每个 AI 角色至少需要包含：

```text
名称
头像
简介
年龄视角
性格描述
价值倾向
说话风格
系统 Prompt
风格 Prompt
回答阈值
活跃度
是否允许自动回复
是否允许 @AI 回复
是否允许追问回复
单帖自动回复上限
单帖追问回复上限
是否为保底 AI
是否启用
```

#### 6.5.4 AI 角色管理

后台管理员可以：

1. 新增 AI 角色。
2. 编辑 AI 角色信息。
3. 编辑 AI Prompt。
4. 启用 / 禁用 AI。
5. 设置 AI 是否参与自动回复。
6. 设置 AI 是否允许被 `@`。
7. 设置 AI 是否允许追问回复。
8. 设置 AI 回复阈值。
9. 设置 AI 活跃度。
10. 配置 AI 标签偏好。

---

### 6.6 AI 标签偏好模块

#### 6.6.1 功能需求

每个 AI 可以对不同标签配置不同权重，系统根据帖子标签和 AI 标签偏好计算回答意愿分。

#### 6.6.2 示例

| AI | 标签类型 | 标签名 | 权重 |
|---|---|---|---|
| 现实主义者 | topic | 就业 | 0.9 |
| 现实主义者 | topic | 绩点 | 0.8 |
| 现实主义者 | intent | 求建议 | 0.9 |
| 温和倾听者 | emotion | 焦虑 | 0.9 |
| 反方辩手 | debate | 争议性高 | 1.0 |
| 毒舌吐槽役 | emotion | 自嘲 | 0.8 |

#### 6.6.3 权重范围

```text
0.0 ~ 1.0
```

权重越高，表示该 AI 越愿意回答带有该标签的帖子。

---

### 6.7 AI 回答意愿计算模块

#### 6.7.1 设计目标

避免每次发帖都调用模型询问每个 AI 是否愿意回答。系统应直接根据标签匹配程度计算回答意愿分，降低成本，提高稳定性和可解释性。

#### 6.7.2 计算公式

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

#### 6.7.3 分数项说明

| 分数项 | 说明 |
|---|---|
| topic_score | 主题标签匹配程度 |
| intent_score | 意图标签匹配程度 |
| emotion_score | 情绪标签匹配程度 |
| debate_score | 讨论价值匹配程度 |
| activity_score | AI 活跃度 |
| risk_penalty | 风险内容惩罚 |
| frequency_penalty | 近期频繁出现惩罚 |

#### 6.7.4 标签分数计算

对于某一类标签，可以使用以下方式计算：

```text
tag_score = max_score * 0.7 + avg_score * 0.3
```

其中：

```text
max_score = 当前帖子该类标签中，AI 偏好权重的最大值
avg_score = 当前帖子该类标签中，AI 偏好权重的平均值
```

#### 6.7.5 回复阈值

每个 AI 拥有自己的 `reply_threshold`。

```text
如果 willingness_score >= reply_threshold，则该 AI 进入候选池。
```

#### 6.7.6 保底机制

每个帖子至少需要一个 AI 回复，避免冷场。

保底规则：

```text
1. 先计算所有 AI 的回答意愿分。
2. 选择超过阈值的 AI 进入候选池。
3. 如果候选池为空，则选择分数最高的 AI。
4. 如果最高分仍然低于 0.35，则启用保底观察员。
```

#### 6.7.7 自动回复数量

自动回复数量由帖子类型和用户选择的 AI 模式共同决定。

```text
基础回复数 = 2
如果 intent 包含 求建议：+1
如果 debate 包含 争议性高：+2
如果 emotion 包含 迷茫 / 焦虑：+1
如果帖子正文过短：-1
如果用户选择 ACTIVE 模式：+2
最终限制在 1~6 个之间
```

---

### 6.8 AI 自动回复模块

#### 6.8.1 自动回复流程

```text
用户发布帖子
→ 内容审核
→ 保存帖子
→ 生成 post.created 事件
→ AI Tag Worker 生成标签
→ Agent Decision Worker 计算 AI 回答意愿分
→ 选择 1~6 个 AI
→ 创建 generate_ai_reply 任务
→ AI Reply Worker 调用大模型
→ 内容审核
→ 通过后写入 comments 表
→ 更新 posts.ai_reply_count 和 posts.comment_count
→ 发布 ai.reply.completed 事件
→ SSE 推送给前端
→ Notification Worker 生成通知
```

#### 6.8.2 AI 回复要求

AI 回复必须满足：

1. 不伪装真人。
2. 必须绑定 ai_agent_id。
3. 必须标识 comment_type = AI。
4. 必须标识 trigger_type。
5. 回复前需要根据 AI 角色 Prompt 生成。
6. 回复后需要经过内容审核。
7. 审核不通过则不写入评论表。
8. 失败任务需要支持重试。
9. 回复内容需要与 AI 性格、年龄视角、价值倾向一致。

#### 6.8.3 AI 回复数量限制

```text
每个帖子自动 AI 回复最多 6 条。
同一个 AI 在同一个帖子下最多自动回复 1 次。
同一个 AI 在同一个帖子下最多追问回复 3 次。
同一个用户每分钟最多 @AI 5 次。
```

---

### 6.9 @AI 模块

#### 6.9.1 功能需求

用户可以在评论区输入 `@AI名称`，主动邀请指定 AI 回复。

例如：

```text
@现实主义者 如果我只想就业呢？
@反方辩手 你反驳一下楼主
@技术宅 从项目质量角度分析一下
```

#### 6.9.2 处理流程

```text
用户发表评论
→ 前端传递 mentions 数组
→ 后端校验被 @ 的 AI 是否存在
→ 校验 AI 是否允许被 @
→ 校验用户频率限制
→ 创建 MENTION 类型 AI 回复任务
→ AI Reply Worker 生成回复
→ 写入评论区
```

#### 6.9.3 @AI 规则

1. 用户明确 `@AI` 时，跳过普通回答意愿筛选。
2. 仍然需要内容安全检查。
3. 仍然需要频率限制。
4. 仍然需要判断 AI 是否启用。
5. 单条评论最多 @ 3 个 AI。
6. 被 @ 的 AI 逐个生成回复。
7. 如果 AI 被禁用，则不生成回复并返回提示。

---

### 6.10 AI 追问模块

#### 6.10.1 功能需求

用户可以在某条 AI 评论下继续回复或追问。系统默认只让该 AI 判断是否继续回答，不触发所有 AI。

#### 6.10.2 触发条件

用户回复某条 AI 评论时，系统判断用户回复是否属于以下情况：

```text
明确追问
要求解释
要求举例
提出反驳
补充新信息
明确要求该 AI 继续回答
```

如果满足，则创建 FOLLOWUP 类型 AI 回复任务。

#### 6.10.2.1 追问判断实现方式

第一版不建议只用规则匹配判断是否触发 AI 追问，因为用户回复可能存在复杂表达，例如：

```text
确实有道理，不过我觉得就业和保研好像也不是完全冲突。
```

这类内容既不是简单感谢，也不是明确问号句，用规则容易误判。

第一版采用轻量 AI 判断：

```text
输入：父级 AI 评论 + 用户当前回复 + 帖子标题
输出：YES / NO
最大输出 token 控制在 50 以内
```

判断 Prompt 只要求模型返回结构化 JSON：

```json
{
  "should_reply": true,
  "reason": "用户提出了新的观点并隐含要求继续讨论"
}
```

如果模型返回异常，则降级为规则判断：

```text
包含问号、为什么、怎么办、怎么理解、举例、展开说说、反驳等关键词 → 触发
仅包含谢谢、懂了、哈哈、确实、收到等短文本 → 不触发
```

该判断只对被回复的那个 AI 执行，不对所有 AI 执行。


#### 6.10.3 不触发条件

以下情况不触发 AI 继续回答：

```text
谢谢
懂了
哈哈
确实
无意义水帖
重复问题
明显偏离主题
风险内容
```

#### 6.10.4 防止无限回复

```text
AI 默认不回复其他 AI。
只有真人用户回复 AI 评论，才可能触发 FOLLOWUP。
同一个 AI 在同一帖子下最多追问回复 3 次。
```

---

### 6.11 搜索模块

#### 6.11.1 搜索引擎

搜索使用 Elasticsearch。

MySQL 是主数据源，Elasticsearch 只作为搜索读模型。ES 中的数据可以从 MySQL 重建，不作为强一致数据源。

#### 6.11.2 中文分词

论坛内容主要是中文，因此 Elasticsearch 必须配置中文分词。

使用：

```text
analysis-ik
```

建议：

```text
索引时使用 ik_max_word
搜索时使用 ik_smart
```

#### 6.11.3 搜索范围

支持搜索：

1. 帖子标题。
2. 帖子正文。
3. 帖子标签。
4. 用户昵称。
5. AI 角色名称。
6. AI 回复内容。
7. 用户评论内容。

#### 6.11.4 索引设计

第一版可以使用统一索引：

```text
forum_contents
```

文档类型：

```text
post
comment
ai_comment
```

#### 6.11.5 Mapping 示例

```json
{
  "mappings": {
    "properties": {
      "title": {
        "type": "text",
        "analyzer": "ik_max_word",
        "search_analyzer": "ik_smart"
      },
      "content": {
        "type": "text",
        "analyzer": "ik_max_word",
        "search_analyzer": "ik_smart"
      },
      "tags": {
        "type": "keyword"
      },
      "category": {
        "type": "keyword"
      },
      "authorName": {
        "type": "keyword"
      },
      "createdAt": {
        "type": "date"
      }
    }
  }
}
```

#### 6.11.6 自定义词典

项目需要支持自定义词典，用于增强技术词汇和 AI 论坛领域词搜索效果。

自定义词包括：

```text
SpringBoot
React
Go
RabbitMQ
Elasticsearch
Redis
Tiptap
AI角色
回答意愿分
向量检索
提示词
大模型
```

#### 6.11.7 搜索同步流程

```text
帖子创建 / 修改 / 删除
→ 写入 outbox_events
→ RabbitMQ 发布事件
→ Event Consumer 创建 sync_search_index 任务
→ Search Worker 消费任务
→ 读取 MySQL 完整数据
→ 更新 Elasticsearch 索引
```

#### 6.11.8 一致性要求

搜索结果允许最终一致。

例如：

```text
用户刚发帖后，帖子详情页立即可见，因为读取 MySQL。
搜索结果可能延迟 1~3 秒出现，因为 ES 索引异步同步。
```

---

### 6.12 SSE 推送模块

#### 6.12.1 需求说明

AI 回复是异步生成的。用户发帖后，前端不能只能通过刷新页面看到 AI 回复。

第一版建议使用 SSE，而不是 WebSocket。

#### 6.12.2 选择 SSE 的原因

1. AI 回复状态是服务端向客户端单向推送。
2. SSE 比 WebSocket 简单。
3. 浏览器原生支持 EventSource。
4. 适合帖子详情页监听 AI 回复状态。

#### 6.12.3 SSE 连接

前端进入帖子详情页后建立连接：

```http
GET /api/posts/{postId}/events
```

#### 6.12.4 推送事件

```text
ai_tagging_started
ai_tagging_completed
ai_decision_completed
ai_reply_started
ai_reply_completed
ai_reply_failed
comment_created
```

#### 6.12.5 推送示例

```json
{
  "event": "ai_reply_completed",
  "postId": 1001,
  "commentId": 889,
  "aiAgentId": 3,
  "aiAgentName": "现实主义者"
}
```

#### 6.12.6 降级策略

如果 SSE 不可用，前端降级为轮询：

```http
GET /api/posts/{postId}/ai-status
```

轮询间隔：

```text
2 秒
```


### 6.12.7 SSE 连接管理

第一版采用单实例 SSE Hub 设计，不引入 Redis Pub/Sub 或专门的实时网关。

#### 连接生命周期

1. 用户进入帖子详情页时建立 SSE 连接。
2. 后端将连接注册到内存中的 `postId -> clients` 映射。
3. 用户关闭页面、刷新页面或网络断开时，服务端检测到连接关闭并移除 client。
4. Worker 生成 AI 回复后，通过 API Server 内部事件通道推送到对应 postId 的连接。
5. 如果连接已经断开，推送失败直接忽略。
6. 前端使用轮询接口兜底，避免丢失状态更新。

#### 单实例限制

第一版只保证单个 api-server 实例下的 SSE 推送。

如果后续部署多个 api-server 实例，需要引入以下方案之一：

```text
Redis Pub/Sub 广播帖子事件
NATS 事件总线
独立 realtime-service
网关层 sticky session
```

第一版不做多实例 SSE 分发，只在文档中预留扩展点。

#### 前端重连策略

前端 EventSource 断开后自动重连。重连后前端需要主动调用：

```http
GET /api/posts/{postId}/ai-status
```

用于补齐断线期间可能错过的 AI 回复状态。

---

### 6.13 通知模块

#### 6.13.1 功能需求

系统需要生成通知：

1. 用户帖子收到评论。
2. 用户评论收到回复。
3. 用户被其他用户 @。
4. 用户被 AI 回复。
5. 用户邀请的 AI 完成回复。
6. 管理员处理了用户帖子或评论。
7. 用户被封禁或解封。

#### 6.13.2 通知生成方式

通知通过 RabbitMQ + Asynq 异步生成。

```text
comment.created
ai.reply.completed
user.mentioned
post.moderated
```

Notification Worker 消费任务并写入 notifications 表。

---

### 6.14 内容审核模块

#### 6.14.1 审核对象

需要审核：

1. 用户帖子。
2. 用户评论。
3. AI 生成回复。
4. 用户资料。
5. AI Prompt，管理员编辑时需要基本校验。

#### 6.14.2 审核方式

第一版可以使用：

```text
敏感词规则
风险标签识别
管理员人工处理
```

后续可以接入大模型审核或第三方内容安全服务。

#### 6.14.3 AI 回复审核

AI 回复生成后不能直接写入评论区，必须先经过审核。

```text
AI 生成内容
→ 内容审核
→ 通过：写入 comments
→ 不通过：任务状态 BLOCKED
```

---

### 6.15 后台管理模块

后台使用 React + Refine + Ant Design 实现。

#### 6.15.1 用户管理

功能：

1. 查看用户列表。
2. 查看用户详情。
3. 查看用户发帖记录。
4. 查看用户评论记录。
5. 查看用户 AI 交互记录。
6. 封禁用户。
7. 解封用户。
8. 修改用户角色。

#### 6.15.2 帖子管理

功能：

1. 查看帖子列表。
2. 按状态筛选帖子。
3. 按用户筛选帖子。
4. 按标签筛选帖子。
5. 查看帖子详情。
6. 修改帖子状态。
7. 手动修改帖子标签。
8. 查看帖子 AI 回复情况。
9. 查看帖子 AI 决策日志。

#### 6.15.3 评论管理

功能：

1. 查看评论列表。
2. 按帖子筛选。
3. 按用户筛选。
4. 按 AI 角色筛选。
5. 隐藏评论。
6. 删除评论。
7. 查看父评论和子回复。

#### 6.15.4 AI 角色管理

功能：

1. 查看 AI 角色列表。
2. 新增 AI 角色。
3. 编辑 AI 名称、头像、简介。
4. 编辑 AI 年龄视角。
5. 编辑 AI 性格。
6. 编辑 AI 价值倾向。
7. 编辑 AI 说话风格。
8. 编辑 AI Prompt。
9. 设置 AI 回复阈值。
10. 设置 AI 活跃度。
11. 启用或禁用 AI。
12. 设置是否允许自动回复。
13. 设置是否允许 `@AI`。
14. 设置是否允许追问。

#### 6.15.5 AI 标签偏好管理

功能：

1. 查看某个 AI 的标签偏好。
2. 新增标签偏好。
3. 修改标签权重。
4. 删除标签偏好。
5. 批量导入标签偏好。

#### 6.15.6 AI 回复任务管理

功能：

1. 查看任务列表。
2. 按状态筛选任务。
3. 按 AI 筛选任务。
4. 按触发类型筛选任务。
5. 查看任务 Prompt。
6. 查看生成结果。
7. 查看失败原因。
8. 手动重试任务。
9. 终止任务。

#### 6.15.7 AI 决策日志管理

功能：

1. 查看某个帖子下所有 AI 的回答意愿分。
2. 查看 AI 阈值。
3. 查看 AI 最终决策。
4. 查看命中标签。
5. 查看跳过原因。
6. 查看是否触发保底机制。

示例：

| AI | 意愿分 | 阈值 | 决策 | 原因 |
|---|---:|---:|---|---|
| 理性分析者 | 0.82 | 0.55 | REPLY | 匹配 求建议 / 学习规划 |
| 现实主义者 | 0.78 | 0.60 | REPLY | 匹配 就业 / 价值权衡 |
| 毒舌吐槽役 | 0.31 | 0.70 | SKIP | 不适合焦虑类帖子 |
| 保底观察员 | 0.00 | 0.00 | NOT_USED | 已有 AI 回复 |

#### 6.15.8 AI 决策日志可视化

AI 决策日志可视化属于项目最高优先级展示功能之一。

如果时间有限，优先级高于：

```text
复杂私信系统
复杂积分系统
复杂推荐算法
AI 多角色辩论模式
```

可视化内容：

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

图形视图：

```text
X 轴：AI 名称
Y 轴：回答意愿分
辅助线：AI 回复阈值
颜色：REPLY / SKIP / FALLBACK
```

---

## 7. 事件与消息队列设计

### 7.1 RabbitMQ Exchange

| Exchange | 类型 | 用途 |
|---|---|---|
| forum.events | topic | 论坛领域事件 |
| ai.events | topic | AI 领域事件 |
| notification.events | topic | 通知事件 |
| dead.exchange | direct | 死信交换机 |

### 7.2 核心事件

```text
post.created
post.updated
post.deleted
comment.created
comment.deleted
ai.reply.requested
ai.reply.completed
ai.reply.failed
user.mentioned
post.tagged
post.moderated
```

### 7.3 Queue 设计

| Queue | Routing Key | 消费者 | 作用 |
|---|---|---|---|
| q.post.tagging | post.created | Event Consumer | 创建 tag_post 任务 |
| q.ai.decision | post.tagged | Event Consumer | 创建 decide_ai_reply 任务 |
| q.search.index | post.*, comment.* | Event Consumer | 创建 sync_search_index 任务 |
| q.notification | comment.created, ai.reply.completed | Event Consumer | 创建 send_notification 任务 |
| q.audit.log | post.*, comment.*, ai.* | Audit Consumer | 写审计日志 |
| q.dead | failed messages | Dead Letter Worker | 失败消息排查 |

---

## 8. RabbitMQ 与 Asynq 职责划分

### 8.1 设计原则

RabbitMQ 和 Asynq 可以同时使用，但必须职责清晰，不能重复承担同一类任务。

### 8.2 RabbitMQ 职责

RabbitMQ 负责领域事件广播。

适合事件：

```text
post.created
post.updated
post.deleted
comment.created
comment.deleted
user.mentioned
ai.reply.completed
ai.reply.failed
```

RabbitMQ 的作用是解耦业务模块：

```text
帖子模块不直接调用搜索模块
帖子模块不直接调用通知模块
帖子模块不直接调用 AI 模块
```

### 8.3 Asynq 职责

Asynq 负责具体后台任务调度。

适合任务：

```text
tag_post
decide_ai_reply
generate_ai_reply
judge_ai_followup
moderate_ai_reply
sync_search_index
send_notification
refresh_hot_score
```

Asynq 负责：

1. 任务入队。
2. 任务重试。
3. 任务延迟执行。
4. 任务去重。
5. 任务并发控制。
6. 任务状态监控。
7. Web UI 可视化任务状态。

### 8.4 组合流程

以发帖为例：

```text
用户发帖
→ MySQL 写入 posts
→ outbox_events 写入 post.created
→ Outbox Publisher 发布 post.created 到 RabbitMQ
→ AI Event Consumer 收到 post.created
→ 创建 Asynq 任务 tag_post
→ tag_post 完成后创建 decide_ai_reply 任务
→ decide_ai_reply 完成后创建 generate_ai_reply 任务
→ generate_ai_reply 完成后写入 comments
→ 发布 ai.reply.completed 事件
```

### 8.5 为什么不只用 RabbitMQ

只用 RabbitMQ 可以完成异步消费，但项目还需要：

1. 任务级别重试。
2. 延迟任务。
3. 任务去重。
4. 任务状态面板。
5. AI 任务可视化调试。

这些更适合由 Asynq 承担。

### 8.6 为什么不只用 Asynq

只用 Asynq 会让所有异步行为都变成任务调用，不利于表达领域事件。

RabbitMQ 更适合表达：

```text
系统中发生了什么事件
```

Asynq 更适合表达：

```text
系统接下来要执行什么任务
```

---

## 9. Outbox Pattern 设计

### 9.1 设计原因

发帖时需要同时完成两件事：

```text
写入 MySQL 帖子数据
发布 post.created 消息到 RabbitMQ
```

如果 MySQL 写入成功，但 RabbitMQ 发布失败，就会出现数据不一致：

```text
帖子已经创建，但 AI 不会回复，搜索也不会同步。
```

因此使用 Outbox Pattern。

### 9.2 发帖事务

```text
BEGIN
  INSERT INTO posts ...
  INSERT INTO outbox_events ...
COMMIT
```

### 9.3 Outbox Publisher

独立进程扫描 `outbox_events` 表：

```text
读取 PENDING 事件
→ 发布到 RabbitMQ
→ 发布成功后标记为 PUBLISHED
→ 发布失败则增加 retry_count
```

MySQL 环境中采用定时轮询，推荐间隔：

```text
500ms ~ 1s
```

### 9.4 outbox_events 表字段

```text
id
event_id
event_type
aggregate_type
aggregate_id
payload
status
retry_count
created_at
published_at
```

---

## 10. 幂等性与重试

### 10.1 幂等要求

RabbitMQ 消息可能重复投递，Asynq 任务也可能因为失败重试再次执行，因此 Worker 必须支持幂等消费。

### 10.2 幂等方式

1. 每个事件拥有全局唯一 event_id。
2. 消费记录写入 processed_events 表。
3. AI 回复任务增加唯一约束。
4. 搜索索引使用固定 document id。
5. 通知生成使用业务唯一键防重。
6. processed_events 表保留 processed_at 索引，并定期清理历史消费记录。

### 10.3 AI 回复任务唯一约束

```text
post_id + parent_comment_id + ai_agent_id + trigger_type
```

同一个 AI 不应因为重复消息在同一帖子下生成多条相同类型回复。

### 10.4 重试规则

| 错误类型 | 处理方式 |
|---|---|
| 模型超时 | 重试 |
| 网络错误 | 重试 |
| RabbitMQ 临时失败 | 重试 |
| 内容审核失败 | 不重试，状态 BLOCKED |
| 参数错误 | 不重试，状态 FAILED |
| AI 被禁用 | 不重试，状态 SKIPPED |

### 10.5 死信队列

任务失败超过最大重试次数后进入死信队列。后台管理员可以查看死信消息并手动重试。

---

## 11. Worker 并发控制

### 11.1 需求说明

AI 任务属于外部 IO 密集型任务，单线程消费会导致队列阻塞；但无限制 goroutine 又容易导致 API 限流、内存膨胀和任务雪崩。

因此 Worker 必须支持并发控制。

### 11.2 选型

优先使用 Asynq 自带并发配置。

如果某些 RabbitMQ Consumer 需要自己管理 goroutine，则使用：

```text
ants
```

### 11.3 并发配置

```yaml
worker:
  ai_reply_concurrency: 4
  tagging_concurrency: 2
  search_index_concurrency: 2
  notification_concurrency: 4
```

### 11.4 并发原则

1. AI 生成任务并发数不能太高。
2. 搜索索引同步可以中等并发。
3. 通知任务可以较高并发。
4. 审核任务根据模型接口限制设置。
5. 所有并发参数必须可配置。

---

## 12. AI API 限流

### 12.1 需求说明

AI Worker 调用外部模型接口时必须进行限流，避免并发请求过多导致 429 或服务拒绝。

### 12.2 选型

使用：

```text
golang.org/x/time/rate
```

### 12.3 限流层级

至少需要两层限流：

```text
全局 AI Provider 限流
单个用户触发 AI 任务限流
```

### 12.4 示例配置

```yaml
ai:
  max_concurrency: 4
  request_per_second: 2
  burst: 2
```

### 12.5 限流规则

1. 所有 AI 模型调用必须经过统一 AIModelService。
2. AIModelService 内部维护 provider 级别 rate limiter。
3. Worker 获取任务后，在调用模型前等待限流器许可。
4. 如果等待超时，任务进入重试。
5. 用户恶意频繁 `@AI` 时，通过 Redis 做用户级限流。

---

## 13. Redis 使用场景

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

### 13.1 限流示例

```text
同一个用户每分钟最多 @AI 5 次。
同一个用户每分钟最多发布评论 10 条。
同一个用户每天最多发布帖子 30 条。
```

---

## 14. 数据库迁移

### 14.1 需求说明

项目必须引入数据库迁移工具，不能依赖手动执行 SQL 文件。

数据库迁移用于管理：

1. 表结构创建。
2. 表字段新增。
3. 索引新增。
4. 表结构回滚。
5. 本地开发环境重建。
6. 多人协作时数据库结构同步。
7. Docker Compose 初始化数据库。

### 14.2 选型

使用：

```text
golang-migrate
```

### 14.3 目录设计

```text
backend/
├── migrations/
│   ├── 000001_create_users_table.up.sql
│   ├── 000001_create_users_table.down.sql
│   ├── 000002_create_posts_table.up.sql
│   ├── 000002_create_posts_table.down.sql
│   └── ...
```

### 14.4 使用方式

开发环境执行：

```bash
migrate -path ./migrations -database "mysql://user:password@tcp(localhost:3306)/ai_forum" up
```

回滚：

```bash
migrate -path ./migrations -database "mysql://user:password@tcp(localhost:3306)/ai_forum" down 1
```

### 14.5 迁移要求

1. 每次数据库结构变更必须新增 migration。
2. 禁止直接手动修改线上数据库结构。
3. 每个 `.up.sql` 必须配套 `.down.sql`。
4. Docker Compose 初始化环境时必须能够自动执行迁移。
5. 后端启动前应检查数据库迁移状态。

---

## 15. 配置管理

### 15.1 需求说明

项目中所有环境相关配置必须从配置文件或环境变量读取，禁止硬编码。

需要管理的配置包括：

1. HTTP 服务端口。
2. MySQL 连接信息。
3. Redis 连接信息。
4. RabbitMQ 连接信息。
5. Elasticsearch 连接信息。
6. JWT 密钥和过期时间。
7. AI API Key。
8. AI 模型名称。
9. Worker 并发数。
10. AI 任务限流参数。
11. 日志级别。
12. CORS 配置。

### 15.2 选型

使用：

```text
Viper
```

### 15.3 配置文件结构

```text
backend/
├── config/
│   ├── config.dev.yaml
│   ├── config.prod.yaml
│   └── config.test.yaml
```

### 15.4 配置示例

```yaml
server:
  port: 8080
  mode: debug

mysql:
  host: localhost
  port: 3306
  username: root
  password: root
  database: ai_forum

redis:
  addr: localhost:6379
  password: ""
  db: 0

rabbitmq:
  url: amqp://guest:guest@localhost:5672/

elasticsearch:
  addresses:
    - http://localhost:9200
  username: ""
  password: ""

jwt:
  secret: "replace-this-in-prod"
  expire_hours: 168

ai:
  provider: openai
  model: gpt-4o-mini
  api_key: ${AI_API_KEY}
  max_concurrency: 4
  request_per_second: 2

worker:
  ai_reply_concurrency: 4
  tagging_concurrency: 2
  search_index_concurrency: 2

log:
  level: info
  encoding: json
```

### 15.5 配置优先级

配置读取优先级：

```text
环境变量 > 指定配置文件 > 默认配置
```

生产环境中，敏感信息必须通过环境变量注入。

---

## 16. 结构化日志

### 16.1 需求说明

系统需要引入结构化日志，便于排查接口错误、AI 任务失败、消息消费异常和搜索同步问题。

### 16.2 选型

使用：

```text
zap
```

### 16.3 日志字段要求

核心日志至少包含：

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

### 16.4 API 请求日志

每次 HTTP 请求需要记录：

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

### 16.5 Worker 日志

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

### 16.6 日志目标

开发环境：

```text
控制台可读日志
```

生产环境：

```text
JSON 格式日志
```

---

## 17. 优雅关机

### 17.1 需求说明

所有 Go 进程必须支持优雅关机。

涉及进程：

```text
api-server
worker-service
outbox-publisher
```

### 17.2 API Server 优雅关机

当收到 `SIGTERM` 或 `SIGINT` 时：

1. 停止接收新的 HTTP 请求。
2. 等待正在处理的请求完成。
3. 超时后强制退出。
4. 关闭 MySQL、Redis、RabbitMQ 连接。

### 17.3 Worker 优雅关机

当 Worker 收到 `SIGTERM` 或 `SIGINT` 时：

1. 停止拉取新任务。
2. 当前正在执行的任务继续执行。
3. 当前任务执行完成后确认消息。
4. 如果超过最大等待时间仍未完成，则记录日志并退出。
5. 未完成任务不能直接丢失，必须依靠任务状态和重试机制恢复。

### 17.4 Outbox Publisher 优雅关机

当 Outbox Publisher 收到退出信号时：

1. 停止扫描新的 outbox 事件。
2. 当前正在发布的事件完成后再退出。
3. 发布失败的事件保持 PENDING 或 FAILED 状态，等待下次重试。

---

## 18. 前端需求

### 18.1 用户侧页面

#### 18.1.1 页面列表

```text
/login
/register
/
/posts
/posts/:id
/create-post
/search
/ai-agents
/ai-agents/:id
/users/:id
/me
/notifications
```

#### 18.1.2 首页帖子流

功能：

1. 展示帖子列表。
2. 支持最新、最热、待回复、AI 参与最多排序。
3. 支持分类筛选。
4. 支持标签筛选。
5. 支持无限滚动。
6. 使用 React Virtuoso 优化长列表性能。

#### 18.1.3 帖子详情页

展示内容：

1. 帖子标题。
2. 作者信息。
3. 发布时间。
4. 分类和标签。
5. 正文内容。
6. 点赞、收藏、评论数、浏览量。
7. AI 参与状态。
8. 评论区。
9. 邀请 AI 回复按钮。
10. AI 回复任务状态。
11. SSE 实时接收 AI 回复完成事件。

#### 18.1.4 评论区

功能：

1. 展示一级评论。
2. 展示二级回复。
3. 支持回复评论。
4. 支持点赞评论。
5. 支持删除自己的评论。
6. 支持 `@AI`。
7. AI 评论需要明显标识 AI 身份。
8. AI 评论展示角色标签，如“现实主义者 · AI”。
9. 使用虚拟列表优化长评论区。

#### 18.1.5 发帖页

功能：

1. 输入标题。
2. 输入正文。
3. 选择分类。
4. 添加标签，可选。
5. 选择 AI 参与模式。
6. 发布帖子。
7. 发布成功后跳转帖子详情页。

#### 18.1.6 AI 角色广场

功能：

1. 展示所有启用 AI。
2. 展示 AI 名称、头像、简介。
3. 展示 AI 年龄视角、性格、价值倾向。
4. 展示 AI 擅长话题。
5. 查看 AI 最近参与的帖子。
6. 用户可以在帖子中 @AI。

#### 18.1.7 搜索页

功能：

1. 搜索帖子。
2. 搜索评论。
3. 搜索 AI 回复。
4. 按类型筛选结果。
5. 支持高亮关键词。
6. 支持分页。

---

### 18.2 管理后台页面

后台基于 Refine + Ant Design。

#### 18.2.1 页面列表

```text
/admin/login
/admin/dashboard
/admin/users
/admin/posts
/admin/comments
/admin/ai-agents
/admin/ai-tag-preferences
/admin/ai-tasks
/admin/ai-decisions
/admin/reports
/admin/settings
```

---

## 19. 数据库设计初稿

### 19.1 users

```sql
CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    nickname VARCHAR(50),
    avatar VARCHAR(255),
    bio VARCHAR(255),
    role VARCHAR(30) DEFAULT 'USER',
    status VARCHAR(30) DEFAULT 'NORMAL',
    created_at DATETIME,
    updated_at DATETIME
);
```

### 19.2 posts

```sql
CREATE TABLE posts (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    category VARCHAR(50),
    status VARCHAR(30) DEFAULT 'NORMAL',
    ai_mode VARCHAR(30) DEFAULT 'STANDARD',
    view_count INT DEFAULT 0,
    like_count INT DEFAULT 0,
    comment_count INT DEFAULT 0,
    ai_reply_count INT DEFAULT 0,
    hot_score DOUBLE DEFAULT 0,
    created_at DATETIME,
    updated_at DATETIME,
    INDEX idx_posts_user_id (user_id),
    INDEX idx_posts_status_created_at (status, created_at),
    INDEX idx_posts_hot_score (hot_score)
);
```

### 19.3 comments

```sql
CREATE TABLE comments (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    post_id BIGINT NOT NULL,
    parent_id BIGINT,
    root_id BIGINT,
    user_id BIGINT,
    ai_agent_id BIGINT,
    comment_type VARCHAR(20) NOT NULL,
    trigger_type VARCHAR(30),
    content TEXT NOT NULL,
    like_count INT DEFAULT 0,
    status VARCHAR(30) DEFAULT 'NORMAL',
    created_at DATETIME,
    updated_at DATETIME,
    INDEX idx_comments_post_id (post_id),
    INDEX idx_comments_root_id (root_id),
    INDEX idx_comments_parent_id (parent_id),
    INDEX idx_comments_ai_agent_id (ai_agent_id)
);
```

### 19.4 post_tags

```sql
CREATE TABLE post_tags (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    post_id BIGINT NOT NULL,
    tag_type VARCHAR(30) NOT NULL,
    tag_name VARCHAR(50) NOT NULL,
    confidence DECIMAL(5,4) DEFAULT 1.0000,
    source VARCHAR(20) DEFAULT 'AI',
    created_at DATETIME,
    INDEX idx_post_tags_post_id (post_id),
    INDEX idx_post_tags_type_name (tag_type, tag_name)
);
```

### 19.5 ai_agents

```sql
CREATE TABLE ai_agents (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(50) NOT NULL,
    avatar VARCHAR(255),
    description VARCHAR(255),
    age_viewpoint VARCHAR(50),
    personality VARCHAR(255),
    value_orientation VARCHAR(255),
    speaking_style VARCHAR(255),
    system_prompt TEXT,
    style_prompt TEXT,
    reply_threshold DECIMAL(5,4) DEFAULT 0.6000,
    activity_level DECIMAL(5,4) DEFAULT 0.5000,
    allow_auto_reply BOOLEAN DEFAULT TRUE,
    allow_mention_reply BOOLEAN DEFAULT TRUE,
    allow_followup_reply BOOLEAN DEFAULT TRUE,
    max_auto_replies_per_post INT DEFAULT 1,
    max_followup_replies_per_post INT DEFAULT 3,
    is_fallback BOOLEAN DEFAULT FALSE,
    active BOOLEAN DEFAULT TRUE,
    created_at DATETIME,
    updated_at DATETIME
);
```

### 19.6 ai_agent_tag_preferences

```sql
CREATE TABLE ai_agent_tag_preferences (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    ai_agent_id BIGINT NOT NULL,
    tag_type VARCHAR(30) NOT NULL,
    tag_name VARCHAR(50) NOT NULL,
    weight DECIMAL(5,4) NOT NULL,
    created_at DATETIME,
    updated_at DATETIME,
    UNIQUE KEY uk_agent_tag (ai_agent_id, tag_type, tag_name)
);
```

### 19.7 ai_reply_decisions

```sql
CREATE TABLE ai_reply_decisions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    post_id BIGINT NOT NULL,
    comment_id BIGINT,
    ai_agent_id BIGINT NOT NULL,
    trigger_type VARCHAR(30) NOT NULL,
    willingness_score DECIMAL(5,4),
    threshold_value DECIMAL(5,4),
    decision VARCHAR(30) NOT NULL,
    reason VARCHAR(255),
    created_at DATETIME,
    INDEX idx_ai_reply_decisions_post_id (post_id),
    INDEX idx_ai_reply_decisions_agent_id (ai_agent_id)
);
```

### 19.8 ai_reply_tasks

```sql
CREATE TABLE ai_reply_tasks (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    post_id BIGINT NOT NULL,
    parent_comment_id BIGINT,
    parent_comment_id_norm BIGINT
        GENERATED ALWAYS AS (COALESCE(parent_comment_id, 0)) STORED,
    target_comment_id BIGINT,
    ai_agent_id BIGINT NOT NULL,
    trigger_type VARCHAR(30) NOT NULL,
    status VARCHAR(30) DEFAULT 'PENDING',
    prompt TEXT,
    result TEXT,
    error_message TEXT,
    retry_count INT DEFAULT 0,
    created_at DATETIME,
    started_at DATETIME,
    finished_at DATETIME,
    UNIQUE KEY uk_ai_reply_task (
        post_id,
        parent_comment_id_norm,
        ai_agent_id,
        trigger_type
    )
);
```

说明：

```text
MySQL 中 NULL != NULL。
如果直接用 parent_comment_id 参与唯一索引，多个 parent_comment_id 为 NULL 的 POST_AUTO 任务不会被唯一约束拦住。
因此使用生成列 parent_comment_id_norm，将 NULL 统一映射为 0，再参与唯一索引。
```
```

### 19.9 comment_mentions

```sql
CREATE TABLE comment_mentions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    comment_id BIGINT NOT NULL,
    mention_type VARCHAR(30) NOT NULL,
    target_id BIGINT NOT NULL,
    display_name VARCHAR(100),
    created_at DATETIME,
    INDEX idx_comment_mentions_comment_id (comment_id),
    INDEX idx_comment_mentions_target (mention_type, target_id)
);
```

### 19.10 outbox_events

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

### 19.11 processed_events

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

说明：

```text
processed_events 用于幂等消费记录。
该表会持续增长，需要定期清理。
第一版保留最近 30 天记录，每天由定时任务删除 30 天前的数据。
```
```

### 19.12 notifications

```sql
CREATE TABLE notifications (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(200),
    content VARCHAR(500),
    related_type VARCHAR(50),
    related_id BIGINT,
    is_read BOOLEAN DEFAULT FALSE,
    created_at DATETIME,
    INDEX idx_notifications_user_id (user_id),
    INDEX idx_notifications_user_read (user_id, is_read)
);
```

---

## 20. 接口设计初稿

### 20.1 认证接口

```http
POST /api/auth/register
POST /api/auth/login
POST /api/auth/logout
GET  /api/users/me
```

### 20.2 帖子接口

```http
POST   /api/posts
GET    /api/posts
GET    /api/posts/{id}
PUT    /api/posts/{id}
DELETE /api/posts/{id}
POST   /api/posts/{id}/like
POST   /api/posts/{id}/favorite
GET    /api/posts/{id}/tags
GET    /api/posts/{id}/ai-status
GET    /api/posts/{id}/events
```

### 20.3 评论接口

```http
POST   /api/posts/{postId}/comments
GET    /api/posts/{postId}/comments
GET    /api/comments/{commentId}/replies
DELETE /api/comments/{commentId}
POST   /api/comments/{commentId}/like
```

### 20.4 AI 接口

```http
GET  /api/ai/agents
GET  /api/ai/agents/{id}
POST /api/posts/{postId}/ai-invitations
GET  /api/ai/tasks/{taskId}
```

### 20.5 搜索接口

```http
GET /api/search?q=xxx&type=post
```

### 20.6 通知接口

```http
GET  /api/notifications
PUT  /api/notifications/{id}/read
PUT  /api/notifications/read-all
```

### 20.7 后台接口

```http
GET  /api/admin/users
PUT  /api/admin/users/{id}/status
PUT  /api/admin/users/{id}/role

GET  /api/admin/posts
PUT  /api/admin/posts/{id}/status

GET  /api/admin/comments
PUT  /api/admin/comments/{id}/status

GET  /api/admin/ai-agents
POST /api/admin/ai-agents
PUT  /api/admin/ai-agents/{id}
PUT  /api/admin/ai-agents/{id}/status

GET    /api/admin/ai-tag-preferences
POST   /api/admin/ai-tag-preferences
PUT    /api/admin/ai-tag-preferences/{id}
DELETE /api/admin/ai-tag-preferences/{id}

GET  /api/admin/ai-tasks
GET  /api/admin/ai-tasks/{id}
POST /api/admin/ai-tasks/{id}/retry

GET  /api/admin/ai-decisions
GET  /api/admin/ai-decisions?postId=1001
```

---

## 21. 非功能需求

### 21.1 性能需求

1. 首页帖子列表必须分页或无限滚动。
2. 评论区必须分页加载。
3. 长评论区使用虚拟列表。
4. 热门帖子列表使用 Redis 缓存。
5. 搜索查询通过 Elasticsearch 完成。
6. AI 回复异步生成，不阻塞发帖接口。
7. Worker 必须支持并发控制和限流。

### 21.2 可用性需求

1. RabbitMQ 暂时不可用时，发帖仍然能写入 MySQL。
2. Outbox 事件保留，RabbitMQ 恢复后继续发布。
3. Elasticsearch 不可用时，不影响发帖、评论、详情页。
4. AI 服务失败时，不影响普通论坛功能。
5. AI 任务失败后进入可重试状态。
6. Worker 重启时不能丢失已领取但未完成的任务。

### 21.3 安全需求

1. 密码必须加密存储。
2. API 使用 JWT 鉴权。
3. 后台接口必须进行 RBAC 权限校验。
4. 用户输入内容需要防 XSS。
5. AI 输出必须经过审核。
6. 评论和帖子删除需要权限校验。
7. 用户操作需要限流。
8. 管理员操作写入审计日志。
9. 生产环境敏感配置通过环境变量注入。

### 21.4 可维护性需求

1. 后端按照领域模块拆分。
2. AI 相关逻辑不能直接写在 PostService 中。
3. RabbitMQ 消息需要统一定义。
4. Worker 任务需要支持幂等。
5. 数据库迁移脚本需要统一管理。
6. 接口需要维护 Swagger 文档。
7. 配置必须集中管理。
8. 日志必须结构化。

---

## 22. 容器化部署需求

### 22.1 Docker Compose 服务

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

### 22.2 开发环境启动方式

```bash
docker compose up -d mysql redis rabbitmq elasticsearch
go run cmd/api/main.go
go run cmd/worker/main.go
go run cmd/outbox-publisher/main.go
pnpm dev
```

### 22.3 部署环境启动方式

```bash
docker compose up -d
```

---

## 23. 项目目录建议

```text
ai-forum/
├── backend/
│   ├── cmd/
│   │   ├── api/
│   │   ├── worker/
│   │   └── outbox-publisher/
│   ├── internal/
│   │   ├── auth/
│   │   ├── user/
│   │   ├── forum/
│   │   │   ├── post/
│   │   │   ├── comment/
│   │   │   ├── tag/
│   │   │   └── like/
│   │   ├── ai/
│   │   │   ├── agent/
│   │   │   ├── tagging/
│   │   │   ├── decision/
│   │   │   └── reply/
│   │   ├── search/
│   │   ├── notification/
│   │   ├── moderation/
│   │   ├── event/
│   │   ├── outbox/
│   │   ├── mq/
│   │   ├── task/
│   │   ├── cache/
│   │   ├── rbac/
│   │   ├── config/
│   │   ├── logger/
│   │   └── common/
│   ├── migrations/
│   ├── config/
│   └── docs/
├── web/
├── admin/
└── docker-compose.yml
```

---

## 24. 开发阶段规划

### 24.1 第一阶段：论坛基础功能

目标：完成普通论坛主干。

功能：

1. 用户注册登录。
2. 发帖。
3. 帖子列表。
4. 帖子详情。
5. 评论。
6. 评论树。
7. 点赞。
8. 分类。
9. 标签。
10. 基础后台管理。

### 24.2 第二阶段：AI 角色系统

目标：完成 AI 角色配置和 AI 自动回复。

功能：

1. AI 角色表。
2. AI 角色后台管理。
3. AI 标签偏好配置。
4. 帖子自动打标签。
5. 回答意愿分计算。
6. AI 自动回复任务。
7. AI 评论写入评论区。
8. 保底机制。

### 24.3 第三阶段：AI 互动增强

目标：完成 `@AI` 和追问机制。

功能：

1. 前端 `@AI` 输入。
2. comment_mentions 表。
3. MENTION 类型任务。
4. FOLLOWUP 类型任务。
5. AI 回复限制。
6. AI 任务状态展示。

### 24.4 第四阶段：分布式增强

目标：体现分布式系统能力。

功能：

1. RabbitMQ 领域事件。
2. Outbox Pattern。
3. Asynq 任务调度。
4. Elasticsearch + IK 中文搜索。
5. Redis 缓存和限流。
6. 死信队列。
7. 幂等消费。
8. Docker Compose 部署。
9. SSE 推送 AI 回复。

### 24.5 第五阶段：后台治理与展示

目标：完成项目展示亮点。

功能：

1. AI 决策日志后台。
2. AI 任务后台。
3. AI 标签偏好后台。
4. 用户封禁。
5. 帖子审核。
6. 评论审核。
7. 审计日志。
8. 系统配置。
9. AI 决策日志可视化。

---

## 25. 验收标准

### 25.1 基础论坛验收

1. 用户可以注册登录。
2. 用户可以发布帖子。
3. 用户可以查看帖子列表。
4. 用户可以查看帖子详情。
5. 用户可以评论帖子。
6. 用户可以回复评论。
7. 用户可以点赞帖子和评论。
8. 管理员可以管理帖子和评论。

### 25.2 AI 功能验收

1. 用户发帖后系统自动生成标签。
2. 系统根据标签计算 AI 回答意愿分。
3. 系统自动选择 1~6 个 AI 参与回复。
4. 没有 AI 匹配时触发保底机制。
5. AI 回复作为评论展示。
6. AI 回复带有 AI 角色标识。
7. 用户可以 `@AI`。
8. 用户可以在 AI 回复下继续追问。
9. AI 可以根据上下文继续回答。
10. 后台可以查看 AI 决策日志。
11. 后台可以查看 AI 决策日志可视化图表。

### 25.3 分布式架构验收

1. 发帖后通过 Outbox 生成事件。
2. Outbox Publisher 可以发布事件到 RabbitMQ。
3. Event Consumer 可以消费事件并创建 Asynq 任务。
4. Worker 可以执行 Asynq 任务。
5. AI 回复任务异步生成。
6. 搜索索引通过异步任务同步到 Elasticsearch。
7. Redis 可以完成限流和缓存。
8. Worker 消费支持幂等。
9. 失败消息可以进入死信队列。
10. Docker Compose 可以启动完整环境。
11. Elasticsearch 支持中文分词。
12. Worker 支持优雅关机。
13. 服务使用结构化日志。
14. 配置不硬编码。

---

## 26. 暂不优先实现

第一版暂不优先实现：

```text
复杂私信系统
复杂关注系统
积分等级系统
推荐算法
复杂 AI 辩论模式
移动端 App
```

### 26.1 私信模块说明

私信系统不属于本项目核心亮点，第一版不实现。

原因：

1. 私信是独立业务。
2. 与多 AI 角色论坛主线关系弱。
3. 会增加消息会话、未读数、隐私权限、举报等额外复杂度。
4. 不如优先完成 AI 决策日志、AI 回复任务和分布式异步流程。

后续可作为扩展模块预留。

---

## 27. 项目亮点总结

本项目的主要亮点包括：

1. 多 AI 角色参与式论坛。
2. AI 角色具有不同性格、年龄视角和价值倾向。
3. 基于帖子标签和 AI 偏好的回答意愿分机制。
4. 自动回复、`@AI`、追问三种 AI 参与方式。
5. 保底回复机制，避免帖子冷场。
6. AI 决策日志可解释，后台可以查看 AI 为什么回复或跳过。
7. AI 决策日志可视化，展示意愿分、阈值、命中标签和最终决策。
8. RabbitMQ 异步解耦领域事件。
9. Asynq 调度具体 AI 任务。
10. Elasticsearch + IK 构建中文搜索读模型。
11. Redis 实现热点缓存和频率限制。
12. Outbox Pattern 解决数据库写入和消息发布一致性问题。
13. Docker Compose 完成多组件容器化部署。
14. React Virtuoso 优化帖子流和评论区长列表性能。
15. Refine + Ant Design 快速构建后台管理系统。
16. zap 结构化日志提升问题排查能力。
17. Viper 统一配置管理。
18. golang-migrate 管理数据库迁移。
19. Worker 支持并发控制、限流和优雅关机。

---

## 28. 当前版本范围

第一版建议优先实现以下内容：

```text
用户注册登录
发帖
帖子列表
帖子详情
评论区
点赞
AI 角色表
AI 标签偏好表
帖子打标签
回答意愿分计算
AI 自动回复
@AI
AI 回复下追问
AI 决策日志可视化
RabbitMQ + Outbox 领域事件
Asynq AI 任务调度
Redis 限流和热度缓存
Elasticsearch + IK 中文搜索
SSE 推送 AI 回复
Refine 后台管理
Docker Compose 部署
```

第一版的核心验收逻辑：

```text
用户发帖
→ 系统自动打标签
→ 系统计算每个 AI 的回答意愿分
→ 后台可视化展示决策日志
→ 系统选择合适 AI 异步回复
→ 前端通过 SSE 收到 AI 回复完成事件
→ AI 回复作为评论显示在帖子详情页
```
