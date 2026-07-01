# Handoff Report: Simulated Event Loop Verification

This report provides the findings of the empirical challenge of the simulated AI event loop in the `web/` package.

## 1. Observation
- Verified file path: `web/src/sse/simulator.ts`
- Verified file path: `web/src/sse/emitter.ts`
- Verified file path: `web/src/api/db.ts`
- Verified E2E environment using: `npx playwright test verify_test.spec.ts --project=web` (logs captured in `verification.md` under the working directory `.agents/challenger_m1_1/`).
- Verbatim execution output for race condition:
```
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

    Error: expect(received).toBe(expected) // Object.is equality

    Expected: "PROCESSING"
    Received: "COMPLETED"
```
- In `web/src/sse/simulator.ts` lines 146-156:
```typescript
  if (replyQueue.length === 0) {
    // If no agent replies, transition post state to COMPLETED after a brief period
    setTimeout(() => {
      const latestPost = db.getPost(postId);
      if (latestPost) {
        db.updatePost(postId, { aiStatus: "COMPLETED" });
        sseEmitter.emit("post.updated", db.getPost(postId));
      }
    }, 1000);
    return;
  }
```

## 2. Logic Chain
1. A post creation triggers `runBackgroundAISimulation(postId, null)` which transitions `post.aiStatus` to `PROCESSING` at `T=0`.
2. Staggered tasks are created at `T=0ms`, `T=1000ms`, `T=2000ms`. These tasks run asynchronously over a total span of `7000ms` before the final post status update is scheduled to become `COMPLETED`.
3. If a user publishes a comment at `T=1000ms`, a separate simulation loop is triggered `runBackgroundAISimulation(postId, commentId)`.
4. If no agents reply to the comment, the queue is empty (`replyQueue.length === 0`).
5. As observed in the code snippet from `simulator.ts:146-156`, the loop schedules a direct database write of `aiStatus = "COMPLETED"` after a `1000ms` timeout (`T = 1000ms + 1000ms = 2000ms`).
6. Because the original post tasks are still actively processing at `T=2000ms` (e.g. DevilsAdvocate task starts processing at `3500ms`), setting `aiStatus = "COMPLETED"` at `T=2000ms` is premature and breaks sequential consistency.

## 3. Caveats
- Verified only in-memory mock client-side simulations. The production Go backend outbox-publisher and worker-service are not involved since web application features mock behaviors client-side for Milestone 1.

## 4. Conclusion
- The simulated event loop correctly conforms to all timing constraints, event ordering, and staggering thresholds under isolated post-only or comment-only conditions.
- However, there is a verified race condition when concurrent flows target the same post. In-flight simulations will overwrite each other's status, leading to premature transitions of the post status to `COMPLETED` when comment replies are ignored.

## 5. Verification Method
- Re-run E2E validation spec in `verify_test.spec.ts` (using the source code copied in `verification.md`) by copying it back to `e2e/tests/verify_test.spec.ts` and running:
  ```bash
  npx playwright test verify_test.spec.ts --project=web
  ```
- File to inspect: `web/src/sse/simulator.ts`
