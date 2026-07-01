# Original User Request

## Initial Request — 2026-06-30T13:10:05+08:00

Act as the E2E Testing Orchestrator.
Your working directory is: /Users/mac/Documents/ai_forum/.agents/sub_orch_e2e
Your mission is to establish the E2E testing infrastructure and write the test cases for both web/ and admin/ of the AI Forum, according to the Dual Track principles and the 4-tier testing methodology.
Please read:
- ORIGINAL_REQUEST.md: /Users/mac/Documents/ai_forum/.agents/orchestrator/ORIGINAL_REQUEST.md
- PROJECT.md: /Users/mac/Documents/ai_forum/.agents/orchestrator/PROJECT.md
- design_cohere.md: /Users/mac/Documents/ai_forum/stitch_ai_forum/design_cohere.md
- DESIGN.md: /Users/mac/Documents/ai_forum/stitch_ai_forum/synthetica_ai_forum/DESIGN.md

Your responsibilities:
1. Decompose the test suite creation into milestones and write them to SCOPE.md in your working directory.
2. Initialize the E2E test runner, configuration, and dependencies.
3. Write Tiers 1-4 test cases (at least 5 per feature for Tier 1 and 2, pairwise combinations for Tier 3, and real-world application scenarios for Tier 4). Keep the tests opaque-box (driven by user requirements and CLI/web interfaces).
4. Run/verify the tests.
5. Once complete, publish TEST_READY.md at the project root (/Users/mac/Documents/ai_forum/TEST_READY.md) following the exact template in the Project Pattern.
6. Report your status and handoff back to the Project Orchestrator (Conversation ID: c33e6d99-4c5a-402e-a647-972d46dc1f4b).
7. Update progress.md in your directory regularly.

MANDATORY INTEGRITY WARNING:
DO NOT CHEAT. All implementations must be genuine. DO NOT hardcode test results, create dummy/facade implementations, or circumvent the intended task. A Forensic Auditor will independently verify your work.
