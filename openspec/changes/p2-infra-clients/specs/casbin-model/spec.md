## ADDED Requirements

### Requirement: Casbin model definition
`internal/rbac/model.conf` SHALL define a Casbin model with request shape `r = sub, obj, act` (subject, object, action) suitable for the permission set in architecture §12.2 (e.g. `post:create`, `user:ban`, `ai_task:retry`). No enforcement or policy storage is implemented in this phase.

#### Scenario: Model loads
- **WHEN** the model file is loaded by Casbin
- **THEN** an enforcer can be constructed with a trivial in-memory policy and evaluated without error
