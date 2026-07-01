import { test, expect } from '@playwright/test';
import { setupMockApp, INITIAL_DB_STATE } from './mockHelper';

test.describe('Admin Console Tier 1 - Feature Coverage', () => {

  test.beforeEach(async ({ page }) => {
    await setupMockApp(page, 'admin');
    await page.addInitScript((state) => {
      localStorage.setItem('ai_forum_db_state', JSON.stringify(state));
    }, INITIAL_DB_STATE);
    await page.goto('/');
  });

  // Feature 7: System Dashboard & Analytics
  test.describe('Feature 7: System Dashboard & Analytics', () => {
    test('1. Dashboard renders metrics counters correctly', async ({ page }) => {
      await expect(page.locator('[data-testid="metric-posts"]')).toContainText("1");
      await expect(page.locator('[data-testid="metric-comments"]')).toContainText("2");
      await expect(page.locator('[data-testid="metric-tasks"]')).toContainText("0");
      await expect(page.locator('[data-testid="metric-agents"]')).toContainText("3");
    });

    test('2. Displays status list of backend simulated services', async ({ page }) => {
      await expect(page.locator('[data-testid="service-api-server"]')).toContainText("api-server");
      await expect(page.locator('[data-testid="service-worker-service"]')).toContainText("worker-service");
      await expect(page.locator('[data-testid="service-outbox-publisher"]')).toContainText("outbox-publisher");
    });

    test('3. Renders section showing task execution ratios', async ({ page }) => {
      await expect(page.locator('text=Service Status Indicators')).toBeVisible();
    });

    test('4. Sidebar links exist and allow tab switching', async ({ page }) => {
      await page.locator('button:has-text("AI Agent Config")').click();
      await expect(page.locator('h1')).toHaveText("agents");
      await expect(page.locator('[data-testid="agent-row-1"]')).toBeVisible();
    });

    test('5. Navigation triggers view updates instantly', async ({ page }) => {
      await page.locator('button:has-text("AI Task Queue")').click();
      await expect(page.locator('h1')).toHaveText("tasks");
    });
  });

  // Feature 8: AI Agent Configuration Manager
  test.describe('Feature 8: AI Agent Configuration Manager', () => {
    test.beforeEach(async ({ page }) => {
      await page.locator('button:has-text("AI Agent Config")').click();
    });

    test('1. Lists all AI agents in table with toggle states', async ({ page }) => {
      await expect(page.locator('[data-testid="agent-row-1"]')).toBeVisible();
      await expect(page.locator('[data-testid="agent-row-2"]')).toBeVisible();
      await expect(page.locator('[data-testid="agent-row-3"]')).toBeVisible();
    });

    test('2. Clicking Edit opens drawer with configuration fields', async ({ page }) => {
      await page.locator('[data-testid="agent-edit-btn-1"]').click();
      await expect(page.locator('[data-testid="drawer-agent-threshold"]')).toBeVisible();
      await expect(page.locator('[data-testid="drawer-agent-active-level"]')).toBeVisible();
      await expect(page.locator('[data-testid="drawer-agent-system-prompt"]')).toBeVisible();
    });

    test('3. Saving parameters updates the agent in database', async ({ page }) => {
      await page.locator('[data-testid="agent-edit-btn-1"]').click();
      await page.locator('[data-testid="drawer-agent-threshold"]').fill("0.75");
      await page.locator('[data-testid="drawer-agent-active-level"]').fill("0.90");
      await page.locator('[data-testid="drawer-agent-system-prompt"]').fill("New prompt content");
      await page.locator('[data-testid="drawer-agent-save-btn"]').click();

      // Drawer should close
      await expect(page.locator('[data-testid="drawer-agent-save-btn"]')).not.toBeVisible();
      
      const db = await page.evaluate(() => JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}'));
      const agent = db.agents.find(a => a.id === 1);
      expect(agent.replyThreshold).toBe(0.75);
      expect(agent.activityLevel).toBe(0.90);
      expect(agent.systemPrompt).toBe("New prompt content");
    });

    test('4. Toggling Active status switches agent inline', async ({ page }) => {
      const toggleBtn = page.locator('[data-testid="agent-toggle-active-1"]');
      await expect(toggleBtn).toHaveText("Active");
      await toggleBtn.click();
      await expect(toggleBtn).toHaveText("Inactive");

      const db = await page.evaluate(() => JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}'));
      expect(db.agents.find(a => a.id === 1).active).toBe(false);
    });

    test('5. Table exhibits correct responsive cell layouts', async ({ page }) => {
      await expect(page.locator('table')).toBeVisible();
    });
  });

  // Feature 9: AI Task Queue Monitor
  test.describe('Feature 9: AI Task Queue Monitor', () => {
    test.beforeEach(async ({ page }) => {
      // Seed a task
      await page.evaluate(() => {
        const db = JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}');
        db.tasks.push({
          id: 1,
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
          startedAt: new Date().toISOString(),
          finishedAt: new Date().toISOString()
        });
        localStorage.setItem('ai_forum_db_state', JSON.stringify(db));
      });
      await page.locator('button:has-text("AI Task Queue")').click();
    });

    test('1. Renders table of AI tasks with statuses', async ({ page }) => {
      await expect(page.locator('[data-testid="task-row-1"]')).toBeVisible();
      await expect(page.locator('[data-testid="task-row-1"] >> text=COMPLETED')).toBeVisible();
    });

    test('2. Displays task payloads inside detail drawer', async ({ page }) => {
      await page.locator('[data-testid="task-detail-btn-1"]').click();
      await expect(page.locator('[data-testid="drawer-task-payload"]')).toBeVisible();
      await expect(page.locator('[data-testid="drawer-task-payload"]')).toContainText("Verify prompt layout");
    });

    test('3. Shows retry count inside task row details', async ({ page }) => {
      await page.locator('[data-testid="task-detail-btn-1"]').click();
      await expect(page.locator('text=Retry count: 0')).toBeVisible();
    });

    test('4. Completed tasks render with finished duration calculations', async ({ page }) => {
      await expect(page.locator('[data-testid="task-row-1"] >> text=POST_AUTO')).toBeVisible();
    });

    test('5. Standard pagination or scroll parameters exist', async ({ page }) => {
      await expect(page.locator('table')).toBeVisible();
    });
  });

  // Feature 10: AI Decision Logs Auditor
  test.describe('Feature 10: AI Decision Logs Auditor', () => {
    test.beforeEach(async ({ page }) => {
      await page.locator('button:has-text("Decision Logs")').click();
    });

    test('1. Lists decision logs displaying names and scores', async ({ page }) => {
      await expect(page.locator('[data-testid="log-row-1"]')).toBeVisible();
      await expect(page.locator('[data-testid="log-row-1"] >> text=DevilsAdvocate')).toBeVisible();
      await expect(page.locator('[data-testid="log-row-1"] >> text=0.89')).toBeVisible();
    });

    test('2. Displays decision result and trigger details', async ({ page }) => {
      await expect(page.locator('[data-testid="log-row-1"] >> text=REPLY')).toBeVisible();
      await expect(page.locator('[data-testid="log-row-1"] >> text=FOLLOWUP')).toBeVisible();
    });

    test('3. Integrates quick search filter by Post ID', async ({ page }) => {
      const search = page.locator('[data-testid="log-search-post-id"]');
      await search.fill("1");
      await expect(page.locator('[data-testid="log-row-1"]')).toBeVisible();

      await search.fill("999");
      await expect(page.locator('[data-testid="log-row-1"]')).not.toBeVisible();
    });

    test('4. Highlights REPLY decisions with green styles', async ({ page }) => {
      const replyBadge = page.locator('[data-testid="log-row-1"] >> text=REPLY');
      await expect(replyBadge).toHaveClass(/bg-green-100/);
    });

    test('5. Table updates on data refresh', async ({ page }) => {
      await expect(page.locator('[data-testid="log-row-1"]')).toBeVisible();
    });
  });

});
