## ADDED Requirements

### Requirement: Domain event registry and envelope
The `event` package SHALL define constants for every architecture §8.5 event type: `post.created`, `post.updated`, `post.deleted`, `comment.created`, `comment.deleted`, `post.tagged`, `ai.reply.completed`, `ai.reply.failed`, `post.moderated`, `user.mentioned`. It SHALL provide an envelope builder producing the §7.5 shape (`eventId`, `eventType`, `aggregateType`, `aggregateId`, `occurredAt`, `payload`).

#### Scenario: Envelope shape
- **WHEN** an event is built for `post.created` with aggregateId 1001
- **THEN** the envelope contains a unique `eventId`, `eventType='post.created'`, `aggregateType='post'`, `aggregateId=1001`, an `occurredAt`, and a `payload` with only business IDs
