## MODIFIED Requirements

### Requirement: Comment create writes row and outbox event in-tx
Creating a comment SHALL, in one transaction, insert a `comments` row, update `posts.comment_count`, and append a `comment.created` outbox event. Comments support `comment_type` (USER vs AI) — P4 only creates USER comments; AI comments are created by P7. When a USER comment mentions AI agents, the create path SHALL additionally write `comment_mentions`, enforce mention limits, and enqueue `generate_ai_reply`(MENTION) tasks. When a USER comment replies to an AI comment, the create path SHALL enqueue a `judge_ai_followup` task. The mention/followup task enqueuing happens after the comment transaction commits.

#### Scenario: User comment increments count and emits outbox
- **WHEN** a user creates a comment on a post
- **THEN** `posts.comment_count` is incremented and a `comment.created` outbox row is appended in the same transaction

#### Scenario: Mention comment enqueues MENTION tasks after commit
- **WHEN** a user comment mentions AI agents
- **THEN** after the comment transaction commits, `generate_ai_reply`(MENTION) tasks are enqueued for valid mentioned agents

#### Scenario: Reply to AI comment enqueues followup judge
- **WHEN** a real user replies to an AI comment
- **THEN** a `judge_ai_followup` task is enqueued after the comment commits
