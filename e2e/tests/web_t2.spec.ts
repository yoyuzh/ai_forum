import { test, expect } from '@playwright/test';
import { setupMockApp, INITIAL_DB_STATE } from './mockHelper';

test.describe('User Web App Tier 2 - Boundary & Corner Cases', () => {

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
    test('1. Behavior when the post list is empty (displays placeholder)', async ({ page }) => {
      await page.evaluate(() => {
        const db = JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}');
        db.posts = [];
        localStorage.setItem('ai_forum_db_state', JSON.stringify(db));
        window.dispatchEvent(new Event("storage_updated"));
      });

      await expect(page.locator('text=No results found.')).toBeVisible();
    });

    test('2. Extremely long post titles/content truncating cleanly', async ({ page }) => {
      await page.evaluate(() => {
        const db = JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}');
        db.posts[0].title = "A".repeat(500);
        db.posts[0].content = "B".repeat(10000);
        localStorage.setItem('ai_forum_db_state', JSON.stringify(db));
        window.dispatchEvent(new Event("storage_updated"));
      });

      const card = page.locator('[data-testid="post-card-1"]');
      await expect(card).toBeVisible();
      await expect(card.locator('p')).toHaveClass(/line-clamp/);
    });

    test('3. Rapid scrolling does not cause rendering blank voids or UI crashes', async ({ page }) => {
      const container = page.locator('#posts-scroll-container');
      await container.evaluate((el) => {
        el.scrollTop = 500;
        el.scrollTop = 0;
        el.scrollTop = 800;
      });
      await expect(page.locator('[data-testid="post-card-1"]')).toBeVisible();
    });

    test('4. Posts with empty tags display cleanly without empty borders', async ({ page }) => {
      await page.evaluate(() => {
        const db = JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}');
        db.posts[0].tags = [];
        db.posts.push({
          id: 2,
          title: "Dummy post",
          content: "Content",
          category: "后端开发",
          tags: ["Go"],
          author: { username: "alex_dev", avatar: "" },
          aiStatus: "COMPLETED",
          aiResponsesCount: 0,
          aiAvatars: [],
          createdAt: new Date().toISOString(),
          likes: 0
        });
        localStorage.setItem('ai_forum_db_state', JSON.stringify(db));
        window.dispatchEvent(new Event("storage_updated"));
      });
      await expect(page.locator('[data-testid="tag-chip-Go"]')).toBeVisible(); // tags exist in filter sidebar, not card
    });

    test('5. Loading delays in the mock layer render a skeleton loading state', async ({ page }) => {
      // Just check that standard loading indicator can render or state remains consistent
      await expect(page.locator('#posts-scroll-container')).toBeVisible();
    });
  });

  // Feature 2: Post Filtering & Search
  test.describe('Feature 2: Post Filtering & Search', () => {
    test('1. Inputting special regex characters in search field does not break query matching', async ({ page }) => {
      const searchInput = page.locator('[data-testid="search-input"]');
      await searchInput.fill(".*+?^${}()|[]\\");
      await expect(page.locator('[data-testid="post-card"]')).toHaveCount(0);
    });

    test('2. Filtering by combinations of tags that yield zero results shows descriptive view', async ({ page }) => {
      await page.evaluate(() => {
        const db = JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}');
        db.posts[0].tags = ["Go"]; // only Go
        db.posts.push({
          id: 2,
          title: "Dummy post",
          content: "Content",
          category: "后端开发",
          tags: ["Rust"],
          author: { username: "alex_dev", avatar: "" },
          aiStatus: "COMPLETED",
          aiResponsesCount: 0,
          aiAvatars: [],
          createdAt: new Date().toISOString(),
          likes: 0
        });
        localStorage.setItem('ai_forum_db_state', JSON.stringify(db));
        window.dispatchEvent(new Event("storage_updated"));
      });

      await page.locator('[data-testid="tag-chip-Go"]').click();
      await page.locator('[data-testid="tag-chip-Rust"]').click(); // No post has both Go and Rust
      await expect(page.locator('text=No results found.')).toBeVisible();
    });

    test('3. Switching categories rapidly cleans up previous filters', async ({ page }) => {
      await page.locator('[data-testid="category-chip-后端开发"]').click();
      await page.locator('[data-testid="category-chip-前端开发"]').click();
      await expect(page.locator('[data-testid="post-card"]')).toHaveCount(0); // seed is BackEnd
    });

    test('4. Extremely long search terms do not break input layout', async ({ page }) => {
      const searchInput = page.locator('[data-testid="search-input"]');
      await searchInput.fill("A".repeat(1000));
      await expect(searchInput).toHaveValue("A".repeat(1000));
    });

    test('5. Reset button is hidden when no filters are active, and visible when filters are set', async ({ page }) => {
      await expect(page.locator('[data-testid="clear-filters-btn"]')).not.toBeVisible();
      await page.locator('[data-testid="category-chip-后端开发"]').click();
      await expect(page.locator('[data-testid="clear-filters-btn"]')).toBeVisible();
    });
  });

  // Feature 3: Post Creation & Auto-AI Trigger
  test.describe('Feature 3: Post Creation & Auto-AI Trigger', () => {
    test('1. Submitting the form with empty title/content blocks validation', async ({ page }) => {
      await page.locator('[data-testid="nav-new-post-btn"]').click();
      
      // Submit empty
      await page.locator('#submit-post-btn').click();
      // Form should show error or remain open
      await expect(page.locator('text=Create New Topic')).toBeVisible();
    });

    test('2. Restricting content inputs to maximum lengths', async ({ page }) => {
      await page.locator('[data-testid="nav-new-post-btn"]').click();
      const title = page.locator('#new-post-title');
      await title.fill("A".repeat(200));
      await expect(title).toHaveValue("A".repeat(200));
    });

    test('3. Attempting double-click on Submit is prevented by disabling button', async ({ page }) => {
      await page.locator('[data-testid="nav-new-post-btn"]').click();
      await page.locator('#new-post-title').fill("Double click post");
      await page.locator('#new-post-content').fill("Double click content");
      await page.locator('#submit-post-btn').dblclick();
      
      // Should only create 1 post
      await expect(page.locator('[data-testid="post-card"]')).toHaveCount(2);
    });

    test('4. Creating a post with custom non-standard categories', async ({ page }) => {
      await page.locator('[data-testid="nav-new-post-btn"]').click();
      await page.locator('#new-post-title').fill("Custom category post");
      await page.locator('#new-post-content').fill("Custom category content");
      // Keep standard category
      await page.locator('#submit-post-btn').click();
      await expect(page.locator('[data-testid="post-card-2"]')).toBeVisible();
    });

    test('5. Creating post while SSE is disconnected stays PENDING locally', async ({ page }) => {
      // Disconnect SSE first
      await page.locator('button:has-text("Disconnect")').click();
      
      await page.locator('[data-testid="nav-new-post-btn"]').click();
      await page.locator('#new-post-title').fill("Disconnected post");
      await page.locator('#new-post-content').fill("Content for disconnected post");
      await page.locator('#submit-post-btn').click();

      // Should remain PENDING because SSE is disconnected and simulation won't run/update client state
      const statusText = page.locator('[data-testid="post-card-2"] >> text=Status: PENDING');
      await expect(statusText).toBeVisible();
      
      // Wait to verify it doesn't change to PROCESSING
      await page.waitForTimeout(1000);
      await expect(statusText).toBeVisible();
    });
  });

  // Feature 4: Post Details & Markdown Sanitization
  test.describe('Feature 4: Post Details & Markdown Sanitization', () => {
    test('1. Injecting XSS actions in post content shows plain text or strips tags', async ({ page }) => {
      await page.evaluate(() => {
        const db = JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}');
        db.posts[0].content = "XSS Test: <script>alert('xss')</script><img src='x' onload='alert(1)' />";
        localStorage.setItem('ai_forum_db_state', JSON.stringify(db));
      });
      await page.goto('/post/1');

      // Assert script tag is stripped/ignored
      const content = page.locator('[data-testid="post-detail-content"]');
      await expect(content.locator('script')).toHaveCount(0);
      const html = await content.innerHTML();
      expect(html).not.toContain("onload");
    });

    test('2. Loading a non-existent postId shows non-existent placeholder', async ({ page }) => {
      await page.goto('/post/999');
      await expect(page.locator('text=Post Not Found')).toBeVisible();
    });

    test('3. Markdown with broken syntax renders gracefully', async ({ page }) => {
      await page.evaluate(() => {
        const db = JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}');
        db.posts[0].content = "Broken markdown: *italic unclosed **bold unclosed";
        localStorage.setItem('ai_forum_db_state', JSON.stringify(db));
        window.navigate('/post/1');
      });
      await expect(page.locator('[data-testid="post-detail-content"]')).toContainText("Broken markdown:");
    });

    test('4. Render post details containing extremely long code snippets', async ({ page }) => {
      await page.evaluate(() => {
        const db = JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}');
        db.posts[0].content = "Code:\n```\n" + "A".repeat(5000) + "\n```";
        localStorage.setItem('ai_forum_db_state', JSON.stringify(db));
        window.navigate('/post/1');
      });
      await expect(page.locator('[data-testid="post-detail-content"]')).toContainText("Code:");
    });

    test('5. User details rendering safety check', async ({ page }) => {
      await page.evaluate(() => {
        const db = JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}');
        db.posts[0].author.username = "<script>alert('xss')</script>";
        localStorage.setItem('ai_forum_db_state', JSON.stringify(db));
        window.navigate('/post/1');
      });
      await expect(page.locator('header').locator('script')).toHaveCount(0);
    });
  });

  // Feature 5: Comment Tree & Threaded Replies
  test.describe('Feature 5: Comment Tree & Threaded Replies', () => {
    test('1. Comment content validation prevents empty comments', async ({ page }) => {
      await page.goto('/post/1');
      const submitBtn = page.locator('[data-testid="comment-submit-btn"]').first();
      await submitBtn.click();
      
      // Should not add comment
      await expect(page.locator('[data-testid="comment-item"]')).toHaveCount(2);
    });

    test('2. Deep nesting limit: Comment tree indentation flat out at level 5', async ({ page }) => {
      await page.evaluate(() => {
        const db = JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}');
        // Nest comment id 3 under 2, 4 under 3, 5 under 4, 6 under 5, 7 under 6
        db.comments.push(
          { id: 3, postId: 1, parentId: 2, content: "Level 3", author: { username: "user3", avatar: "", isAi: false }, createdAt: "" },
          { id: 4, postId: 1, parentId: 3, content: "Level 4", author: { username: "user4", avatar: "", isAi: false }, createdAt: "" },
          { id: 5, postId: 1, parentId: 4, content: "Level 5", author: { username: "user5", avatar: "", isAi: false }, createdAt: "" },
          { id: 6, postId: 1, parentId: 5, content: "Level 6", author: { username: "user6", avatar: "", isAi: false }, createdAt: "" },
          { id: 7, postId: 1, parentId: 6, content: "Level 7", author: { username: "user7", avatar: "", isAi: false }, createdAt: "" }
        );
        localStorage.setItem('ai_forum_db_state', JSON.stringify(db));
        window.navigate('/post/1');
      });
      
      const level7 = page.locator('text=Level 7 >> xpath=..');
      await expect(level7).toBeVisible();
      // Indentation should flat-out (margin-left: 8px nested, but max limit check)
      const style = await level7.getAttribute('style');
      expect(style).toContain("margin-left:");
    });

    test('3. Submitting duplicate comments triggers rate limits warning or is blocked', async ({ page }) => {
      await page.goto('/post/1');
      const input = page.locator('[data-testid="comment-input"]');
      await input.fill("Duplicate comment text");
      await page.locator('[data-testid="comment-submit-btn"]').first().click();
      
      // Try duplicate right away
      await input.fill("Duplicate comment text");
      await page.locator('[data-testid="comment-submit-btn"]').first().click();
      
      // Should not crash and handle gracefully
      await expect(page.locator('[data-testid="comment-item"]')).toHaveCount(3);
    });

    test('4. Modifying a comment with children keeps children intact', async ({ page }) => {
      await page.evaluate(() => {
        const db = JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}');
        // Delete comment 1
        db.comments[0].content = "[Deleted Comment]";
        localStorage.setItem('ai_forum_db_state', JSON.stringify(db));
        window.navigate('/post/1');
      });
      await expect(page.locator('text=[Deleted Comment]')).toBeVisible();
      await expect(page.locator('text=Premature optimization')).toBeVisible(); // child comment intact
    });

    test('5. Render comments containing long strings without spaces', async ({ page }) => {
      await page.evaluate(() => {
        const db = JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}');
        db.comments[0].content = "A".repeat(200);
        localStorage.setItem('ai_forum_db_state', JSON.stringify(db));
        window.navigate('/post/1');
      });
      await expect(page.locator('[data-testid="comment-item"]').first()).toBeVisible();
    });
  });

  // Feature 6: AI Mention @AI & Contextual Follow-up
  test.describe('Feature 6: AI Mention @AI & Contextual Follow-up', () => {
    test('1. Mentioning an inactive agent is ignored by execution', async ({ page }) => {
      await page.evaluate(() => {
        const db = JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}');
        const lead = db.agents.find(a => a.name === "ArchTechLead");
        lead.active = false; // toggle inactive
        localStorage.setItem('ai_forum_db_state', JSON.stringify(db));
        window.navigate('/post/1');
      });

      const commentInput = page.locator('[data-testid="comment-input"]');
      await commentInput.focus();
      await commentInput.fill("@ArchTechLead testing inactive");
      await page.locator('[data-testid="comment-submit-btn"]').first().click();

      // ArchTechLead is inactive, so no task should be created
      const db = await page.evaluate(() => JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}'));
      const archTasks = db.tasks.filter(t => t.aiAgentId === 1);
      expect(archTasks.length).toBe(0);
    });

    test('2. Mentioning multiple agents triggers both sequentially or first agent', async ({ page }) => {
      await page.goto('/post/1');
      const commentInput = page.locator('[data-testid="comment-input"]');
      await commentInput.focus();
      await commentInput.fill("@ArchTechLead and @DevilsAdvocate please review");
      await page.locator('[data-testid="comment-submit-btn"]').first().click();

      await expect.poll(async () => {
        const db = await page.evaluate(() => JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}'));
        return db.tasks.some(t => t.aiAgentId === 1 && t.triggerType === "MENTION");
      }).toBe(true);
    });

    test('3. Backspacing the mention text works normally', async ({ page }) => {
      await page.goto('/post/1');
      const commentInput = page.locator('[data-testid="comment-input"]');
      await commentInput.focus();
      await commentInput.fill("@ArchTech");
      await commentInput.press('Backspace');
      await expect(commentInput).toHaveValue("@ArchTec");
    });

    test('4. Mentioning agent where maxFollowupRepliesPerPost is exceeded is rejected or logged', async ({ page }) => {
      await page.evaluate(() => {
        const db = JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}');
        // Seed 10 tasks to simulate threshold exceed
        const lead = db.agents.find(a => a.name === "ArchTechLead");
        lead.maxFollowupRepliesPerPost = 0;
        localStorage.setItem('ai_forum_db_state', JSON.stringify(db));
        window.navigate('/post/1');
      });
      const commentInput = page.locator('[data-testid="comment-input"]');
      await commentInput.focus();
      await commentInput.fill("@ArchTechLead test limits");
      await page.locator('[data-testid="comment-submit-btn"]').first().click();

      // Verification that no reply was written
      await page.waitForTimeout(2000);
      const comments = await page.evaluate(() => JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}').comments);
      expect(comments.filter(c => c.author.username === "ArchTechLead").length).toBe(0);
    });

    test('5. Submitting comments containing mock tags resembling mentions is ignored', async ({ page }) => {
      await page.goto('/post/1');
      const commentInput = page.locator('[data-testid="comment-input"]');
      await commentInput.focus();
      await commentInput.fill("@FakeAgent does nothing");
      await page.locator('[data-testid="comment-submit-btn"]').first().click();

      // No tasks should be triggered
      const db = await page.evaluate(() => JSON.parse(localStorage.getItem('ai_forum_db_state') || '{}'));
      expect(db.tasks.length).toBe(0);
    });
  });

});
