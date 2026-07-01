## ADDED Requirements

### Requirement: tag_post generates five tag types
The `tag_post` Asynq handler SHALL read the post, generate tags of types `topic`, `intent`, `emotion`, `debate`, and `risk`, write them to `post_tags`, and append a `post.tagged` outbox event. It SHALL be idempotent via `processed_events`.

#### Scenario: post.created triggers tagging
- **WHEN** a `post.created` event is consumed
- **THEN** a `tag_post` task runs, `post_tags` is populated with the five types, and a `post.tagged` outbox event is appended

#### Scenario: Redelivery does not duplicate tags
- **WHEN** the `post.created` event is redelivered
- **THEN** `processed_events` prevents a second `tag_post` execution
