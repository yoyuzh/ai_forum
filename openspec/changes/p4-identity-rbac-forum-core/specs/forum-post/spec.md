## ADDED Requirements

### Requirement: Post create writes row and outbox event in-tx
`PostService.CreatePost` SHALL, within a single transaction, insert a `posts` row and append exactly one `post.created` event to `outbox_events`. It SHALL NOT publish to RabbitMQ inside the transaction. It SHALL NOT contain AI, search, notification, or moderation logic.

#### Scenario: One outbox row shares the write's transaction
- **WHEN** a user creates a post
- **THEN** exactly one `outbox_events` row with `event_type='post.created'` exists for that post, and forcing the service to error after the outbox append rolls back both the post and the outbox row

#### Scenario: No RabbitMQ publish in-tx
- **WHEN** a post is created
- **THEN** the RabbitMQ queue depth for forum events remains 0 (no publish occurs at this phase)

### Requirement: PostService has no cross-domain logic
`PostService` SHALL NOT import `ai`, `search`, `notification`, or `moderation` packages.

#### Scenario: Import guard
- **WHEN** the `forum/post` package imports are inspected
- **THEN** none of `ai/`, `search/`, `notification/`, `moderation/` appear

### Requirement: Admin post moderation status emits event
Changing a post's moderation/status field through the backend admin route SHALL update the post row and append exactly one `post.moderated` outbox event in the same transaction. This route SHALL represent the state transition only; moderation policy execution is owned by the moderation module, not `PostService`.

#### Scenario: Admin hides a post
- **WHEN** an authorized admin changes a post status from `NORMAL` to `HIDDEN`
- **THEN** the post status is updated and a `post.moderated` outbox row is appended in-tx

### Requirement: Post read, update, delete
The system SHALL support reading a post by id, listing posts, updating own post, and deleting (own or any with permission). Updates and deletes SHALL append the corresponding `post.updated`/`post.deleted` outbox event in-tx.

#### Scenario: Owner updates own post
- **WHEN** the post owner updates their post
- **THEN** the post is updated and a `post.updated` outbox event is appended in-tx
