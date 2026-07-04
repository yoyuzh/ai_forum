## ADDED Requirements

### Requirement: Web search uses backend search
The web feed search box SHALL query a backend search surface backed by Elasticsearch when a search term is present. Client-side filtering MAY remain as a mock-mode fallback only.

#### Scenario: Search returns ES-backed posts
- **WHEN** a user searches for a term contained in an indexed post
- **THEN** the results are returned by the backend search endpoint and match the Elasticsearch read model

### Requirement: Backend search query is a read-model operation
The `search` package SHALL provide a query method over the Elasticsearch post index. Query results SHALL be re-checked against MySQL for authorization and status before being returned to the caller. Elasticsearch is a read model only and SHALL NOT be used for authorization or business decisions.

#### Scenario: ES hit filtered out by MySQL status
- **WHEN** a post exists in the Elasticsearch index but has been deleted or hidden in MySQL
- **THEN** the backend search result excludes that post

### Requirement: Search outage does not break normal feed
If Elasticsearch is unavailable, search SHALL show an unavailable/error state, while normal feed loading from MySQL remains usable.

#### Scenario: Search unavailable but feed works
- **WHEN** Elasticsearch is down
- **AND** a user performs a search
- **THEN** the search UI shows a recoverable error
- **AND** clearing the search returns to the normal `/api/posts` feed

### Requirement: Search E2E verifies rebuild consistency
The E2E suite SHALL verify that a post rebuilt into Elasticsearch can be found from the web search UI.

#### Scenario: Rebuilt post is searchable
- **WHEN** the search rebuild entrypoint runs after posts exist in MySQL
- **THEN** the web search UI can find a rebuilt post by title/content
