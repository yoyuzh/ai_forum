## ADDED Requirements

### Requirement: Like and unlike
A user SHALL be able to like and unlike a post. Liking SHALL increment the post's like count; unliking SHALL decrement it. A user cannot like the same post twice (uniqueness on `(user_id, post_id)`).

#### Scenario: Like then unlike
- **WHEN** a user likes a post and then unlikes it
- **THEN** the like count returns to its prior value and no duplicate like row exists

#### Scenario: Duplicate like rejected
- **WHEN** a user likes a post they already liked
- **THEN** the request is a no-op (or 409) and the count does not double
