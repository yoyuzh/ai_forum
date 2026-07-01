## ADDED Requirements

### Requirement: Refine REST dataProvider and authProvider
The admin app SHALL use a Refine REST dataProvider targeting the backend and an authProvider wired to backend login/JWT. Data fetching SHALL go through the dataProvider or `admin/src/api`, not scattered in components.

#### Scenario: Admin login and CRUD
- **WHEN** an admin logs in and opens a resource screen
- **THEN** the resource list loads from the backend and CRUD actions persist

### Requirement: RBAC visibility is frontend-only
The admin SHALL use an `accessControlProvider` to hide buttons/routes based on backend permissions. Frontend checks control visibility only and are NOT security. A denied action SHALL return 403 from the backend even if the frontend would have shown it.

#### Scenario: Denied action returns 403
- **WHEN** an admin without `ai_task:retry` triggers a retry server-side
- **THEN** the backend returns 403

### Requirement: Operational screens are dense and operator-focused
Admin screens (Users, Posts, Comments, AI Agents, AI Tasks, Tags/Preferences) SHALL be operational and dense (operator-focused), not marketing-style. AI Agent screen SHALL surface `replyThreshold`, `activityLevel`, and trigger permissions with inline edit.

#### Scenario: Agent config editable inline
- **WHEN** an admin edits an agent's reply threshold
- **THEN** the change persists to the backend
