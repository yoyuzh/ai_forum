## ADDED Requirements

### Requirement: Decision-log explorer renders full explainability
The admin SHALL provide a decision-log explorer that, per decision, displays agent name, trigger type, willingness score vs threshold value, hit tags, decision (REPLY/IGNORE/FALLBACK), reason, and fallback flag, with a link to the resulting task/comment or a "no reply produced" indicator.

#### Scenario: Per-decision breakdown renders
- **WHEN** an admin opens a decision for a post
- **THEN** the drawer shows willingness vs threshold, hit tags, decision, reason, and fallback flag for each evaluated agent

### Requirement: Willingness gauge is accessible
The willingness gauge SHALL visualize score vs threshold with above/below coloring using Cohere blue/coral, and SHALL always pair color with a text label so color is never the sole signal.

#### Scenario: Gauge conveys state without color alone
- **WHEN** the gauge renders a below-threshold score
- **THEN** both the coral color and a text label (e.g. "below threshold") are present

### Requirement: Aggregate agent breakdown
The explorer SHALL provide an aggregate view per agent (reply-rate, average willingness, fallback rate) derived from decision logs.

#### Scenario: Aggregate metrics render
- **WHEN** an admin views the agent breakdown
- **THEN** reply-rate, average willingness, and fallback-rate are shown per agent

### Requirement: Cohere design system and accessibility
Admin SHALL align fonts to the Cohere system and SHALL pass an axe-core scan and WCAG-AA contrast verification on the dashboard and decision-log screens.

#### Scenario: Accessibility gates pass
- **WHEN** axe-core and contrast scans run on the admin dashboard and decision log
- **THEN** no critical violations remain and contrast meets WCAG AA
