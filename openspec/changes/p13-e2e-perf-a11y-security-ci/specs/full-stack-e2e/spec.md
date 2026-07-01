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
