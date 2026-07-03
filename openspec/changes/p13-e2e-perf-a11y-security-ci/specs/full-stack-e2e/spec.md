## ADDED Requirements

### Requirement: Full AI reply chain integration
A Playwright integration spec SHALL verify, against the live 3-process stack: user creates a post in web → backend scores willingness → decision logged → worker emits the AI reply → web SSE shows the reply live → admin decision-log shows the willingness/threshold/hit-tags breakdown for that decision.

#### Scenario: End-to-end AI chain is green
- **WHEN** the integration spec runs against the docker-compose stack
- **THEN** the post, AI reply, SSE delivery, and admin decision-log breakdown all succeed in one flow

### Requirement: Sanity against live stack
A sanity spec SHALL assert both web (5173) and admin (5174) return 200 with a visible h1 against the live backend with no mock fallback (`VITE_API_MODE=real`).

#### Scenario: Both apps reachable live
- **WHEN** the sanity spec runs
- **THEN** web and admin respond 200 with a visible h1 and no mock-fallback indicator

### Requirement: Internal API not publicly reachable
A spec SHALL assert `/internal/**` returns 404 through the public proxy and that api-server is not directly host-accessible.

#### Scenario: Internal path blocked publicly
- **WHEN** a public request hits `/internal/posts/1/events`
- **THEN** it returns 404 and is not forwarded to api-server

### Requirement: Notification read contract E2E
The live web app SHALL expose generated notifications, unread count, mark-one-read, and mark-all-read behavior against the real backend.

#### Scenario: Notification unread count changes
- **WHEN** a notification is generated for the current user
- **AND** the user marks it read or marks all read
- **THEN** the visible unread count updates without a page reload

### Requirement: Search rebuild entrypoint smoke
The documented search rebuild entrypoint SHALL be triggerable in the live stack and rebuild Elasticsearch from MySQL using the same document shape as incremental sync.

#### Scenario: Search rebuild restores documents
- **WHEN** the rebuild entrypoint runs after the ES index is cleared
- **THEN** searchable documents are restored from MySQL-backed source rows

### Requirement: Reports scope guard
The `/admin/reports` surface SHALL either be implemented by an explicit OpenSpec phase or documented as v1 out-of-scope. A visible reports route or menu without backend behavior is not allowed.

#### Scenario: Reports are not half-wired
- **WHEN** the admin app is inspected in real mode
- **THEN** reports are either functional per their owning phase or absent/documented out-of-scope
