## ADDED Requirements

### Requirement: AI agent profile and tag preferences
The `ai_agents` table SHALL store each agent's enabled state, reply threshold, activity level, and trigger-type permissions (`allowAutoReply`, `allowMention`, `allowFollowup`). The `ai_agent_tag_preferences` table SHALL store per-agent tag-type/tag-name weights (0.0–1.0).

#### Scenario: Enabled agents with preferences are readable
- **WHEN** the decision handler reads enabled agents
- **THEN** each returned agent carries its tag preferences and threshold

### Requirement: Dev AI seed data
After creating `ai_agents` and `ai_agent_tag_preferences`, the P6 migrations SHALL insert dev-only AI agent rows and tag preferences matching the design docs. The seed SHALL be reversible and SHALL remove only the seeded AI rows by fixed IDs/names.

#### Scenario: Seed enables AI decision testing
- **WHEN** `make migrate-up` completes through P6 in dev mode
- **THEN** the database contains at least 3 enabled AI agents, thresholds/activity defaults, trigger permissions, and tag preferences
