## ADDED Requirements

### Requirement: Search worker re-fetches from MySQL
The `sync_search_index` handler SHALL re-fetch the full row from MySQL using the event's aggregate ID before assembling the ES document. It SHALL NOT trust the event payload for document content.

#### Scenario: Document built from current MySQL state
- **WHEN** a `post.created` event is consumed
- **THEN** the ES document is assembled from the freshly fetched MySQL row, not the event payload

### Requirement: Search reflects writes within lag budget
A post or comment SHALL be searchable in Elasticsearch within 1–3 seconds of the MySQL write (§6.7).

#### Scenario: Created post becomes searchable
- **WHEN** a post is created and ~3 seconds pass
- **THEN** the post is retrievable via Elasticsearch search

### Requirement: Deletes and moderation sync
`post.deleted` and `comment.deleted` SHALL delete the corresponding ES document. `post.moderated` SHALL update or remove the document per the moderation outcome. `ai.reply.failed` SHALL be consumed (acked/logged) without writing an ES document.

#### Scenario: Deleted post is removed from search
- **WHEN** a post is deleted
- **THEN** its ES document is removed and it is no longer searchable

### Requirement: ES outage does not block MySQL
If Elasticsearch is unavailable, the MySQL write path SHALL still succeed and return 200. The search sync task SHALL retry per §9.4 without affecting business writes.

#### Scenario: ES down, write still succeeds
- **WHEN** the ES container is killed and a user creates a post
- **THEN** the request returns 200 and the post is persisted in MySQL

### Requirement: Full rebuild equals incremental
A full re-index from MySQL SHALL produce documents identical to those produced by incremental sync for the same data.

#### Scenario: Rebuild matches incremental
- **WHEN** a full rebuild and an incremental sync run over the same dataset
- **THEN** the resulting ES documents are identical
