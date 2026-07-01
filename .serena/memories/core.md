# Core

- AI Forum is a modular monolith with three Go backend processes only: `api-server`, `worker-service`, `outbox-publisher`.
- Repository root contains product/architecture docs plus generated skeleton directories: `backend/`, `web/`, `admin/`, `docs/`, `deploy/`, `scripts/`, `tools/`.
- Read `mem:tech_stack` for selected technologies and `mem:conventions` for architecture boundaries before adding code.
- Backend domain boundaries are encoded in per-directory `AGENTS.md`; preserve them when implementing features.
- Current source-of-truth docs: `ai_forum_requirements_v2.md`, `ai_forum_architecture_v1.md`; do not move or overwrite them.