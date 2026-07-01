## ADDED Requirements

### Requirement: Willingness score formula
The `decide_ai_reply` handler SHALL compute each enabled agent's `willingness_score` as `topic_score*0.35 + intent_score*0.25 + emotion_score*0.15 + debate_score*0.15 + activity_score*0.10 - risk_penalty - frequency_penalty`, where each tag-type score is `max_score*0.7 + avg_score*0.3` (architecture §11.2).

#### Scenario: Score matches hand-computed fixture
- **WHEN** an agent's tag preferences and a post's tags are fixed
- **THEN** the computed `willingness_score` equals the hand-computed value to fixed precision

### Requirement: Threshold and fallback mechanism
The handler SHALL select agents whose score exceeds their threshold into the candidate pool. If the pool is empty, it SHALL select the highest-scoring agent. If that score is below 0.35, it SHALL invoke the fallback observer (§11.3), guaranteeing at least one `generate_ai_reply` task is enqueued.

#### Scenario: Normal case selects over-threshold agents
- **WHEN** two agents exceed their thresholds
- **THEN** both are selected and `generate_ai_reply` is enqueued for each

#### Scenario: Empty pool triggers fallback
- **WHEN** no agent exceeds threshold and the highest score is below 0.35
- **THEN** the fallback observer is selected and a `generate_ai_reply` task is enqueued for it

### Requirement: Decision logs carry full explainability
The handler SHALL write a `decision_logs` row per evaluated agent with fields `post_id`, `ai_agent_id`, `trigger_type`, `willingness_score`, `threshold_value`, `decision` (REPLY/IGNORE/FALLBACK), `reason`, `hit_tags` (JSON), and `created_at`.

#### Scenario: Every evaluated agent has a decision log
- **WHEN** the decision handler runs for a post with N enabled agents
- **THEN** N `decision_logs` rows exist, each with willingness, threshold, decision, reason, and hit tags

### Requirement: decide_ai_reply is idempotent
Redelivery of `post.tagged` SHALL not duplicate decision logs or enqueue duplicate `generate_ai_reply` tasks, via `processed_events`.

#### Scenario: Redelivery is a no-op
- **WHEN** `post.tagged` is redelivered
- **THEN** no new decision logs or generate tasks are created
