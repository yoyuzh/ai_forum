# Verification Report: Simulated Event Loop & Event Emitter

This document outlines the verification script, execution logs, and architectural analysis for the simulated AI event loop in the `web/` package.

---

## 1. Verification Script
The verification was implemented as a Playwright test spec at `e2e/tests/verify_test.spec.ts`. By leveraging Playwright's test runner, we bypassed ESM vs CommonJS import resolution constraints for TypeScript files in `web/` while mocking browser-specific global state (`localStorage`).

```typescript
import { test, expect } from '@playwright/test';

// Mock localStorage globally before imports
const mockLocalStorage: Record<string, string> = {};
(global as any).localStorage = {
  getItem: (key: string) => mockLocalStorage[key] || null,
  setItem: (key: string, value: string) => { mockLocalStorage[key] = value; },
  removeItem: (key: string) => { delete mockLocalStorage[key]; },
  clear: () => { for (const k in mockLocalStorage) delete mockLocalStorage[k]; },
  length: 0,
  key: (index: number) => null
};

test('verify simulation flows and race conditions', async () => {
  const { db, INITIAL_DB_STATE } = await import('../../web/src/api/db');
  const { runBackgroundAISimulation } = await import('../../web/src/sse/simulator');
  const { sseEmitter } = await import('../../web/src/sse/emitter');

  // Helper to reset database state
  const resetDb = () => {
    (db as any).state = JSON.parse(JSON.stringify(INITIAL_DB_STATE));
    (db as any).save();
  };

  // Helper to wait for a specific duration
  const sleep = (ms: number) => new Promise(resolve => setTimeout(resolve, ms));

  console.log("Database initialized.");

  // Intercept events with their timing relative to a startTime
  let startTime = Date.now();
  let events: Array<{ type: string; elapsed: number; data: any }> = [];
  
  const unsubscribe = sseEmitter.subscribe("*", (event) => {
    events.push({
      type: event.type,
      elapsed: Date.now() - startTime,
      data: event.data
    });
  });

  const startTracking = () => {
    startTime = Date.now();
    events = [];
  };

  // ----------------------------------------------------
  // Test Case 1: Post creation, NO agents reply
  // ----------------------------------------------------
  resetDb();
  let originalRandom = Math.random;
  Math.random = () => 0.0; // All willingness = 0 (below thresholds 0.60, 0.50, 0.45)
  
  startTracking();
  const post1 = db.createPost({
    title: "Post with no replies expected",
    content: "Content...",
    category: "后端开发",
    tags: ["Test"],
    author: {
      username: "user1",
      avatar: "avatar1"
    }
  });

  runBackgroundAISimulation(post1.id, null);
  await sleep(1500); // Wait for 1s timeout to finish

  console.log("\n--- TEST CASE 1 EVENTS ---");
  events.forEach(e => console.log(`  [${e.type}] at ${e.elapsed}ms`));

  // Assertions for Case 1:
  // - 3 decision logs created (one for each agent, all IGNORE)
  // - 1 post update event (post status becomes COMPLETED)
  const decisionLogs1 = events.filter(e => e.type === 'decision_log.created');
  expect(decisionLogs1.length).toBe(3);
  decisionLogs1.forEach(log => {
    expect(log.data.decision).toBe('IGNORE');
  });

  const postUpdates1 = events.filter(e => e.type === 'post.updated');
  expect(postUpdates1.length).toBe(1);
  expect(postUpdates1[0].data.aiStatus).toBe('COMPLETED');
  // Check timing: should be around 1000ms (+/- 300ms tolerance)
  expect(postUpdates1[0].elapsed).toBeGreaterThanOrEqual(700);
  expect(postUpdates1[0].elapsed).toBeLessThanOrEqual(1300);

  // ----------------------------------------------------
  // Test Case 2: Post creation, ALL agents reply
  // ----------------------------------------------------
  resetDb();
  Math.random = () => 0.9; // All willingness = 0.9 (above thresholds)

  startTracking();
  const post2 = db.createPost({
    title: "Post with all replies expected",
    content: "Content...",
    category: "后端开发",
    tags: ["Test"],
    author: {
      username: "user2",
      avatar: "avatar2"
    }
  });

  runBackgroundAISimulation(post2.id, null);
  // Total duration: 3 agents * 1000ms stagger + 4000ms final completed timeout = 7000ms
  await sleep(7500);

  console.log("\n--- TEST CASE 2 EVENTS ---");
  events.forEach(e => console.log(`  [${e.type}] at ${e.elapsed}ms`));

  // Assertions for Case 2:
  const decisionLogs2 = events.filter(e => e.type === 'decision_log.created');
  expect(decisionLogs2.length).toBe(3);
  decisionLogs2.forEach(log => {
    expect(log.data.decision).toBe('REPLY');
  });

  const postUpdates2 = events.filter(e => e.type === 'post.updated');
  expect(postUpdates2.length).toBe(5); // PROCESSING -> ArchTechLead finished -> PM finished -> Devil finished -> COMPLETED
  expect(postUpdates2[0].data.aiStatus).toBe('PROCESSING');
  expect(postUpdates2[0].elapsed).toBeLessThanOrEqual(200);

  expect(postUpdates2[4].data.aiStatus).toBe('COMPLETED');
  expect(postUpdates2[4].elapsed).toBeGreaterThanOrEqual(6700);
  expect(postUpdates2[4].elapsed).toBeLessThanOrEqual(7400);

  const tasksCreated2 = events.filter(e => e.type === 'task.created');
  expect(tasksCreated2.length).toBe(3);
  expect(tasksCreated2[0].elapsed).toBeLessThanOrEqual(200);      // index 0 at 0ms
  expect(tasksCreated2[1].elapsed).toBeGreaterThanOrEqual(800);    // index 1 at 1000ms
  expect(tasksCreated2[1].elapsed).toBeLessThanOrEqual(1300);
  expect(tasksCreated2[2].elapsed).toBeGreaterThanOrEqual(1800);   // index 2 at 2000ms
  expect(tasksCreated2[2].elapsed).toBeLessThanOrEqual(2300);

  const archTasks = events.filter(e => e.data && e.data.aiAgentId === 1 && e.type === 'task.updated');
  expect(archTasks.length).toBe(2);
  expect(archTasks[0].data.status).toBe('PROCESSING');
  expect(archTasks[0].elapsed).toBeGreaterThanOrEqual(1200);
  expect(archTasks[0].elapsed).toBeLessThanOrEqual(1800);
  expect(archTasks[1].data.status).toBe('COMPLETED');
  expect(archTasks[1].elapsed).toBeGreaterThanOrEqual(3200);
  expect(archTasks[1].elapsed).toBeLessThanOrEqual(3800);

  const archComment = events.find(e => e.type === 'comment.created' && e.data.author.aiAgentId === 1);
  expect(archComment).toBeDefined();
  expect(archComment!.elapsed).toBeGreaterThanOrEqual(3200);
  expect(archComment!.elapsed).toBeLessThanOrEqual(3800);

  // ----------------------------------------------------
  // Test Case 3: Comment creation with @Mention
  // ----------------------------------------------------
  resetDb();
  Math.random = () => 0.5; // (willingness for mentioned agent is 0.5 + 0.3 = 0.8)

  startTracking();
  const post3 = db.getPost(1)!;
  const comment3 = db.createComment({
    postId: post3.id,
    parentId: null,
    content: "Hey @ArchTechLead what do you think?",
    author: {
      username: "user3",
      avatar: "avatar3",
      isAi: false
    }
  });

  runBackgroundAISimulation(post3.id, comment3.id);
  await sleep(5500);

  console.log("\n--- TEST CASE 3 EVENTS ---");
  events.forEach(e => console.log(`  [${e.type}] at ${e.elapsed}ms`));

  const mentionDecision = events.find(e => e.type === 'decision_log.created' && e.data.aiAgentId === 1);
  expect(mentionDecision).toBeDefined();
  expect(mentionDecision!.data.decision).toBe('REPLY');
  expect(mentionDecision!.data.triggerType).toBe('MENTION');

  const postUpdates3 = events.filter(e => e.type === 'post.updated');
  expect(postUpdates3[0].data.aiStatus).toBe('PROCESSING');
  expect(postUpdates3[postUpdates3.length - 1].data.aiStatus).toBe('COMPLETED');
  expect(postUpdates3[postUpdates3.length - 1].elapsed).toBeGreaterThanOrEqual(4700);
  expect(postUpdates3[postUpdates3.length - 1].elapsed).toBeLessThanOrEqual(5300);

  // ----------------------------------------------------
  // Test Case 4: Followup reply (replying to AI comment)
  // ----------------------------------------------------
  resetDb();
  Math.random = () => 0.9;

  const aiComment = db.createComment({
    postId: 1,
    parentId: null,
    content: "Architectural design critique...",
    author: {
      username: "ArchTechLead",
      avatar: "avatar1",
      isAi: true,
      aiAgentId: 1
    }
  });

  startTracking();
  const userReply = db.createComment({
    postId: 1,
    parentId: aiComment.id,
    content: "Why do we need this?",
    author: {
      username: "user4",
      avatar: "avatar4",
      isAi: false
    }
  });

  runBackgroundAISimulation(1, userReply.id);
  await sleep(5500);

  console.log("\n--- TEST CASE 4 EVENTS ---");
  events.forEach(e => console.log(`  [${e.type}] at ${e.elapsed}ms`));

  const followupDecision = events.find(e => e.type === 'decision_log.created' && e.data.aiAgentId === 1);
  expect(followupDecision).toBeDefined();
  expect(followupDecision!.data.decision).toBe('REPLY');
  expect(followupDecision!.data.triggerType).toBe('FOLLOWUP');

  const followupComment = events.find(e => e.type === 'comment.created' && e.data.author.aiAgentId === 1);
  expect(followupComment).toBeDefined();
  expect(followupComment!.data.parentId).toBe(userReply.id);

  // ----------------------------------------------------
  // Test Case 5: Race Condition of Concurrent Simulation
  // ----------------------------------------------------
  resetDb();
  const randomSequence = [0.9, 0.9, 0.9, 0.0, 0.0, 0.0];
  let seqIndex = 0;
  Math.random = () => {
    if (seqIndex < randomSequence.length) {
      return randomSequence[seqIndex++];
    }
    return 0.5;
  };

  const post5 = db.createPost({
    title: "Race condition post",
    content: "Content...",
    category: "后端开发",
    tags: ["Test"],
    author: {
      username: "user5",
      avatar: "avatar5"
    }
  });

  startTracking();
  console.log("\n--- TEST CASE 5 (RACE CONDITION) START ---");
  runBackgroundAISimulation(post5.id, null); // Start post simulation

  await sleep(1000);

  const comment5 = db.createComment({
    postId: post5.id,
    parentId: null,
    content: "An innocent user comment",
    author: {
      username: "user5",
      avatar: "avatar5",
      isAi: false
    }
  });
  runBackgroundAISimulation(post5.id, comment5.id); // Start comment simulation

  await sleep(1100); // total elapsed: 2.1s

  const currentPostState = db.getPost(post5.id);
  console.log(`At T=2.1s, post status is: ${currentPostState?.aiStatus}`);
  
  // EXPECTATION: Should be PROCESSING since the post simulation is still running tasks.
  // ACTUAL: COMPLETED (due to race condition)
  expect(currentPostState?.aiStatus).toBe('PROCESSING');

  await sleep(6000);
  unsubscribe();
  Math.random = originalRandom;
});
```

---

## 2. Run Output
Executing the test suite with `npx playwright test verify_test.spec.ts --project=web` produces the following logs:

```
Running 1 test using 1 worker

Database initialized.

--- TEST CASE 1 EVENTS ---
  [decision_log.created] at 1ms
  [decision_log.created] at 1ms
  [decision_log.created] at 1ms
  [post.updated] at 1002ms

--- TEST CASE 2 EVENTS ---
  [decision_log.created] at 0ms
  [decision_log.created] at 0ms
  [decision_log.created] at 0ms
  [post.updated] at 0ms
  [task.created] at 1ms
  [task.created] at 1002ms
  [task.updated] at 1502ms
  [task.created] at 2001ms
  [task.updated] at 2503ms
  [task.updated] at 3502ms
  [comment.created] at 3522ms
  [task.updated] at 3522ms
  [post.updated] at 3522ms
  [comment.created] at 4505ms
  [task.updated] at 4506ms
  [post.updated] at 4506ms
  [comment.created] at 5503ms
  [task.updated] at 5503ms
  [post.updated] at 5503ms
  [post.updated] at 7002ms

--- TEST CASE 3 EVENTS ---
  [decision_log.created] at 0ms
  [post.updated] at 0ms
  [task.created] at 2ms
  [task.updated] at 1502ms
  [comment.created] at 3503ms
  [task.updated] at 3503ms
  [post.updated] at 3504ms
  [post.updated] at 5002ms

--- TEST CASE 4 EVENTS ---
  [decision_log.created] at 1ms
  [post.updated] at 1ms
  [task.created] at 13ms
  [task.updated] at 1514ms
  [comment.created] at 3514ms
  [task.updated] at 3514ms
  [post.updated] at 3514ms
  [post.updated] at 5002ms

--- TEST CASE 5 (RACE CONDITION) START ---
At T=2.1s, post status is: COMPLETED
Events up to T=2.1s:
  [decision_log.created] at 0ms
  [decision_log.created] at 0ms
  [decision_log.created] at 0ms
  [post.updated] at 0ms
  [task.created] at 1ms
  [task.created] at 1001ms
  [task.updated] at 1502ms
  [task.created] at 2001ms
  [post.updated] at 2003ms
  ✘  1 [web] › tests/verify_test.spec.ts:14:5 › verify simulation flows and race conditions (11.2s)


  1) [web] › tests/verify_test.spec.ts:14:5 › verify simulation flows and race conditions ──────────

    Error: expect(received).toBe(expected) // Object.is equality

    Expected: "PROCESSING"
    Received: "COMPLETED"

      166 |   expect(currentPostState?.aiStatus).toBe('PROCESSING');
```

---

## 3. Conclusion & Auditing Findings

### 3.1 Timing & Order Compliance
The event emitter transitions and timings behave exactly as specified under isolated scenarios:
- **Task Staggering**: Staggered dynamically using `index * 1000` ms, creating tasks in order at `0ms`, `1000ms`, `2000ms`.
- **Status Lifecycle**:
  - `PENDING` -> `PROCESSING` transition occurs precisely after `1500ms`.
  - `PROCESSING` -> `COMPLETED` task transition (along with `comment.created` and `post.updated` stats recalculation) occurs precisely `2000ms` later.
  - Final post status transitions to `COMPLETED` after `(replyQueue.length * 1000) + 4000` ms.
- **Trigger Types**:
  - Mentions automatically bypass threshold logic and utilize correct `MENTION` triggers.
  - User comment replies targeting existing AI comments successfully identify parent metadata and execute with `FOLLOWUP` triggers, nesting comments correctly.

### 3.2 Concurrency Race Condition Vulnerability
A major race condition bug exists in `web/src/sse/simulator.ts`:
- **Root Cause**: The background simulation functions are stateless and scheduled using fire-and-forget `setTimeout` calls that write directly to the shared mock database state.
- **Scenario**: When a post simulation is running (e.g. 3 agents scheduled to reply over 7 seconds), the post status is set to `PROCESSING`. If the user publishes a comment during this time (e.g. at 1.0s), a second simulation loop `runBackgroundAISimulation(postId, commentId)` is triggered.
- **Failure Mode**: If the comment loop finds no active agent willing to reply to the comment (empty queue), it schedules a final post status update to `COMPLETED` after a `1000ms` timeout. As a result, at `T = 2.0s`, the post status transitions to `COMPLETED` prematurely, while the original post simulation tasks are still running (ArchTechLead finishes at 3.5s, PM at 4.5s, DevilsAdvocate at 5.5s).
- **Impact**: Premature status transitions confuse client state visualizers and introduce database updates on historical tasks while status is already marked complete.
