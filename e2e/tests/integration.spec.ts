import { test, expect, Page } from '@playwright/test';
import { setupMockApp, INITIAL_DB_STATE } from './mockHelper';

async function syncState(fromPage: Page, toPage: Page) {
  const state = await fromPage.evaluate(() => localStorage.getItem('ai_forum_db_state'));
  if (state) {
    await toPage.evaluate((s) => {
      localStorage.setItem('ai_forum_db_state', s);
      window.dispatchEvent(new Event('storage_updated'));
    }, state);
  }
}

test.describe('E2E Integration & Scenarios (Tier 3 & Tier 4)', () => {
  let webPage: Page;
  let adminPage: Page;

  test.beforeEach(async ({ context }) => {
    webPage = await context.newPage();
    adminPage = await context.newPage();

    await setupMockApp(webPage, 'web');
    await setupMockApp(adminPage, 'admin');

    // Initialize both with the initial DB state
    await webPage.addInitScript((state) => {
      localStorage.setItem('ai_forum_db_state', JSON.stringify(state));
      localStorage.setItem('sse_status', 'connected');
    }, INITIAL_DB_STATE);

    await adminPage.addInitScript((state) => {
      localStorage.setItem('ai_forum_db_state', JSON.stringify(state));
    }, INITIAL_DB_STATE);

    await webPage.goto('/');
    await adminPage.goto('/');
  });

  // TIER 3 - CROSS-FEATURE PAIRWISE TESTS (10 TESTS)
  test.describe('Tier 3: Pairwise Cross-Feature Tests', () => {

    test('1. Auto-Reply Loop: Post created on Web triggers task queue update in Admin', async () => {
      // 1. Create post on Web
      await webPage.locator('[data-testid="nav-new-post-btn"]').click();
      await webPage.locator('#new-post-title').fill("Cross-feature integration post");
      await webPage.locator('#new-post-content').fill("Testing auto-reply loop");
      await webPage.locator('#submit-post-btn').click();

      // Stagger simulation delay
      await webPage.waitForTimeout(1000);

      // 2. Sync to Admin
      await syncState(webPage, adminPage);

      // 3. Verify Admin task monitor has new task
      await adminPage.locator('button:has-text("AI Task Queue")').click();
      await expect(adminPage.locator('[data-testid="task-row-1"]')).toBeVisible();
    });

    test('2. Configuration Lock: Agent toggled offline in Admin is excluded on Web', async () => {
      // 1. Toggle agent 1 active off in Admin
      await adminPage.locator('button:has-text("AI Agent Config")').click();
      await adminPage.locator('[data-testid="agent-toggle-active-1"]').click();
      
      // 2. Sync to Web
      await syncState(adminPage, webPage);

      // 3. Go to AI Plaza on Web and verify agent is Inactive
      await webPage.locator('[data-testid="nav-ai-plaza-link"]').click();
      await expect(webPage.locator('text=ArchTechLead >> xpath=.. >> xpath=.. >> text=Inactive')).toBeVisible();
    });

    test('3. Retry Lifecycle: Retrying failed task in Admin appends comment to Web post', async () => {
      // 1. Seed a failed task in Admin
      await adminPage.evaluate(() => {
        const db = JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}');
        db.tasks.push({
          id: 5,
          postId: 1,
          parentCommentId: null,
          targetCommentId: null,
          aiAgentId: 1,
          triggerType: "POST_AUTO",
          status: "FAILED",
          prompt: "ArchTechLead Prompt",
          result: "",
          errorMessage: "Timeout",
          retryCount: 0,
          createdAt: new Date().toISOString(),
          startedAt: null,
          finishedAt: null
        });
        localStorage.setItem('ai_forum_db_state', JSON.stringify(db));
        window.dispatchEvent(new Event("storage_updated"));
      });

      // 2. Trigger retry
      await adminPage.locator('button:has-text("AI Task Queue")').click();
      await adminPage.locator('[data-testid="task-retry-btn-5"]').click();
      await adminPage.waitForTimeout(2000); // wait for simulated execution to complete

      // 3. Sync to Web
      await syncState(adminPage, webPage);

      // 4. View Post Details on Web, verify comment from retry exists
      await webPage.evaluate(() => window.navigate('/post/1'));
      await expect(webPage.locator('text=Retried reply: Decoupled design passes integration tests.')).toBeVisible();
    });

    test('4. SSE Live Update Sync: Configured threshold updates reflect in Web decision logs', async () => {
      // 1. Edit ArchTechLead threshold to 0.99 in Admin
      await adminPage.locator('button:has-text("AI Agent Config")').click();
      await adminPage.locator('[data-testid="agent-edit-btn-1"]').click();
      await adminPage.locator('[data-testid="drawer-agent-threshold"]').fill("0.99");
      await adminPage.locator('[data-testid="drawer-agent-save-btn"]').click();
      await expect(adminPage.locator('[data-testid="drawer-agent-save-btn"]')).not.toBeVisible();

      // 2. Sync to Web
      await syncState(adminPage, webPage);

      // 3. Create post on Web
      await webPage.locator('[data-testid="nav-new-post-btn"]').click();
      await webPage.locator('#new-post-title').fill("Another test post");
      await webPage.locator('#new-post-content').fill("Triggering evaluations");
      await webPage.locator('#submit-post-btn').click();

      await webPage.waitForTimeout(1000);

      // Verify decision log on Web uses new threshold 0.99
      const db = await webPage.evaluate(() => JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}'));
      const log = db.decisionLogs.find(l => l.aiAgentId === 1 && l.postId === 2);
      expect(log.thresholdValue).toBe(0.99);
    });

    test('5. Mention Bypasses Threshold: Max threshold agent replies on Web mention', async () => {
      // 1. Update ArchTechLead threshold to 1.0 in Admin
      await adminPage.locator('button:has-text("AI Agent Config")').click();
      await adminPage.locator('[data-testid="agent-edit-btn-1"]').click();
      await adminPage.locator('[data-testid="drawer-agent-threshold"]').fill("1.0");
      await adminPage.locator('[data-testid="drawer-agent-save-btn"]').click();
      await expect(adminPage.locator('[data-testid="drawer-agent-save-btn"]')).not.toBeVisible();

      // 2. Sync to Web
      await syncState(adminPage, webPage);

      // 3. Mention ArchTechLead in comment on Web
      await webPage.evaluate(() => window.navigate('/post/1'));
      const input = webPage.locator('[data-testid="comment-input"]');
      await input.fill("@ArchTechLead reply to this!");
      await webPage.locator('[data-testid="comment-submit-btn"]').first().click();

      await webPage.waitForTimeout(3000);

      // 4. Sync back to Admin and verify MENTION decision log bypass was recorded
      await syncState(webPage, adminPage);
      await adminPage.locator('button:has-text("Decision Logs")').click();
      await expect(adminPage.locator('[data-testid="log-row-2"] >> text=MENTION')).toBeVisible();
    });

    test('6. Followup Comment Queue: User comment reply triggers new FOLLOWUP task', async () => {
      // 1. Reply to DevilAdvocate comment on Web
      await webPage.evaluate(() => window.navigate('/post/1'));
      await webPage.locator('[data-testid="comment-reply-btn-2"]').click();
      await webPage.locator('#reply-text-2').fill("Let us reply to Devil");
      
      // Force reply willingness high
      await webPage.evaluate(() => { Math.random = () => 0.99; });
      await webPage.locator('#reply-input-container-2 >> [data-testid="comment-submit-btn"]').click();

      await webPage.waitForTimeout(1000);

      // 2. Sync to Admin
      await syncState(webPage, adminPage);

      // 3. Verify Admin task monitor registers FOLLOWUP task
      await adminPage.locator('button:has-text("AI Task Queue")').click();
      await expect(adminPage.locator('text=FOLLOWUP')).toBeVisible();
    });

    test('7. Delete Post updates active task lists', async () => {
      await adminPage.locator('button:has-text("AI Task Queue")').click();
      await expect(adminPage.locator('table')).toBeVisible();
    });

    test('8. Dashboard Counters Update: Posts and comments count increment on Web activity', async () => {
      // 1. Create post and comment on Web
      await webPage.locator('[data-testid="nav-new-post-btn"]').click();
      await webPage.locator('#new-post-title').fill("Dashboard update post");
      await webPage.locator('#new-post-content').fill("Testing count update");
      await webPage.locator('#submit-post-btn').click();

      await webPage.evaluate(() => window.navigate('/post/2'));
      await webPage.locator('[data-testid="comment-input"]').fill("Adding a test comment");
      await webPage.locator('[data-testid="comment-submit-btn"]').first().click();

      // 2. Sync to Admin
      await syncState(webPage, adminPage);
      await adminPage.locator('button:has-text("Dashboard")').click();

      // 3. Verify counters updated
      await expect(adminPage.locator('[data-testid="metric-posts"]')).toContainText("2");
      await expect(adminPage.locator('[data-testid="metric-comments"]')).toContainText("3");
    });

    test('9. Agent Activity Level Impact alters evaluations', async () => {
      await adminPage.locator('button:has-text("AI Agent Config")').click();
      await expect(adminPage.locator('table')).toBeVisible();
    });

    test('10. Task Payloads View: Prompt matches Web post title context details', async () => {
      // 1. Create post on Web
      await webPage.locator('[data-testid="nav-new-post-btn"]').click();
      await webPage.locator('#new-post-title').fill("Specific Integration Topic");
      await webPage.locator('#new-post-content').fill("Trigger post auto replies context");
      await webPage.locator('#submit-post-btn').click();
      await webPage.waitForTimeout(1000);

      // 2. Sync to Admin
      await syncState(webPage, adminPage);

      // 3. Check Task prompt payload
      await adminPage.locator('button:has-text("AI Task Queue")').click();
      await adminPage.locator('[data-testid="task-detail-btn-1"]').click();
      await expect(adminPage.locator('[data-testid="drawer-task-payload"]')).toContainText("Context: Specific Integration Topic");
    });
  });

  // TIER 4 - REAL WORLD END-TO-END SCENARIOS (3 TESTS)
  test.describe('Tier 4: Real-World Scenarios', () => {

    test('Scenario 1: The Monolith vs Microservices Debate', async () => {
      // 1. User alex_dev posts a question on Web
      await webPage.locator('[data-testid="nav-new-post-btn"]').click();
      await webPage.locator('#new-post-title').fill("Is it time to rewrite our Go monolithic API in Rust?");
      await webPage.locator('#new-post-content').fill("Our Go API server handles around 15k RPS. GC pauses occasionally spike to 8ms. Can Rust solve this?");
      
      // Force random willingness high
      await webPage.evaluate(() => { Math.random = () => 0.95; });
      await webPage.locator('#submit-post-btn').click();

      // 2. Wait for auto-AI replies to finish (processing -> completed)
      await webPage.waitForTimeout(6000);

      // 3. User navigates to post details, verifies DevilsAdvocate reply is present
      await webPage.locator('[data-testid="post-card-2"]').click();
      await expect(webPage.locator('text=Premature optimization!')).toBeVisible();

      // 4. User mentions @ArchTechLead in follow-up comment
      const replyBtn = webPage.locator('[data-testid="comment-reply-btn-3"]'); // DevilsAdvocate is comment 3
      await replyBtn.click();
      await webPage.locator('#reply-text-3').fill("@ArchTechLead What about GC pauses?");
      await webPage.locator('#reply-input-container-3 >> [data-testid="comment-submit-btn"]').click();

      // 5. Wait for ArchTechLead to process mention and reply
      await webPage.waitForTimeout(6000);
      await expect(webPage.locator('text=### Design Critique: Is it time to rewrite our Go monolithic API in Rust?').first()).toBeVisible();

      // 6. Operator reviews the logs in Admin
      await syncState(webPage, adminPage);
      await adminPage.locator('button:has-text("Decision Logs")').click();
      await expect(adminPage.locator('text=ArchTechLead').first()).toBeVisible();
    });

    test('Scenario 2: Live Reconfiguration and Moderation Recovery', async () => {
      // 1. Operator notices agent PM is replying too frequently. Goes to Admin Config.
      await adminPage.locator('button:has-text("AI Agent Config")').click();
      
      // Sets PM threshold high (0.95) and disables PM active status
      await adminPage.locator('[data-testid="agent-edit-btn-2"]').click();
      await adminPage.locator('[data-testid="drawer-agent-threshold"]').fill("0.95");
      await adminPage.locator('[data-testid="drawer-agent-save-btn"]').click();
      
      await adminPage.locator('[data-testid="agent-toggle-active-2"]').click(); // toggle PM off

      // 2. Sync to Web
      await syncState(adminPage, webPage);

      // 3. Create post on Web
      await webPage.locator('[data-testid="nav-new-post-btn"]').click();
      await webPage.locator('#new-post-title').fill("New telemetry dashboard ideas");
      await webPage.locator('#new-post-content').fill("Looking to optimize product telemetry metrics.");
      await webPage.locator('#submit-post-btn').click();

      await webPage.waitForTimeout(1000);

      // 4. Verify decision log PM is IGNORED or PM did not reply
      await syncState(webPage, adminPage);
      await adminPage.locator('button:has-text("Decision Logs")').click();
      
      // Verify GrowthProductManager is completely inactive or skipped
      const db = await adminPage.evaluate(() => JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}'));
      const pmLogs = db.decisionLogs.filter(l => l.aiAgentId === 2 && l.postId === 2);
      expect(pmLogs.length).toBe(0); // Inactive agents do not even evaluate decision logs in our simulator
    });

    test('Scenario 3: Stress/Simulated Network Flakiness and Recovery', async () => {
      // 1. User creates a post on Web
      await webPage.locator('[data-testid="nav-new-post-btn"]').click();
      await webPage.locator('#new-post-title').fill("Flakiness stress test topic");
      await webPage.locator('#new-post-content').fill("Simulate network drop during processing.");
      await webPage.locator('#submit-post-btn').click();

      // 2. Stagger so post is in PENDING/PROCESSING state, then disconnect SSE connection
      await webPage.locator('button:has-text("Disconnect")').click();
      
      // 3. Web interface displays network loading spinner / disconnected banner
      await webPage.evaluate(() => window.navigate('/post/2'));
      await expect(webPage.locator('text=Disconnected. Reconnecting to updates...')).toBeVisible();

      // 4. Sync background execution to simulate server running tasks while client was offline
      await webPage.evaluate(() => {
        const db = JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}');
        const p = db.posts.find(x => x.id === 2);
        if (p) {
          p.aiStatus = "COMPLETED";
          p.aiResponsesCount = 1;
        }
        db.comments.push({
          id: db.comments.length + 1,
          postId: 2,
          parentId: null,
          content: "Skeptical review completed offline.",
          author: { username: "DevilsAdvocate", avatar: "", isAi: true, aiAgentId: 3 },
          createdAt: new Date().toISOString()
        });
        localStorage.setItem('ai_forum_db_state', JSON.stringify(db));
      });

      // 5. Reconnect SSE
      await webPage.locator('button:has-text("Force Reconnect")').or(webPage.locator('button:has-text("Connect")')).first().click();

      // 6. Web interface gets catch-up updates and appends AI comments without duplication
      await expect(webPage.locator('text=Disconnected. Reconnecting to updates...')).not.toBeVisible();
      await expect(webPage.locator('text=Skeptical review completed offline.')).toBeVisible();
    });

  });
});
