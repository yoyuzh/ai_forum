## ADDED Requirements

### Requirement: Favorite and unfavorite
A user SHALL be able to favorite and unfavorite a post, with uniqueness on `(user_id, post_id)`.

#### Scenario: Favorite persists
- **WHEN** a user favorites a post
- **THEN** the favorite is persisted and visible in the user's favorites list
