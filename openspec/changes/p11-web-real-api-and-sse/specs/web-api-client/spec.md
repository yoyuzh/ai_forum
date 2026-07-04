## ADDED Requirements

### Requirement: Env-gated API client with shared contract
The web app SHALL select between a mock and a real API client via `VITE_API_MODE` (`mock` | `real`). Both clients SHALL implement the same function signatures defined in `types.ts`, so the contract cannot diverge. The real client SHALL target `VITE_API_BASE_URL`.

#### Scenario: Real mode round-trips
- **WHEN** `VITE_API_MODE=real` and a user loads the feed
- **THEN** posts are fetched from the live backend and persist on create

#### Scenario: Mock mode still works
- **WHEN** `VITE_API_MODE=mock`
- **THEN** the app behaves as before against the mock layer

### Requirement: Auth and error handling
The HTTP client SHALL redirect to login on 401, surface a permission message on 403, and surface a rate-limit message on 429. Auth state SHALL live in Zustand (client state only); server data SHALL be fetched via TanStack Query and not duplicated into Zustand.

#### Scenario: 401 redirects to login
- **WHEN** a request returns 401
- **THEN** the user is redirected to login and auth state is cleared

#### Scenario: 429 surfaced
- **WHEN** a request returns 429
- **THEN** a rate-limit message is shown to the user

### Requirement: Notification read contract
The web API client SHALL support listing notifications, reading an unread count, marking one notification read, and marking all notifications read in both mock and real modes.

#### Scenario: Unread badge updates after mark-read
- **WHEN** a user opens notifications with unread rows
- **AND** marks one notification read
- **THEN** the unread badge decreases without a page reload
