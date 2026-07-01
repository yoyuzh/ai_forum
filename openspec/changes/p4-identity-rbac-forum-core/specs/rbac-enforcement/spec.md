## ADDED Requirements

### Requirement: Backend RBAC is authoritative
Authorization SHALL be enforced by the backend using Casbin. Frontend permission checks control visibility only and MUST NOT be treated as security. A denied action SHALL return 403 even if the frontend would have shown it.

#### Scenario: Denied action returns 403
- **WHEN** an authenticated user without `post:delete-any` calls `DELETE /api/posts/{id}` on another user's post
- **THEN** the backend returns 403

### Requirement: JWT middleware populates subject
JWT middleware SHALL validate the token and populate the authenticated user (subject) into the request context for downstream RBAC enforcement. Invalid/expired tokens return 401.

#### Scenario: Expired token rejected
- **WHEN** a request carries an expired JWT
- **THEN** the middleware returns 401

### Requirement: Public routes do not require JWT
The routes `GET /api/posts`, `GET /api/posts/{id}`, `GET /api/ai/agents`, and `GET /api/search` SHALL be accessible without authentication (§12.1).

#### Scenario: Guest reads post list
- **WHEN** an unauthenticated request calls `GET /api/posts`
- **THEN** it returns 200 with the post list
