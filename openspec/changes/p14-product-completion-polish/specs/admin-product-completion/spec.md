## ADDED Requirements

### Requirement: Admin dashboard uses live data
In real mode, admin dashboard metrics, trends, service cards, recent posts/tasks, and decision timeline SHALL be fetched from backend endpoints or derived from live backend data. Mock dashboard data SHALL NOT be used in real mode.

#### Scenario: Empty database shows empty dashboard
- **WHEN** the backend has no posts, tasks, or decision logs
- **THEN** the dashboard shows zero/empty states rather than mock sample activity

#### Scenario: New task appears in dashboard
- **WHEN** a real AI task exists
- **THEN** the dashboard task summary and recent task list reflect that task

### Requirement: Visible admin operations are backed or hidden
Admin create, update, delete, retry, terminate, and mark-processed actions SHALL either call a backend endpoint and surface success/failure, or be hidden/disabled with an explicit read-only state. Visible actions SHALL NOT fail with client-side `not implemented` errors.

#### Scenario: Unsupported delete is not shown
- **WHEN** a resource has no backend delete endpoint
- **THEN** the admin UI does not show a delete action for that resource

#### Scenario: Agent edit persists
- **WHEN** an admin updates an AI agent threshold
- **THEN** the backend persists the value and a reload shows the saved value

#### Scenario: Supported task/post actions are live
- **WHEN** an admin retries/terminates/mark-processed an AI task, or updates a post status
- **THEN** the action calls the existing backend endpoint and the list/record reflects the change without a client-side `not implemented` error

### Requirement: Admin reports are not half-wired
The admin app SHALL NOT expose `/admin/reports` navigation or routes unless an owning OpenSpec phase implements the backend report workflow.

#### Scenario: Reports route absent
- **WHEN** reports are out of scope
- **THEN** no reports menu item or functional-looking reports route is visible
