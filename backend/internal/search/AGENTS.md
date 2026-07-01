# Module Instructions

## Responsibility

Own Elasticsearch indexing and search read-model queries.

## Owns

- Elasticsearch index mappings, sync handlers, search DTOs/services/repositories.

## Must Not

- Do not act as a source of truth for business decisions.
- Do not mutate MySQL forum/AI tables except through owning services.

## Allowed Dependencies

- `event`, `task`, database read interfaces, Elasticsearch client adapters.

## Communication Rules

- React to domain events and Asynq `sync_search_index` tasks.
- Consumers must be idempotent and tolerate eventual consistency.

## Data Rules

- Elasticsearch documents are rebuildable from MySQL.
- Store only search-optimized read-model data.

## Testing Rules

- Test idempotency, retry behavior, and boundary rules when handlers/adapters are implemented.

## Notes for Codex

- Keep infrastructure packages free of unrelated domain shortcuts.
