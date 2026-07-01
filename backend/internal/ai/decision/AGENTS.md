# Module Instructions

## Responsibility

Own AI reply willingness score calculation and decision logging.

## Owns

- `ai_reply_decisions` writes.
- Decision score inputs, outputs, DTOs, repositories, and services.

## Must Not

- Do not call large language models to generate replies.
- Do not directly write comments.
- Do not enqueue unrelated tasks outside the decision boundary.

## Allowed Dependencies

- AI agent/preference read interfaces, forum post/tag read interfaces, database, common, logger, and task contracts.

## Communication Rules

- Receive post/tag/agent context through explicit interfaces.
- Return or persist decision outcomes for downstream reply task creation.
- The scoring formula comes from the requirements document and must stay explainable.

## Data Rules

- Every decision must write an `ai_reply_decisions` record.
- Store inputs and score components needed for admin explanation.

## Testing Rules

- Test score formula components, threshold behavior, skipped/replied decisions, and decision-log persistence.

## Notes for Codex

- Decision is scoring and logging only; generation belongs to `ai/reply`.
