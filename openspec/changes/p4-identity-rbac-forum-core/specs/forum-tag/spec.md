## ADDED Requirements

### Requirement: Post tag storage
The `post_tags` table SHALL store tags for a post, each with a tag type (`topic`/`intent`/`emotion`/`debate`/`risk`) and tag name. P4 provides storage and read; AI tag **generation** is performed by P6.

#### Scenario: Tags readable for a post
- **WHEN** a post's tags are requested after being stored
- **THEN** the tags are returned grouped by type
