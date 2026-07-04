## ADDED Requirements

### Requirement: Web real mode exposes real AI data
In `VITE_API_MODE=real`, web routes that display AI agents, AI reply tasks, AI activity, or decision logs SHALL fetch live backend data instead of returning hard-coded empty arrays.

#### Scenario: Agents page uses backend data
- **WHEN** a user opens `/agents` in real mode
- **THEN** enabled AI agents from the backend are shown
- **AND** the page does not fall back to mock seed data

#### Scenario: Post detail shows real decision context
- **WHEN** a post has `decision_logs` and `ai_reply_tasks`
- **THEN** the post-detail AI sidebar shows the real decision/task state for that post

### Requirement: Profile edits persist
The profile page SHALL persist supported profile edits through the backend and refetch the saved state after update.

#### Scenario: Display name update survives reload
- **WHEN** a logged-in user changes their display name and reloads the page
- **THEN** the updated display name is read from the backend

### Requirement: User stats are backend-derived
The profile page SHALL show post/comment/like/AI-reply counts derived from backend data instead of fixed zeros.

#### Scenario: Stats reflect authored content
- **WHEN** a user has created posts and comments
- **THEN** the profile stats reflect those backend rows

### Requirement: No visible fake SSO
The web app SHALL NOT show SSO buttons or claims unless a real SSO backend flow exists.

#### Scenario: Login page has no fake provider buttons
- **WHEN** SSO is not implemented
- **THEN** the login page does not render provider buttons that cannot complete authentication
