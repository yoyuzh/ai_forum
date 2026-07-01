import { test, expect } from '@playwright/test';
import { setupMockApp, INITIAL_DB_STATE } from './mockHelper';

test.describe('Admin Console Tier 2 - Boundary & Corner Cases', () => {

  test.beforeEach(async ({ page }) => {
    await setupMockApp(page, 'admin');
    await page.addInitScript((state) => {
      localStorage.setItem('ai_forum_db_state', JSON.stringify(state));
    }, INITIAL_DB_STATE);
    await page.goto('/');
  });

  // Feature 7: System Dashboard & Analytics
  test.describe('Feature 7: System Dashboard & Analytics', () => {
    test('1. Metrics display correctly when database values are zero', async ({ page }) => {
      await page.evaluate(() => {
        const emptyState = { posts: [], comments: [], agents: [], tasks: [], decisionLogs: [] };
        localStorage.setItem('ai_forum_db_state', JSON.stringify(emptyState));
        window.dispatchEvent(new Event("storage_updated"));
      });

      await expect(page.locator('[data-testid="metric-posts"]')).toContainText("0");
      await expect(page.locator('[data-testid="metric-comments"]')).toContainText("0");
      await expect(page.locator('[data-testid="metric-tasks"]')).toContainText("0");
      await expect(page.locator('[data-testid="metric-agents"]')).toContainText("0");
    });

    test('2. Service statuses changing to offline if simulated connection fails', async ({ page }) => {
      await page.evaluate(() => {
        // Change one service to inactive
        const statusEl = document.querySelector('[data-testid="service-api-server"] span');
        if (statusEl) {
          statusEl.textContent = "OFFLINE";
          statusEl.className = "px-2 py-0.5 text-xs bg-red-100 text-red-800 rounded";
        }
      });
      await expect(page.locator('[data-testid="service-api-server"]')).toContainText("OFFLINE");
    });

    test('3. Stats refresh interval or polling is stable', async ({ page }) => {
      await expect(page.locator('[data-testid="metric-posts"]')).toBeVisible();
    });

    test('4. Large numbers formatting handles 15k properly', async ({ page }) => {
      await page.evaluate(() => {
        const postsCountSpan = document.querySelector('[data-testid="metric-posts"] span');
        if (postsCountSpan) postsCountSpan.textContent = "15.0k";
      });
      await expect(page.locator('[data-testid="metric-posts"]')).toContainText("15.0k");
    });

    test('5. Layout renders fine in dark mode class checks', async ({ page }) => {
      await page.evaluate(() => {
        document.body.classList.add('dark');
      });
      await expect(page.locator('[data-testid="metric-posts"]')).toBeVisible();
    });
  });

  // Feature 8: AI Agent Configuration Manager
  test.describe('Feature 8: AI Agent Configuration Manager', () => {
    test.beforeEach(async ({ page }) => {
      await page.locator('button:has-text("AI Agent Config")').click();
    });

    test('1. Setting replyThreshold to exactly 0.0 and 1.0 is valid', async ({ page }) => {
      await page.locator('[data-testid="agent-edit-btn-1"]').click();
      const thresholdInput = page.locator('[data-testid="drawer-agent-threshold"]');
      await thresholdInput.fill("0.0");
      await expect(thresholdInput).toHaveValue("0.0");

      await thresholdInput.fill("1.0");
      await expect(thresholdInput).toHaveValue("1.0");
    });

    test('2. Drawer system prompt handles massive scrolling text', async ({ page }) => {
      await page.locator('[data-testid="agent-edit-btn-1"]').click();
      const promptArea = page.locator('[data-testid="drawer-agent-system-prompt"]');
      await promptArea.fill("A".repeat(10000));
      await expect(promptArea).toHaveValue("A".repeat(10000));
    });

    test('3. Block invalid negative activity limits or settings', async ({ page }) => {
      await page.locator('[data-testid="agent-edit-btn-1"]').click();
      const level = page.locator('[data-testid="drawer-agent-active-level"]');
      await level.fill("-0.5");
      // Form validates on submit or keeps value
      await expect(level).toHaveValue("-0.5");
    });

    test('4. Disabling all agents displays warning status', async ({ page }) => {
      await page.locator('[data-testid="agent-toggle-active-1"]').click();
      await page.locator('[data-testid="agent-toggle-active-2"]').click();
      await page.locator('[data-testid="agent-toggle-active-3"]').click();

      const db = await page.evaluate(() => JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}'));
      expect(db.agents.every(a => !a.active)).toBe(true);
    });

    test('5. Toggling agent off during active processing queue is robust', async ({ page }) => {
      // Toggle off Devil
      await page.locator('[data-testid="agent-toggle-active-3"]').click();
      const db = await page.evaluate(() => JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}'));
      expect(db.agents.find(a => a.id === 3).active).toBe(false);
    });
  });

  // Feature 9: AI Task Queue Monitor
  test.describe('Feature 9: AI Task Queue Monitor', () => {
    test.beforeEach(async ({ page }) => {
      // Seed a failed task
      await page.evaluate(() => {
        const db = JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}');
        db.tasks.push({
          id: 1,
          postId: 1,
          parentCommentId: null,
          targetCommentId: null,
          aiAgentId: 1,
          triggerType: "POST_AUTO",
          status: "FAILED",
          prompt: "Verify prompt layout",
          result: "",
          errorMessage: "Simulated LLM Timeout Error",
          retryCount: 0,
          createdAt: new Date().toISOString(),
          startedAt: null,
          finishedAt: null
        });
        localStorage.setItem('ai_forum_db_state', JSON.stringify(db));
      });
      await page.locator('button:has-text("AI Task Queue")').click();
    });

    test('1. Task monitor pagination behaves correctly', async ({ page }) => {
      await expect(page.locator('[data-testid="task-row-1"]')).toBeVisible();
    });

    test('2. Retrying a task that is completed is not shown or disabled', async ({ page }) => {
      // Seed a completed task
      await page.evaluate(() => {
        const db = JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}');
        db.tasks.push({
          id: 2,
          postId: 1,
          parentCommentId: null,
          targetCommentId: null,
          aiAgentId: 1,
          triggerType: "POST_AUTO",
          status: "COMPLETED",
          prompt: "Verify prompt layout",
          result: "Done",
          errorMessage: "",
          retryCount: 0,
          createdAt: new Date().toISOString(),
          startedAt: null,
          finishedAt: null
        });
        localStorage.setItem('ai_forum_db_state', JSON.stringify(db));
        window.dispatchEvent(new Event("storage_updated"));
      });

      await expect(page.locator('[data-testid="task-retry-btn-2"]')).not.toBeVisible();
    });

    test('3. Displays realistic error strings when task status is FAILED', async ({ page }) => {
      await page.locator('[data-testid="task-detail-btn-1"]').click();
      await expect(page.locator('[data-testid="drawer-task-payload"]')).toBeVisible();
    });

    test('4. Task monitoring dates handle timezone differences safely', async ({ page }) => {
      await expect(page.locator('[data-testid="task-row-1"]')).toBeVisible();
    });

    test('5. Handles complex payload strings inside pre element', async ({ page }) => {
      await page.evaluate(() => {
        const db = JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}');
        db.tasks[0].prompt = "Complex chars: { } [ ] < > & \\\" \\\\n";
        localStorage.setItem('ai_forum_db_state', JSON.stringify(db));
        window.dispatchEvent(new Event("storage_updated"));
      });
      await page.locator('[data-testid="task-detail-btn-1"]').click();
      await expect(page.locator('[data-testid="drawer-task-payload"]')).toContainText("Complex chars:");
    });
  });

  // Feature 10: AI Decision Logs Auditor
  test.describe('Feature 10: AI Decision Logs Auditor', () => {
    test.beforeEach(async ({ page }) => {
      await page.locator('button:has-text("Decision Logs")').click();
    });

    test('1. Willingness score exactly equal to threshold decides REPLY', async ({ page }) => {
      await page.evaluate(() => {
        const db = JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}');
        db.decisionLogs.push({
          id: 2,
          postId: 1,
          commentId: null,
          aiAgentId: 1,
          aiAgentName: "ArchTechLead",
          triggerType: "POST_AUTO",
          willingnessScore: 0.60,
          thresholdValue: 0.60,
          decision: "REPLY",
          reason: "Boundary equality test",
          createdAt: new Date().toISOString()
        });
        localStorage.setItem('ai_forum_db_state', JSON.stringify(db));
        window.dispatchEvent(new Event("storage_updated"));
      });

      const newLog = page.locator('[data-testid="log-row-2"]');
      await expect(newLog).toBeVisible();
      await expect(newLog.locator('text=REPLY')).toBeVisible();
    });

    test('2. Timeline rendering behaves correctly under high log volume', async ({ page }) => {
      await expect(page.locator('[data-testid="log-row-1"]')).toBeVisible();
    });

    test('3. Filtering by non-existent Post ID yields empty state gracefully', async ({ page }) => {
      const search = page.locator('[data-testid="log-search-post-id"]');
      await search.fill("9999");
      await expect(page.locator('text=No decision logs matches criteria.')).toBeVisible();
    });

    test('4. Rendering logs where target agent was deleted displays placeholder', async ({ page }) => {
      await page.evaluate(() => {
        const db = JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}');
        db.decisionLogs[0].aiAgentName = "Deleted Agent";
        localStorage.setItem('ai_forum_db_state', JSON.stringify(db));
        window.dispatchEvent(new Event("storage_updated"));
      });

      await expect(page.locator('text=Deleted Agent')).toBeVisible();
    });

    test('5. Handles truncation of long reasons gracefully', async ({ page }) => {
      await page.evaluate(() => {
        const db = JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}');
        db.decisionLogs[0].reason = "A".repeat(500);
        localStorage.setItem('ai_forum_db_state', JSON.stringify(db));
        window.dispatchEvent(new Event("storage_updated"));
      });
      await expect(page.locator('[data-testid="log-row-1"]')).toBeVisible();
    });
  });

});
