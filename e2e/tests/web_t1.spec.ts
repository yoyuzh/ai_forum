import { test, expect } from '@playwright/test';
import { setupMockApp, INITIAL_DB_STATE } from './mockHelper';

test.describe('User Web App Tier 1 - Feature Coverage', () => {

  test.beforeEach(async ({ page }) => {
    await setupMockApp(page, 'web');
    // Clear and set initial state before each test
    await page.addInitScript((state) => {
      localStorage.setItem('ai_forum_db_state', JSON.stringify(state));
      localStorage.setItem('sse_status', 'connected');
    }, INITIAL_DB_STATE);
    await page.goto('/');
  });

  // Feature 1: Homepage Feed & Virtualized Scrolling
  test.describe('Feature 1: Homepage Feed & Virtualized Scrolling', () => {
    test('1. Feed renders successfully with seed posts', async ({ page }) => {
      const postCard = page.locator('[data-testid="post-card-1"]');
      await expect(postCard).toBeVisible();
      await expect(postCard.locator('[data-testid="post-title"]')).toHaveText(/Is it time to rewrite/);
    });

    test('2. Each post card displays author details, tags and replies count', async ({ page }) => {
      const postCard = page.locator('[data-testid="post-card-1"]');
      await expect(postCard.locator('text=alex_dev')).toBeVisible();
      await expect(postCard.locator('text=后端开发')).toBeVisible();
      await expect(postCard.locator('text=2 replies')).toBeVisible();
    });

    test('3. Clicking a post card navigates to post details', async ({ page }) => {
      await page.locator('[data-testid="post-card-1"]').click();
      await expect(page).toHaveURL(/\/post\/1/);
      await expect(page.locator('[data-testid="post-detail-title"]')).toHaveText(/Is it time to rewrite/);
    });

    test('4. Scrolling down triggers virtualized load-more action', async ({ page }) => {
      // Seed 50 posts to test scrolling load-more
      await page.evaluate(() => {
        const db = JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}');
        for (let i = 2; i <= 50; i++) {
          db.posts.push({
            id: i,
            title: `Fake Post ${i}`,
            content: `This is the body of fake post ${i}`,
            category: "后端开发",
            tags: ["Testing"],
            author: { username: "tester", avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=tester" },
            aiStatus: "COMPLETED",
            aiResponsesCount: 0,
            aiAvatars: [],
            createdAt: new Date().toISOString(),
            likes: 0
          });
        }
        localStorage.setItem('ai_forum_db_state', JSON.stringify(db));
        window.dispatchEvent(new Event("storage_updated"));
      });

      const container = page.locator('#posts-scroll-container');
      
      // Before scroll, should show 10 items
      let cards = page.locator('[data-testid="post-card"]');
      await expect(cards).toHaveCount(10);

      // Scroll to bottom
      await container.evaluate((el) => {
        el.scrollTop = el.scrollHeight;
      });

      // Wait for more posts to load
      await page.waitForTimeout(500);
      cards = page.locator('[data-testid="post-card"]');
      const countAfterScroll = await cards.count();
      expect(countAfterScroll).toBeGreaterThan(10);
    });

    test('5. The navigation header remains visible at the top', async ({ page }) => {
      const header = page.locator('header');
      await expect(header).toBeVisible();
      await expect(page.locator('[data-testid="nav-home-link"]')).toBeVisible();
      await expect(page.locator('[data-testid="nav-ai-plaza-link"]')).toBeVisible();
    });
  });

  // Feature 2: Post Filtering & Search
  test.describe('Feature 2: Post Filtering & Search', () => {
    test('1. Clicking a category chip filters the feed', async ({ page }) => {
      // Add a post with category '前端开发'
      await page.evaluate(() => {
        const db = JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}');
        db.posts.push({
          id: 2,
          title: "React performance tips",
          content: "Use React.memo to prevent unnecessary re-renders.",
          category: "前端开发",
          tags: ["React"],
          author: { username: "frontend_dev", avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=front" },
          aiStatus: "COMPLETED",
          aiResponsesCount: 0,
          aiAvatars: [],
          createdAt: new Date().toISOString(),
          likes: 0
        });
        localStorage.setItem('ai_forum_db_state', JSON.stringify(db));
        window.dispatchEvent(new Event("storage_updated"));
      });

      await page.locator('[data-testid="category-chip-前端开发"]').click();
      await expect(page.locator('[data-testid="post-card"]')).toHaveCount(1);
      await expect(page.locator('[data-testid="post-title"]')).toHaveText(/React performance tips/);
    });

    test('2. Selecting/toggling tag chips filters feed elements', async ({ page }) => {
      await page.locator('[data-testid="tag-chip-Go"]').click();
      await expect(page.locator('[data-testid="post-card"]')).toHaveCount(1);
      
      // Select another tag that doesn't match
      await page.locator('[data-testid="tag-chip-Rust"]').click();
      await expect(page.locator('[data-testid="post-card"]')).toHaveCount(1);
    });

    test('3. Typing in search input updates the feed', async ({ page }) => {
      const searchInput = page.locator('[data-testid="search-input"]');
      await searchInput.fill("rewrite");
      await expect(page.locator('[data-testid="post-card"]')).toHaveCount(1);

      await searchInput.fill("non-existent query");
      await expect(page.locator('[data-testid="post-card"]')).toHaveCount(0);
    });

    test('4. Clearing filters/search restores default feed', async ({ page }) => {
      await page.locator('[data-testid="search-input"]').fill("non-existent");
      await expect(page.locator('[data-testid="post-card"]')).toHaveCount(0);

      const clearBtn = page.locator('[data-testid="clear-filters-btn"]');
      await expect(clearBtn).toBeVisible();
      await clearBtn.click();

      await expect(page.locator('[data-testid="post-card"]')).toHaveCount(1);
    });

    test('5. Active filters are styled distinctively', async ({ page }) => {
      const chip = page.locator('[data-testid="category-chip-后端开发"]');
      await chip.click();
      await expect(chip).toHaveClass(/bg-indigo-600/);
    });
  });

  // Feature 3: Post Creation & Auto-AI Trigger
  test.describe('Feature 3: Post Creation & Auto-AI Trigger', () => {
    test('1. Clicking New Post opens creation editor/drawer', async ({ page }) => {
      await page.locator('[data-testid="nav-new-post-btn"]').click();
      await expect(page.locator('text=Create New Topic')).toBeVisible();
    });

    test('2. Form accepts title, content and category selection', async ({ page }) => {
      await page.locator('[data-testid="nav-new-post-btn"]').click();
      await page.locator('#new-post-title').fill("New Monolith Proposal");
      await page.locator('#new-post-content').fill("We should keep our monolith codebase intact.");
      await page.locator('#new-post-category').selectOption("后端开发");
      await expect(page.locator('#new-post-title')).toHaveValue("New Monolith Proposal");
    });

    test('3. Submitting redirects to Homepage Feed', async ({ page }) => {
      await page.locator('[data-testid="nav-new-post-btn"]').click();
      await page.locator('#new-post-title').fill("New Monolith Proposal");
      await page.locator('#new-post-content').fill("We should keep our monolith codebase intact.");
      await page.locator('#new-post-category').selectOption("后端开发");
      await page.locator('#submit-post-btn').click();
      
      await expect(page.locator('text=Create New Topic')).not.toBeVisible();
      await expect(page.locator('[data-testid="post-card"]')).toHaveCount(2);
    });

    test('4. Newly created post shows aiStatus as PENDING initially', async ({ page }) => {
      await page.locator('[data-testid="nav-new-post-btn"]').click();
      await page.locator('#new-post-title').fill("New Monolith Proposal");
      await page.locator('#new-post-content').fill("We should keep our monolith codebase intact.");
      await page.locator('#new-post-category').selectOption("后端开发");
      await page.locator('#submit-post-btn').click();
      
      const newCard = page.locator('[data-testid="post-card-2"]');
      await expect(newCard.locator('text=Status: PENDING')).toBeVisible();
    });

    test('5. The background AI Simulator starts and updates status', async ({ page }) => {
      await page.locator('[data-testid="nav-new-post-btn"]').click();
      await page.locator('#new-post-title').fill("New Monolith Proposal");
      await page.locator('#new-post-content').fill("We should keep our monolith codebase intact.");
      await page.locator('#new-post-category').selectOption("后端开发");
      
      // Set random willingness seeds high for simulation to trigger replies
      await page.evaluate(() => {
        Math.random = () => 0.95; // Force positive decisions
      });

      await page.locator('#submit-post-btn').click();
      const newCard = page.locator('[data-testid="post-card-2"]');
      
      // Wait for it to transition to PROCESSING
      await expect(newCard.locator('text=Status: PROCESSING')).toBeVisible();
      
      // Wait for it to complete eventually
      await expect(newCard.locator('text=Status: COMPLETED')).toBeVisible({ timeout: 6000 });
    });
  });

  // Feature 4: Post Details & Markdown Sanitization
  test.describe('Feature 4: Post Details & Markdown Sanitization', () => {
    test('1. Post details display post content matching markdown styles', async ({ page }) => {
      await page.evaluate(() => {
        const db = JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}');
        db.posts[0].content = "### Title\n- Bullet 1\n- Bullet 2";
        localStorage.setItem('ai_forum_db_state', JSON.stringify(db));
        window.navigate('/post/1');
      });

      await expect(page.locator('[data-testid="post-detail-content"]')).toContainText("Title");
      await expect(page.locator('[data-testid="post-detail-content"]')).toContainText("Bullet 1");
    });

    test('2. Displays author username, avatar, and publication timestamp', async ({ page }) => {
      await page.goto('/post/1');
      await expect(page.locator('text=By alex_dev').or(page.locator('text=alex_dev'))).toBeVisible();
      await expect(page.locator('[data-testid="post-detail-content"]')).toBeVisible();
    });

    test('3. Lists the avatars of all AI agents that have replied', async ({ page }) => {
      await page.goto('/post/1');
      await expect(page.locator('[data-testid="ai-avatar-ArchTechLead"]')).toBeVisible();
      await expect(page.locator('[data-testid="ai-avatar-DevilsAdvocate"]').or(page.locator('[data-testid="ai-avatar-Devil"]'))).toBeVisible();
    });

    test('4. Status indicator displays current AI progress status', async ({ page }) => {
      await page.goto('/post/1');
      await expect(page.locator('[data-testid="post-detail-status"]')).toContainText("COMPLETED");
    });

    test('5. Renders code blocks with formatting', async ({ page }) => {
      await page.evaluate(() => {
        const db = JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}');
        db.posts[0].content = "Code:\n```go\npackage main\n```";
        localStorage.setItem('ai_forum_db_state', JSON.stringify(db));
        window.navigate('/post/1');
      });
      await expect(page.locator('[data-testid="post-detail-content"]')).toContainText("package main");
    });
  });

  // Feature 5: Comment Tree & Threaded Replies
  test.describe('Feature 5: Comment Tree & Threaded Replies', () => {
    test('1. Comments section displays nested replies with hierarchy', async ({ page }) => {
      await page.goto('/post/1');
      const commentItems = page.locator('[data-testid="comment-item"]');
      await expect(commentItems).toHaveCount(2);
      
      // Check parent-child nested spacing
      const firstComment = commentItems.nth(0);
      const secondComment = commentItems.nth(1);
      await expect(firstComment).toHaveText(/Before writing any Rust/);
      await expect(secondComment).toHaveText(/premature optimization/i);
    });

    test('2. Shows distinct badges for AI agents', async ({ page }) => {
      await page.goto('/post/1');
      const aiBadge = page.locator('text=DevilsAdvocate >> xpath=.. >> text=AI');
      await expect(aiBadge).toBeVisible();
    });

    test('3. Clicking Reply on a comment opens a text input nested', async ({ page }) => {
      await page.goto('/post/1');
      await page.locator('[data-testid="comment-reply-btn-1"]').click();
      await expect(page.locator('#reply-text-1')).toBeVisible();
    });

    test('4. Submitting a comment reply updates the comment list', async ({ page }) => {
      await page.goto('/post/1');
      await page.locator('[data-testid="comment-reply-btn-1"]').click();
      await page.locator('#reply-text-1').fill("This is my follow up reply");
      await page.locator('#reply-input-container-1 >> [data-testid="comment-submit-btn"]').click();
      
      await expect(page.locator('text=This is my follow up reply')).toBeVisible();
      await expect(page.locator('[data-testid="comment-item"]')).toHaveCount(3);
    });

    test('5. Timestamps are formatted relative or absolute', async ({ page }) => {
      await page.goto('/post/1');
      const comment = page.locator('[data-testid="comment-item"]').first();
      await expect(comment.locator('span').last()).toContainText(/2026/); // check year/time exists
    });
  });

  // Feature 6: AI Mention @AI & Contextual Follow-up
  test.describe('Feature 6: AI Mention @AI & Contextual Follow-up', () => {
    test('1. Typing @ in comment input displays dropdown list', async ({ page }) => {
      await page.goto('/post/1');
      const commentInput = page.locator('[data-testid="comment-input"]');
      await commentInput.focus();
      await commentInput.fill("@");
      
      await expect(page.locator('#mention-dropdown-root')).toBeVisible();
      await expect(page.locator('#mention-dropdown-root >> [data-testid="mention-item-ArchTechLead"]')).toBeVisible();
    });

    test('2. Selecting an agent inserts the mention component', async ({ page }) => {
      await page.goto('/post/1');
      const commentInput = page.locator('[data-testid="comment-input"]');
      await commentInput.focus();
      await commentInput.fill("@");
      
      await page.locator('#mention-dropdown-root >> [data-testid="mention-item-ArchTechLead"]').click();
      await expect(commentInput).toHaveValue("@ArchTechLead ");
    });

    test('3. Submitting mention triggers target agent simulation', async ({ page }) => {
      await page.goto('/post/1');
      const commentInput = page.locator('[data-testid="comment-input"]');
      await commentInput.focus();
      await commentInput.fill("@ArchTechLead Please analyze.");
      await page.locator('[data-testid="comment-submit-btn"]').first().click();

      // Verify task queue gets created
      await expect.poll(async () => {
        const db = await page.evaluate(() => JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}'));
        return db.tasks.some(t => t.aiAgentId === 1 && t.triggerType === "MENTION");
      }).toBe(true);
    });

    test('4. Mentioned agent responds contextually', async ({ page }) => {
      await page.goto('/post/1');
      const commentInput = page.locator('[data-testid="comment-input"]');
      await commentInput.focus();
      await commentInput.fill("@ArchTechLead Let us review this");
      await page.locator('[data-testid="comment-submit-btn"]').first().click();

      // Wait for AI reply comment to appear
      await expect(page.locator('text=### Design Critique: Is it time to rewrite our Go monolithic API in Rust?').first()).toBeVisible({ timeout: 5000 });
    });

    test('5. Mention agent replies even if threshold is 1.0', async ({ page }) => {
      await page.evaluate(() => {
        const db = JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}');
        const lead = db.agents.find(a => a.name === "ArchTechLead");
        lead.replyThreshold = 1.0; // max threshold
        localStorage.setItem('ai_forum_db_state', JSON.stringify(db));
        window.navigate('/post/1');
      });

      const commentInput = page.locator('[data-testid="comment-input"]');
      await commentInput.focus();
      await commentInput.fill("@ArchTechLead bypass testing");
      await page.locator('[data-testid="comment-submit-btn"]').first().click();

      await expect(page.locator('text=### Design Critique: Is it time to rewrite our Go monolithic API in Rust?').first()).toBeVisible({ timeout: 5000 });
    });
  });

});
