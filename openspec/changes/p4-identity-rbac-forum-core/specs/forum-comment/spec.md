## ADDED Requirements

### Requirement: Comment create writes row and outbox event in-tx
Creating a comment SHALL, in one transaction, insert a `comments` row, update `posts.comment_count`, and append a `comment.created` outbox event. Comments support `comment_type` (USER vs AI) — P4 only creates USER comments; AI comments are created by P7.

#### Scenario: User comment increments count and emits outbox
- **WHEN** a user creates a comment on a post
- **THEN** `posts.comment_count` is incremented and a `comment.created` outbox row is appended in the same transaction

### Requirement: Comment tree read avoids N+1
Reading a post's comments SHALL load all rows in a bounded number of queries (ideally one) and assemble the tree in memory (`comment/tree.go`).

#### Scenario: Single-query tree load
- **WHEN** a post with N comments is read
- **THEN** the comment tree is returned using at most one comment-list query regardless of N

### Requirement: Comment delete appends outbox event
Deleting a comment (soft-delete) SHALL append a `comment.deleted` outbox event in-tx and decrement `posts.comment_count`.

#### Scenario: Delete emits event
- **WHEN** a user deletes their comment
- **THEN** a `comment.deleted` outbox row is appended in-tx
