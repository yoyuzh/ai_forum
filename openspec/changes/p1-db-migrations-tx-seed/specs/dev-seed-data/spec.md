## ADDED Requirements

### Requirement: Dev admin seed data
Migration `000004_seed_dev` SHALL insert, for development only, an admin user with a bcrypt-hashed password. It SHALL NOT insert AI agent rows because `ai_agents` and `ai_agent_tag_preferences` are created later in P6. The migration SHALL be reversible: its `.down.sql` removes only the seeded admin row.

#### Scenario: Seed enables admin authentication testing
- **WHEN** `make migrate-up` completes in dev mode
- **THEN** the database contains one dev admin user and no AI seed rows are attempted

#### Scenario: Seed is reversible
- **WHEN** `make migrate-down` is run after `migrate-up`
- **THEN** the seeded admin user is removed without affecting any other rows

### Requirement: No real secrets in seed
The dev seed SHALL NOT contain real production credentials. The admin user's password SHALL be a known dev-only bcrypt hash, and no API keys or tokens SHALL appear in seed SQL.

#### Scenario: Seed contains no production secrets
- **WHEN** the seed migration SQL is inspected
- **THEN** it contains no literal values for `JWT_SECRET`, `INTERNAL_API_TOKEN`, or `AI_API_KEY`, and the admin password is a bcrypt hash, not plaintext
