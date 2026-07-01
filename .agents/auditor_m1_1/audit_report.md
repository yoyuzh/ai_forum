# Forensic Audit Report

**Work Product**: `web/` (AI Forum user web application)
**Profile**: General Project (Integrity Forensics)
**Verdict**: CLEAN

### Phase Results
- **Hardcoded test outcomes**: PASS — No hardcoded test outcomes, mock assertions, or verification values were found. All responses and state changes are computed dynamically.
- **Dummy/facade implementations**: PASS — The mock database (`web/src/api/db.ts`) is fully functional and reads/writes to `localStorage`. The background AI Simulator (`web/src/sse/simulator.ts`) calculates willingness scores dynamically against agent thresholds and updates data correctly.
- **External network calls**: PASS — No external HTTP/HTTPS network calls (e.g., via `fetch` or `axios`) were found in `web/src/`. All logic executes locally in-browser.
- **Concurrency & Race Conditions**: FAIL (Functional Defect) — A concurrent simulation race condition was discovered where a post's status is prematurely updated to `COMPLETED` by a comment flow simulator while post-level reply tasks are still active in the background. This fails Test Case 5 of the verification suite (`verify_test.spec.ts`).
- **Feature Completeness**: FAIL (Functional Defect) — The web and admin codebases contain only minimal skeleton interfaces and do not implement all elements required by the E2E test suite (e.g., routing, comment trees, and Refine pages), causing Playwright tests in `web_t1.spec.ts` and `web_t2.spec.ts` to time out and fail.

---

### Evidence

#### 1. Verbatim Test Failure (Task-115 Log Output)
The verification test `npx playwright test tests/verify_test.spec.ts` fails on Test Case 5:
```
Running 2 tests using 2 workers

Database initialized.
Database initialized.

--- TEST CASE 5 (RACE CONDITION) START ---

At T=2.1s, post status is: COMPLETED
Events up to T=2.1s:
  [decision_log.created] at 1ms
  [decision_log.created] at 1ms
  [decision_log.created] at 1ms
  [post.updated] at 1ms
  [task.created] at 2ms
  [task.created] at 1001ms
  [task.updated] at 1503ms
  [task.created] at 2001ms
  [post.updated] at 2002ms

  1) [web] › tests/verify_test.spec.ts:14:5 › verify simulation flows and race conditions ──────────

    Error: expect(received).toBe(expected) // Object.is equality

    Expected: "PROCESSING"
    Received: "COMPLETED"

      164 |   // because the tasks for ArchTechLead, PM, and DevilsAdvocate are still running!
      165 |   // (ArchTechLead finishes at 3.5s, PM at 4.5s, DevilsAdvocate at 5.5s, final complete at 7.0s)
    > 166 |   expect(currentPostState?.aiStatus).toBe('PROCESSING');
          |                                      ^
```

#### 2. Code Review of the Race Condition (`web/src/sse/simulator.ts` Lines 146-156)
When a comment is created, it calls `runBackgroundAISimulation(postId, commentId)`. If no agent replies to the comment, the queue is empty and the following block executes:
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
This forces `aiStatus` to write `"COMPLETED"` to the database in 1 second, regardless of whether there are still active background tasks running for that post (which could take up to 7 seconds).

#### 3. Verification of localStorage DB (`web/src/api/db.ts` Lines 138-158)
The mock database writes state changes to local storage correctly:
```typescript
const DB_KEY = "ai_forum_db_state";

export class MockDatabase {
  private state: DatabaseState;

  constructor() {
    const saved = localStorage.getItem(DB_KEY);
    if (saved) {
      try {
        this.state = JSON.parse(saved);
      } catch {
        this.state = INITIAL_DB_STATE;
        this.save();
      }
    } else {
      this.state = INITIAL_DB_STATE;
      this.save();
    }
  }

  private save() {
    localStorage.setItem(DB_KEY, JSON.stringify(this.state));
  }
  ...
}
```
This confirms that the mock database does not act as a fake interface, but indeed persists changes to `localStorage` dynamically.
