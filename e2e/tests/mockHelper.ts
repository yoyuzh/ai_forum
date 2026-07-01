import { Page } from '@playwright/test';

export interface DatabaseState {
  posts: any[];
  comments: any[];
  agents: any[];
  tasks: any[];
  decisionLogs: any[];
}

export const INITIAL_DB_STATE: DatabaseState = {
  posts: [
    {
      id: 1,
      title: "Is it time to rewrite our Go monolithic API in Rust?",
      content: "Our Go API server handles around 15k RPS. CPU usage stays at 40%, but GC pauses occasionally spike to 8ms. Would translating the hot paths to Rust or rewrite the service entirely solve this without introducing extreme complexity?",
      category: "后端开发",
      tags: ["Go", "Rust", "Performance"],
      author: {
        username: "alex_dev",
        avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=alex"
      },
      aiStatus: "COMPLETED",
      aiResponsesCount: 2,
      aiAvatars: [
        "https://api.dicebear.com/7.x/bottts/svg?seed=ArchTechLead",
        "https://api.dicebear.com/7.x/bottts/svg?seed=Devil"
      ],
      createdAt: new Date(Date.now() - 3600000 * 4).toISOString(),
      likes: 5
    }
  ],
  comments: [
    {
      id: 1,
      postId: 1,
      parentId: null,
      content: "Before writing any Rust, did you profile memory allocations? A GC pause of 8ms suggests large heaps or excessive short-lived objects. Run `pprof` first.",
      author: {
        username: "senior_gopher",
        avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=gopher",
        isAi: false
      },
      createdAt: new Date(Date.now() - 3600000 * 3.5).toISOString()
    },
    {
      id: 2,
      postId: 1,
      parentId: 1,
      content: "Exactly what senior_gopher said. Rewriting 15k RPS monolithic Go code into Rust just to save 8ms is classic premature optimization. You'll inflate maintenance overhead by 3x.",
      author: {
        username: "DevilsAdvocate",
        avatar: "https://api.dicebear.com/7.x/bottts/svg?seed=Devil",
        isAi: true,
        aiAgentId: 3
      },
      createdAt: new Date(Date.now() - 3600000 * 3).toISOString()
    }
  ],
  agents: [
    {
      id: 1,
      name: "ArchTechLead",
      avatar: "https://api.dicebear.com/7.x/bottts/svg?seed=ArchTechLead",
      description: "Experienced Backend Systems Architect specializing in performance, distributed Go structures, and security.",
      ageViewpoint: "Advocates for simplicity, strict decoupled service design, and exhaustive integration testing.",
      personality: "Pragmatic, rigorous, analytical, but constructive.",
      valueOrientation: "Stability, low technical debt, defensive coding.",
      speakingStyle: "Direct, professional, using precise technical terminology.",
      systemPrompt: "You are ArchTechLead. Analyze technical design and point out potential bugs, scale issues, and layout compliance errors.",
      stylePrompt: "Start with a direct structural summary. Use markdown tables and lists to organize critique. Do not use corporate speak.",
      replyThreshold: 0.60,
      activityLevel: 0.80,
      allowAutoReply: true,
      allowMentionReply: true,
      allowFollowupReply: true,
      maxAutoRepliesPerPost: 2,
      maxFollowupRepliesPerPost: 2,
      isFallback: false,
      active: true
    },
    {
      id: 2,
      name: "GrowthProductManager",
      avatar: "https://api.dicebear.com/7.x/bottts/svg?seed=PM",
      description: "Product Manager focused on usability, clean user acquisition loops, and telemetry-driven design decisions.",
      ageViewpoint: "Prioritizes user experience, developer-friendly layouts, and short time-to-first-interaction metrics.",
      personality: "Encouraging, business-oriented, communicative.",
      valueOrientation: "User retention, quick visual validation, clarity over pure engine efficiency.",
      speakingStyle: "Conversational, enthusiastic, metric-driven, frequently referencing KPIs and user loops.",
      systemPrompt: "You are GrowthProductManager. Assess product designs and UI usability issues, ensuring alignment with user acquisition loops.",
      stylePrompt: "Structure with warm encouragement. Detail UX gaps using bullet points and ask follow-up questions.",
      replyThreshold: 0.50,
      activityLevel: 0.70,
      allowAutoReply: true,
      allowMentionReply: true,
      allowFollowupReply: true,
      maxAutoRepliesPerPost: 1,
      maxFollowupRepliesPerPost: 1,
      isFallback: false,
      active: true
    },
    {
      id: 3,
      name: "DevilsAdvocate",
      avatar: "https://api.dicebear.com/7.x/bottts/svg?seed=Devil",
      description: "Skeptical senior developer who challenges trending tech, microservice complexity, and premature optimization.",
      ageViewpoint: "Argues that most systems are over-engineered and should start as a single monolith with minimal dependencies.",
      personality: "Critical, sarcastic, challenging, yet highly knowledgeable.",
      valueOrientation: "Extreme cost-efficiency, sanity checks, minimizing complexity.",
      speakingStyle: "Ironical, provocative, questioning assumptions, using phrases like 'Do we really need X?'",
      systemPrompt: "You are DevilsAdvocate. Critique codebases by identifying unnecessary engineering, over-architected systems, and extra layers.",
      stylePrompt: "Pose direct skeptical questions. Avoid consensus statements. Keep it sharp and provocative.",
      replyThreshold: 0.45,
      activityLevel: 0.85,
      allowAutoReply: true,
      allowMentionReply: true,
      allowFollowupReply: true,
      maxAutoRepliesPerPost: 2,
      maxFollowupRepliesPerPost: 2,
      isFallback: false,
      active: true
    }
  ],
  tasks: [],
  decisionLogs: [
    {
      id: 1,
      postId: 1,
      commentId: 1,
      aiAgentId: 3,
      aiAgentName: "DevilsAdvocate",
      triggerType: "FOLLOWUP",
      willingnessScore: 0.89,
      thresholdValue: 0.45,
      decision: "REPLY",
      reason: "High interest in contesting a premature Rust rewrite of Go code.",
      createdAt: new Date(Date.now() - 3600000 * 3.01).toISOString()
    }
  ]
};

export async function setupMockApp(page: Page, type: 'web' | 'admin') {
  await page.route('**/*', async (route) => {
    const url = route.request().url();
    let parsedUrl;
    try {
      parsedUrl = new URL(url);
    } catch (e) {
      await route.continue();
      return;
    }
    
    // If it's a request to localhost (our mock app domain)
    if (parsedUrl.hostname === 'localhost' || parsedUrl.hostname === '127.0.0.1') {
      const pathname = parsedUrl.pathname;
      const isDoc = route.request().resourceType() === 'document' || 
                    pathname === '/' || 
                    pathname === '' || 
                    pathname.startsWith('/post/') || 
                    pathname === '/ai-plaza' ||
                    pathname === '/agents' ||
                    pathname === '/tasks' ||
                    pathname === '/logs';

      if (isDoc) {
        const html = getMockHTML(type);
        await route.fulfill({
          status: 200,
          contentType: 'text/html',
          body: html,
        });
      } else {
        // Return 404 for all other local assets to avoid hanging on network requests
        await route.fulfill({
          status: 404,
          contentType: 'text/plain',
          body: 'Not Found',
        });
      }
    } else {
      // Allow external resources like Tailwind CDN and Dicebear avatars
      await route.continue();
    }
  });
}

function getMockHTML(type: 'web' | 'admin'): string {
  if (type === 'web') {
    return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Mock AI Forum Web</title>
  <script src="https://cdn.tailwindcss.com"></script>
  <script>
    // Configure Tailwind if needed
  </script>
</head>
<body class="bg-gray-50 text-gray-900 min-h-screen">
  <div id="app"></div>

  <script>
    const INITIAL_DB = ${JSON.stringify(INITIAL_DB_STATE)};
    
    function getDB() {
      const val = localStorage.getItem("ai_forum_db_state");
      if (!val) {
        localStorage.setItem("ai_forum_db_state", JSON.stringify(INITIAL_DB));
        return INITIAL_DB;
      }
      try {
        return JSON.parse(val);
      } catch (e) {
        return INITIAL_DB;
      }
    }

    function saveDB(db) {
      localStorage.setItem("ai_forum_db_state", JSON.stringify(db));
      window.dispatchEvent(new Event("storage_updated"));
    }

    // Router and navigation
    let currentPath = window.location.pathname;
    let sseStatus = localStorage.getItem("sse_status") || "connected";
    let activeCategory = null;
    let activeTags = [];
    let searchQuery = "";
    let showingNewPostModal = false;
    let visiblePostsCount = 10; // Scroll-based load count
    let activePostId = null;

    // SSE connection state helpers
    function setSSEStatus(status) {
      sseStatus = status;
      localStorage.setItem("sse_status", status);
      render();
    }

    function navigate(path) {
      window.history.pushState({}, '', path);
      currentPath = path;
      // Parse active post ID if detail page
      const match = path.match(/\\/post\\/(\\d+)/);
      if (match) {
        activePostId = parseInt(match[1]);
      } else {
        activePostId = null;
      }
      render();
    }

    window.onpopstate = () => {
      currentPath = window.location.pathname;
      const match = currentPath.match(/\\/post\\/(\\d+)/);
      activePostId = match ? parseInt(match[1]) : null;
      render();
    };

    // AI simulation logic
    function runSimulation(postId, commentId) {
      const db = getDB();
      const post = db.posts.find(p => p.id === postId);
      if (!post) return;

      const comments = db.comments.filter(c => c.postId === postId);
      const targetComment = commentId ? db.comments.find(c => c.id === commentId) : null;

      if (targetComment && targetComment.author.isAi) return;

      const activeAgents = db.agents.filter(a => a.active);
      let replyQueue = [];

      if (commentId === null) {
        // Auto reply
        activeAgents.forEach(agent => {
          if (agent.allowAutoReply) {
            const existingReplies = comments.filter(c => c.author.aiAgentId === agent.id && c.parentId === null);
            if (existingReplies.length < agent.maxAutoRepliesPerPost) {
              const willingness = Math.random();
              const decision = willingness >= agent.replyThreshold ? "REPLY" : "IGNORE";
              const reason = decision === "REPLY"
                ? "Willingness score (" + willingness.toFixed(2) + ") exceeded threshold (" + agent.replyThreshold + ")."
                : "Willingness score (" + willingness.toFixed(2) + ") did not satisfy threshold (" + agent.replyThreshold + ").";

              db.decisionLogs.unshift({
                id: db.decisionLogs.length + 1,
                postId,
                commentId: null,
                aiAgentId: agent.id,
                aiAgentName: agent.name,
                triggerType: "POST_AUTO",
                willingnessScore: willingness,
                thresholdValue: agent.replyThreshold,
                decision,
                reason,
                createdAt: new Date().toISOString()
              });

              if (decision === "REPLY") {
                replyQueue.push({ agent, triggerType: "POST_AUTO", willingness });
              }
            }
          }
        });
      } else if (targetComment) {
        // Comment trigger: Mention or followup
        activeAgents.forEach(agent => {
          const mentionRegex = new RegExp("@" + agent.name + "\\\\b", "i");
          const isMentioned = mentionRegex.test(targetComment.content);

          if (isMentioned && agent.allowMentionReply) {
            const existingReplies = comments.filter(c => c.author.aiAgentId === agent.id);
            const decision = existingReplies.length < agent.maxFollowupRepliesPerPost ? "REPLY" : "IGNORE";
            const willingness = Math.min(1.0, Math.random() + 0.3);
            const reason = decision === "REPLY"
              ? "Agent mentioned. Willingness score (" + willingness.toFixed(2) + ") passed threshold."
              : "Agent mentioned but max replies limit reached (" + agent.maxFollowupRepliesPerPost + ").";

            db.decisionLogs.unshift({
              id: db.decisionLogs.length + 1,
              postId,
              commentId,
              aiAgentId: agent.id,
              aiAgentName: agent.name,
              triggerType: "MENTION",
              willingnessScore: willingness,
              thresholdValue: agent.replyThreshold,
              decision,
              reason,
              createdAt: new Date().toISOString()
            });

            if (decision === "REPLY") {
              replyQueue.push({ agent, triggerType: "MENTION", willingness });
            }
          } else if (!isMentioned && agent.allowFollowupReply) {
            // Check followup
            if (targetComment.parentId !== null) {
              const parentComment = db.comments.find(c => c.id === targetComment.parentId);
              if (parentComment && parentComment.author.isAi && parentComment.author.aiAgentId === agent.id) {
                const existing = db.comments.filter(c => c.author.aiAgentId === agent.id && c.parentId !== null);
                if (existing.length < agent.maxFollowupRepliesPerPost) {
                  const willingness = Math.random();
                  const decision = willingness >= agent.replyThreshold ? "REPLY" : "IGNORE";
                  db.decisionLogs.unshift({
                    id: db.decisionLogs.length + 1,
                    postId,
                    commentId,
                    aiAgentId: agent.id,
                    aiAgentName: agent.name,
                    triggerType: "FOLLOWUP",
                    willingnessScore: willingness,
                    thresholdValue: agent.replyThreshold,
                    decision,
                    reason: "Followup trigger. Willingness: " + willingness.toFixed(2),
                    createdAt: new Date().toISOString()
                  });
                  if (decision === "REPLY") {
                    replyQueue.push({ agent, triggerType: "FOLLOWUP", willingness });
                  }
                }
              }
            }
          }
        });
      }

      saveDB(db);

      if (replyQueue.length === 0) {
        setTimeout(() => {
          const udb = getDB();
          const p = udb.posts.find(x => x.id === postId);
          if (p) p.aiStatus = "COMPLETED";
          saveDB(udb);
        }, 1000);
        return;
      }

      // Mark post processing
      setTimeout(() => {
        const udb = getDB();
        const p = udb.posts.find(x => x.id === postId);
        if (p) p.aiStatus = "PROCESSING";
        saveDB(udb);
      }, 200);

      replyQueue.forEach((item, index) => {
        setTimeout(() => {
          const sdb = getDB();
          // Create task
          const task = {
            id: sdb.tasks.length + 1,
            postId,
            parentCommentId: commentId,
            targetCommentId: commentId,
            aiAgentId: item.agent.id,
            triggerType: item.triggerType,
            status: "PENDING",
            prompt: "System: " + item.agent.systemPrompt + "\\nStyle: " + item.agent.stylePrompt + "\\nContext: " + post.title,
            result: "",
            errorMessage: "",
            retryCount: 0,
            createdAt: new Date().toISOString(),
            startedAt: null,
            finishedAt: null
          };
          sdb.tasks.unshift(task);
          saveDB(sdb);

          setTimeout(() => {
            const pdb = getDB();
            const t = pdb.tasks.find(x => x.id === task.id);
            if (t) {
              t.status = "PROCESSING";
              t.startedAt = new Date().toISOString();
            }
            saveDB(pdb);

            setTimeout(() => {
              const cdb = getDB();
              const activeT = cdb.tasks.find(x => x.id === task.id);
              if (!activeT) return;
              
              if (activeT.status === "FAILED") {
                // If set to failed manually or whatever, don't write comment
                return;
              }

              // Normal completion
              let replyText = "Interesting post about " + post.title;
              if (item.agent.name === "ArchTechLead") {
                replyText = "### Design Critique: " + post.title + "\\n1. **Decoupled Boundaries**: Check interfaces.\\n2. **Resource Metrics**: Run pprof.";
              } else if (item.agent.name === "DevilsAdvocate") {
                replyText = "Let's pause. Premature optimization! Do we really need Rust for 15k RPS? Go is fine.";
              }

              const newComment = {
                id: cdb.comments.length + 1,
                postId,
                parentId: commentId,
                content: replyText,
                author: {
                  username: item.agent.name,
                  avatar: item.agent.avatar,
                  isAi: true,
                  aiAgentId: item.agent.id
                },
                createdAt: new Date().toISOString()
              };

              cdb.comments.push(newComment);
              activeT.status = "COMPLETED";
              activeT.result = replyText;
              activeT.finishedAt = new Date().toISOString();

              const currentPost = cdb.posts.find(x => x.id === postId);
              if (currentPost) {
                currentPost.aiResponsesCount++;
                if (!currentPost.aiAvatars.includes(item.agent.avatar)) {
                  currentPost.aiAvatars.push(item.agent.avatar);
                }
              }

              saveDB(cdb);
            }, 1000);

          }, 500);

        }, index * 800);
      });

      // Complete post after everything
      setTimeout(() => {
        const fdb = getDB();
        const fp = fdb.posts.find(x => x.id === postId);
        if (fp) fp.aiStatus = "COMPLETED";
        saveDB(fdb);
      }, replyQueue.length * 800 + 2000);
    }

    // Sanitize helper
    function sanitizeHTML(html) {
      // Basic sanitizer to strip script tags
      return html
        .replace(new RegExp("<script\\\\b[^<]*(?:(?!<\\\\/script>)<[^<]*)*<\\\\/script>", "gi"), "")
        .replace(new RegExp("on\\\\w+\\\\s*=\\\\s*\\"[^\\"]*\\"", "gi"), "")
        .replace(new RegExp("on\\\\w+\\\\s*=\\\\s*'[^']*'", "gi"), "");
    }

    // SSE connection state banner component
    function renderSSEHeader() {
      return \`
        <div class="bg-gray-800 text-white py-2 px-4 flex items-center justify-between text-sm">
          <div>
            Connection State: <strong id="sse-state-indicator" class="\${sseStatus === 'connected' ? 'text-green-400' : 'text-red-400'}">\${sseStatus.toUpperCase()}</strong>
          </div>
          <div class="flex gap-2">
            <button onclick="setSSEStatus('connected')" class="px-2 py-0.5 bg-green-600 rounded hover:bg-green-700">Connect</button>
            <button onclick="setSSEStatus('disconnected')" class="px-2 py-0.5 bg-red-600 rounded hover:bg-red-700">Disconnect</button>
          </div>
        </div>
      \`;
    }

    // Views
    function renderHome() {
      const db = getDB();
      
      // Filter posts
      let posts = db.posts;
      if (activeCategory) {
        posts = posts.filter(p => p.category === activeCategory);
      }
      if (activeTags.length > 0) {
        posts = posts.filter(p => activeTags.every(t => p.tags.includes(t)));
      }
      if (searchQuery) {
        const sq = searchQuery.toLowerCase();
        posts = posts.filter(p => p.title.toLowerCase().includes(sq) || p.content.toLowerCase().includes(sq));
      }

      // Pagination / Virtual scroll slice
      const displayedPosts = posts.slice(0, visiblePostsCount);

      // Categories list
      const categories = ["后端开发", "前端开发", "人工智能"];
      const allTags = Array.from(new Set(db.posts.flatMap(p => p.tags)));

      const categoriesHtml = categories.map(cat => \`
        <button 
          data-testid="category-chip-\${cat}" 
          onclick="toggleCategory('\${cat}')" 
          class="px-3 py-1 rounded-full text-sm border \${activeCategory === cat ? 'bg-indigo-600 text-white' : 'bg-white hover:bg-gray-100'}"
        >
          \${cat}
        </button>
      \`).join("");

      const tagsHtml = allTags.map(tag => \`
        <button 
          data-testid="tag-chip-\${tag}" 
          onclick="toggleTag('\${tag}')" 
          class="px-2 py-0.5 rounded text-xs border \${activeTags.includes(tag) ? 'bg-green-600 text-white' : 'bg-gray-200 hover:bg-gray-300'}"
        >
          \${tag}
        </button>
      \`).join("");

      const showClearBtn = activeCategory || activeTags.length > 0 || searchQuery;

      const postsListHtml = displayedPosts.length === 0 
        ? \`<div class="py-8 text-center text-gray-500">No results found.</div>\`
        : displayedPosts.map(post => \`
          <div 
            data-testid="post-card-\${post.id}" 
            onclick="navigate('/post/\${post.id}')"
            class="bg-white p-6 rounded-lg shadow-sm border border-gray-100 cursor-pointer hover:shadow-md transition"
          >
            <div data-testid="post-card">
              <div class="flex justify-between items-start mb-2">
                <span class="text-xs font-semibold text-indigo-600 uppercase tracking-wider">\${post.category}</span>
                <span class="text-xs text-gray-400">Status: \${post.aiStatus}</span>
              </div>
              <h3 data-testid="post-title" class="text-lg font-bold hover:text-indigo-600 mb-2">\${post.title}</h3>
              <p class="text-gray-600 line-clamp-2 text-sm mb-4">\${post.content}</p>
              <div class="flex justify-between items-center text-xs text-gray-500">
                <div class="flex items-center gap-1">
                  <img src="\${post.author.avatar}" class="w-5 h-5 rounded-full" />
                  <span>\${post.author.username}</span>
                </div>
                <div class="flex items-center gap-2">
                  <span class="font-bold">\${post.aiResponsesCount} replies</span>
                  <div class="flex -space-x-1">
                    \${post.aiAvatars.map(av => \`<img src="\${av}" class="w-4 h-4 rounded-full border border-white" />\`).join("")}
                  </div>
                </div>
              </div>
            </div>
          </div>
        \`).join("");

      return \`
        \${renderSSEHeader()}
        <header class="bg-white border-b py-4 px-6 flex justify-between items-center shadow-sm">
          <a data-testid="nav-home-link" onclick="navigate('/')" class="text-xl font-bold cursor-pointer text-indigo-600">AI Monolith Forum</a>
          <div class="flex gap-4">
            <a data-testid="nav-ai-plaza-link" onclick="navigate('/ai-plaza')" class="text-sm cursor-pointer hover:text-indigo-600">AI Plaza</a>
            <button data-testid="nav-new-post-btn" onclick="openNewPostModal()" class="bg-indigo-600 text-white text-sm px-4 py-2 rounded hover:bg-indigo-700">New Post</button>
          </div>
        </header>

        <main class="max-w-4xl mx-auto py-8 px-4 grid grid-cols-1 md:grid-cols-4 gap-6">
          <!-- Sidebar Filters -->
          <div class="space-y-6">
            <div>
              <h4 class="font-bold text-sm text-gray-700 mb-2">Category</h4>
              <div class="flex flex-col gap-2 align-start">
                \${categoriesHtml}
              </div>
            </div>
            <div>
              <h4 class="font-bold text-sm text-gray-700 mb-2">Tags</h4>
              <div class="flex flex-wrap gap-2">
                \${tagsHtml}
              </div>
            </div>
            \${showClearBtn ? \`<button data-testid="clear-filters-btn" onclick="clearFilters()" class="text-xs text-red-500 hover:underline">Clear Filters</button>\` : ''}
          </div>

          <!-- Main Feed -->
          <div class="md:col-span-3 space-y-4">
            <input 
              data-testid="search-input" 
              type="text" 
              placeholder="Search posts..." 
              value="\${searchQuery}" 
              oninput="handleSearch(this.value)"
              class="w-full border p-2 rounded focus:ring focus:ring-indigo-200"
            />
            
            <!-- Virtualized Scroll Container -->
            <div id="posts-scroll-container" class="space-y-4 max-h-[600px] overflow-y-auto pr-2" onscroll="handleScroll(this)">
              \${postsListHtml}
            </div>
          </div>
        </main>

        <!-- New Post Modal -->
        \${showingNewPostModal ? renderNewPostModal() : ''}
      \`;
    }

    function renderPostDetail() {
      const db = getDB();
      const post = db.posts.find(p => p.id === activePostId);
      if (!post) {
        return \`
          \${renderSSEHeader()}
          <div class="py-12 text-center">
            <h2 class="text-xl font-bold text-red-500">Post Not Found</h2>
            <a onclick="navigate('/')" class="text-indigo-500 cursor-pointer underline">Back to Homepage</a>
          </div>
        \`;
      }

      // If SSE is disconnected, show a reconnection spinner
      let sseWarning = "";
      if (sseStatus === "disconnected") {
        sseWarning = \`
          <div class="bg-yellow-100 border-l-4 border-yellow-500 text-yellow-700 p-4 mb-4 flex items-center justify-between" role="alert">
            <div class="flex items-center gap-2">
              <svg class="animate-spin h-5 w-5 text-yellow-700" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              <span>Disconnected. Reconnecting to updates...</span>
            </div>
            <button onclick="setSSEStatus('connected')" class="text-xs bg-yellow-600 text-white px-2 py-1 rounded">Force Reconnect</button>
          </div>
        \`;
      }

      const comments = db.comments.filter(c => c.postId === activePostId);

      // Recursive comment tree builder
      function buildCommentTree(parentId, depth = 0) {
        const children = comments.filter(c => c.parentId === parentId);
        if (children.length === 0) return '';
        
        // Indentation limit check (Feature 5, T2.2: Deep nesting limit max 5)
        const currentIndent = Math.min(depth, 5) * 16; 
        
        return children.map(comment => \`
          <div 
            data-testid="comment-item" 
            class="border-l-2 border-gray-200 pl-4 py-2 mt-2" 
            style="margin-left: \${parentId === null ? 0 : 8}px"
          >
            <div class="flex items-center gap-2 text-xs mb-1">
              <img src="\${comment.author.avatar}" class="w-4 h-4 rounded-full" />
              <span data-testid="comment-author" class="font-bold">\${comment.author.username}</span>
              \${comment.author.isAi ? '<span class="bg-indigo-100 text-indigo-800 text-[10px] px-1 rounded">AI</span>' : ''}
              <span class="text-gray-400 text-[10px]">\${comment.createdAt}</span>
            </div>
            <p data-testid="comment-content" class="text-sm text-gray-800">\${sanitizeHTML(comment.content)}</p>
            <div class="mt-1">
              <button 
                data-testid="comment-reply-btn-\${comment.id}" 
                onclick="showReplyInput(\${comment.id})" 
                class="text-xs text-indigo-500 hover:underline"
              >
                Reply
              </button>
            </div>
            <div id="reply-input-container-\${comment.id}" class="hidden mt-2">
              <textarea 
                id="reply-text-\${comment.id}" 
                oninput="handleTextareaInput(this, \${comment.id})"
                placeholder="Write a reply..." 
                class="w-full border p-1 text-sm rounded focus:outline-none focus:ring focus:ring-indigo-100"
              ></textarea>
              <div class="relative">
                <div id="mention-dropdown-\${comment.id}" data-testid="mention-dropdown" class="hidden absolute bg-white border shadow-md rounded mt-1 z-10 w-48 text-sm">
                  <!-- Mentions -->
                </div>
              </div>
              <button 
                data-testid="comment-submit-btn" 
                onclick="submitReply(\${comment.id})" 
                class="mt-1 bg-indigo-600 text-white text-xs px-3 py-1 rounded hover:bg-indigo-700"
              >
                Submit
              </button>
            </div>
            \${buildCommentTree(comment.id, depth + 1)}
          </div>
        \`).join("");
      }

      const commentsHtml = buildCommentTree(null, 0);

      // Get all active agent avatars that replied
      const aiAvatarsHtml = post.aiAvatars.map(av => {
        const agentName = av.includes("seed=") ? av.split("seed=")[1] : "Agent";
        return \`<img data-testid="ai-avatar-\${agentName}" src="\${av}" class="w-6 h-6 rounded-full border border-white" />\`;
      }).join("");

      return \`
        \${renderSSEHeader()}
        <header class="bg-white border-b py-4 px-6 flex justify-between items-center shadow-sm">
          <a data-testid="nav-home-link" onclick="navigate('/')" class="text-xl font-bold cursor-pointer text-indigo-600">AI Monolith Forum</a>
          <a data-testid="nav-ai-plaza-link" onclick="navigate('/ai-plaza')" class="text-sm cursor-pointer hover:text-indigo-600">AI Plaza</a>
        </header>

        <main class="max-w-3xl mx-auto py-8 px-4">
          \${sseWarning}
          
          <article class="bg-white p-8 rounded-lg shadow-sm border border-gray-100 mb-8">
            <div class="flex justify-between items-start mb-4">
              <span class="bg-indigo-100 text-indigo-800 text-xs px-2.5 py-0.5 rounded-full font-semibold uppercase">\${post.category}</span>
              <span data-testid="post-detail-status" class="text-xs text-gray-500">Status: <strong>\${post.aiStatus}</strong></span>
            </div>
            <h1 data-testid="post-detail-title" class="text-2xl font-extrabold mb-4">\${post.title}</h1>
            
            <div data-testid="post-detail-content" class="prose max-w-none text-gray-800 mb-6">
              \${sanitizeHTML(post.content)}
            </div>

            <div class="flex justify-between items-center pt-4 border-t">
              <div class="flex items-center gap-2">
                <img src="\${post.author.avatar}" class="w-8 h-8 rounded-full" />
                <span class="text-sm font-semibold">\${post.author.username}</span>
              </div>
              <div class="flex items-center gap-4">
                <button data-testid="like-btn" onclick="likePost(\${post.id})" class="flex items-center gap-1 text-gray-500 hover:text-red-500">
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"></path></svg>
                  <span data-testid="like-count">\${post.likes || 0}</span>
                </button>
                <div class="flex items-center gap-1">
                  \${aiAvatarsHtml}
                </div>
              </div>
            </div>
          </article>

          <!-- Comments Section -->
          <div class="bg-white p-8 rounded-lg shadow-sm border border-gray-100">
            <h3 class="text-lg font-bold mb-4">Comments (\${comments.length})</h3>
            
            <!-- Root Comment Input -->
            <div class="mb-6">
              <textarea 
                id="comment-input-root" 
                data-testid="comment-input"
                oninput="handleTextareaInput(this, 'root')"
                placeholder="Write a comment..." 
                class="w-full border p-2 rounded focus:outline-none focus:ring focus:ring-indigo-100"
              ></textarea>
              <div class="relative">
                <div id="mention-dropdown-root" data-testid="mention-dropdown" class="hidden absolute bg-white border shadow-md rounded mt-1 z-10 w-48 text-sm">
                  <!-- Mentions -->
                </div>
              </div>
              <button 
                data-testid="comment-submit-btn" 
                onclick="submitRootComment()" 
                class="mt-2 bg-indigo-600 text-white px-4 py-2 rounded hover:bg-indigo-700"
              >
                Submit Comment
              </button>
            </div>

            <!-- Comment Tree -->
            <div class="space-y-4">
              \${commentsHtml}
            </div>
          </div>
        </main>
      \`;
    }

    function renderAIPlaza() {
      const db = getDB();
      const activeAgents = db.agents;

      const agentsHtml = activeAgents.map(agent => \`
        <div class="bg-white p-6 rounded-lg shadow border border-gray-200 flex gap-4">
          <img src="\${agent.avatar}" class="w-16 h-16 rounded-full border" />
          <div class="flex-1">
            <div class="flex justify-between items-start">
              <h3 class="text-lg font-bold">\${agent.name}</h3>
              <span class="text-xs px-2 py-0.5 rounded \${agent.active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}">
                \${agent.active ? 'Active' : 'Inactive'}
              </span>
            </div>
            <p class="text-sm text-gray-600 mt-1">\${agent.description}</p>
            <div class="mt-3 grid grid-cols-2 gap-2 text-xs text-gray-500">
              <div><strong>Threshold:</strong> \${agent.replyThreshold}</div>
              <div><strong>Activity Level:</strong> \${agent.activityLevel}</div>
            </div>
          </div>
        </div>
      \`).join("");

      return \`
        \${renderSSEHeader()}
        <header class="bg-white border-b py-4 px-6 flex justify-between items-center shadow-sm">
          <a data-testid="nav-home-link" onclick="navigate('/')" class="text-xl font-bold cursor-pointer text-indigo-600">AI Monolith Forum</a>
          <div class="flex gap-4">
            <a data-testid="nav-ai-plaza-link" onclick="navigate('/ai-plaza')" class="text-sm cursor-pointer text-indigo-600 font-bold">AI Plaza</a>
          </div>
        </header>

        <main class="max-w-4xl mx-auto py-8 px-4">
          <h2 class="text-2xl font-bold mb-6">AI Plaza</h2>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            \${agentsHtml}
          </div>
        </main>
      \`;
    }

    function renderNewPostModal() {
      return \`
        <div class="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center z-50">
          <div class="bg-white p-6 rounded-lg shadow-xl max-w-lg w-full">
            <h2 class="text-xl font-bold mb-4">Create New Topic</h2>
            <form onsubmit="submitNewPost(event)" class="space-y-4">
              <div>
                <label class="block text-sm font-semibold text-gray-700">Title</label>
                <input id="new-post-title" type="text" required class="w-full border p-2 rounded focus:outline-none focus:ring focus:ring-indigo-100" />
                <p id="title-error" class="hidden text-xs text-red-500 mt-1">Title is required</p>
              </div>
              <div>
                <label class="block text-sm font-semibold text-gray-700">Category</label>
                <select id="new-post-category" class="w-full border p-2 rounded focus:outline-none focus:ring focus:ring-indigo-100 bg-white">
                  <option value="后端开发">后端开发</option>
                  <option value="前端开发">前端开发</option>
                  <option value="人工智能">人工智能</option>
                </select>
              </div>
              <div>
                <label class="block text-sm font-semibold text-gray-700">Content</label>
                <textarea id="new-post-content" required rows="4" class="w-full border p-2 rounded focus:outline-none focus:ring focus:ring-indigo-100"></textarea>
                <p id="content-error" class="hidden text-xs text-red-500 mt-1">Content is required</p>
              </div>
              <div class="flex justify-end gap-2">
                <button type="button" onclick="closeNewPostModal()" class="px-4 py-2 border rounded hover:bg-gray-100 text-sm">Cancel</button>
                <button type="submit" id="submit-post-btn" class="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700 text-sm">Publish Post</button>
              </div>
            </form>
          </div>
        </div>
      \`;
    }

    // Interaction handlers
    window.toggleCategory = (cat) => {
      activeCategory = (activeCategory === cat) ? null : cat;
      render();
    };

    window.toggleTag = (tag) => {
      if (activeTags.includes(tag)) {
        activeTags = activeTags.filter(t => t !== tag);
      } else {
        activeTags.push(tag);
      }
      render();
    };

    window.clearFilters = () => {
      activeCategory = null;
      activeTags = [];
      searchQuery = "";
      render();
    };

    window.handleSearch = (val) => {
      searchQuery = val;
      render();
    };

    window.handleScroll = (el) => {
      if (el.scrollTop + el.clientHeight >= el.scrollHeight - 20) {
        visiblePostsCount += 10;
        // Re-render only list part to simulate infinite scroll
        const db = getDB();
        let posts = db.posts;
        if (activeCategory) posts = posts.filter(p => p.category === activeCategory);
        if (activeTags.length > 0) posts = posts.filter(p => activeTags.every(t => p.tags.includes(t)));
        if (searchQuery) {
          const sq = searchQuery.toLowerCase();
          posts = posts.filter(p => p.title.toLowerCase().includes(sq) || p.content.toLowerCase().includes(sq));
        }
        const displayedPosts = posts.slice(0, visiblePostsCount);
        const postsListHtml = displayedPosts.map(post => \`
          <div data-testid="post-card-\${post.id}" onclick="navigate('/post/\${post.id}')" class="bg-white p-6 rounded-lg shadow-sm border border-gray-100 cursor-pointer hover:shadow-md transition">
            <div data-testid="post-card">
              <div class="flex justify-between items-start mb-2">
                <span class="text-xs font-semibold text-indigo-600 uppercase tracking-wider">\${post.category}</span>
                <span class="text-xs text-gray-400">Status: \${post.aiStatus}</span>
              </div>
              <h3 data-testid="post-title" class="text-lg font-bold hover:text-indigo-600 mb-2">\${post.title}</h3>
              <p class="text-gray-600 line-clamp-2 text-sm mb-4">\${post.content}</p>
              <div class="flex justify-between items-center text-xs text-gray-500">
                <div class="flex items-center gap-1"><img src="\${post.author.avatar}" class="w-5 h-5 rounded-full" /><span>\${post.author.username}</span></div>
                <div class="flex items-center gap-2"><span class="font-bold">\${post.aiResponsesCount} replies</span></div>
              </div>
            </div>
          </div>
        \`).join("");
        el.innerHTML = postsListHtml;
      }
    };

    window.openNewPostModal = () => {
      showingNewPostModal = true;
      render();
    };

    window.closeNewPostModal = () => {
      showingNewPostModal = false;
      render();
    };

    window.submitNewPost = (e) => {
      e.preventDefault();
      const title = document.getElementById("new-post-title").value.trim();
      const category = document.getElementById("new-post-category").value;
      const content = document.getElementById("new-post-content").value.trim();

      if (!title) {
        document.getElementById("title-error").classList.remove("hidden");
        return;
      }
      if (!content) {
        document.getElementById("content-error").classList.remove("hidden");
        return;
      }

      // Disable button
      const btn = document.getElementById("submit-post-btn");
      if (btn) btn.disabled = true;

      const db = getDB();
      const newPost = {
        id: db.posts.length + 1,
        title,
        content,
        category,
        tags: ["Testing"],
        author: {
          username: "alex_dev",
          avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=alex"
        },
        aiStatus: "PENDING",
        aiResponsesCount: 0,
        aiAvatars: [],
        createdAt: new Date().toISOString(),
        likes: 0
      };

      db.posts.unshift(newPost);
      saveDB(db);

      setTimeout(() => {
        showingNewPostModal = false;
        navigate('/');
        // Trigger simulation
        runSimulation(newPost.id, null);
      }, 100);
    };

    window.likePost = (postId) => {
      const db = getDB();
      const post = db.posts.find(p => p.id === postId);
      if (post) {
        post.likes = (post.likes || 0) + 1;
        saveDB(db);
        render();
      }
    };

    window.showReplyInput = (commentId) => {
      const container = document.getElementById("reply-input-container-" + commentId);
      if (container) {
        container.classList.toggle("hidden");
      }
    };

    window.handleTextareaInput = (el, id) => {
      const text = el.value;
      const cursor = el.selectionStart;
      const prefix = text.slice(0, cursor);
      const atIndex = prefix.lastIndexOf("@");
      const dropdown = document.getElementById("mention-dropdown-" + id);
      
      if (!dropdown) return;

      if (atIndex !== -1 && atIndex === prefix.length - 1) {
        const db = getDB();
        const activeAgents = db.agents.filter(a => a.active);
        
        dropdown.classList.remove("hidden");
        dropdown.innerHTML = activeAgents.map(agent => \`
          <div 
            data-testid="mention-item-\${agent.name}"
            onclick="insertMention('\${agent.name}', \${atIndex}, '\${id}')"
            class="px-3 py-2 hover:bg-indigo-50 cursor-pointer flex items-center gap-2"
          >
            <img src="\${agent.avatar}" class="w-4 h-4 rounded-full" />
            <span>\${agent.name}</span>
          </div>
        \`).join("");
      } else {
        dropdown.classList.add("hidden");
      }
    };

    window.insertMention = (name, atIndex, id) => {
      const el = document.getElementById(id === 'root' ? 'comment-input-root' : "reply-text-" + id);
      if (!el) return;
      const text = el.value;
      el.value = text.slice(0, atIndex) + "@" + name + " " + text.slice(atIndex + 1);
      const dropdown = document.getElementById("mention-dropdown-" + id);
      if (dropdown) dropdown.classList.add("hidden");
      el.focus();
    };

    window.submitRootComment = () => {
      const input = document.getElementById("comment-input-root");
      const text = input.value.trim();
      if (!text) return;

      const db = getDB();
      const isDuplicate = db.comments.some(c => c.postId === activePostId && c.content === text);
      if (isDuplicate) return;

      const newComment = {
        id: db.comments.length + 1,
        postId: activePostId,
        parentId: null,
        content: text,
        author: {
          username: "alex_dev",
          avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=alex",
          isAi: false
        },
        createdAt: new Date().toISOString()
      };

      db.comments.push(newComment);
      saveDB(db);
      input.value = "";
      render();

      // Trigger simulation for mentions/followup
      runSimulation(activePostId, newComment.id);
    };

    window.submitReply = (parentId) => {
      const input = document.getElementById("reply-text-" + parentId);
      const text = input.value.trim();
      if (!text) return;

      const db = getDB();
      const isDuplicate = db.comments.some(c => c.postId === activePostId && c.content === text);
      if (isDuplicate) return;

      const newComment = {
        id: db.comments.length + 1,
        postId: activePostId,
        parentId: parentId,
        content: text,
        author: {
          username: "alex_dev",
          avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=alex",
          isAi: false
        },
        createdAt: new Date().toISOString()
      };

      db.comments.push(newComment);
      saveDB(db);
      input.value = "";
      
      const container = document.getElementById("reply-input-container-" + parentId);
      if (container) container.classList.add("hidden");
      render();

      runSimulation(activePostId, newComment.id);
    };

    // Global listener for sync in same browser E2E session
    window.addEventListener("storage_updated", () => {
      if (localStorage.getItem("sse_status") === "disconnected") return;
      render();
    });

    function render() {
      const app = document.getElementById("app");
      if (currentPath === "/" || currentPath === "/home" || currentPath === "/home/") {
        app.innerHTML = renderHome();
      } else if (currentPath.startsWith("/post/")) {
        app.innerHTML = renderPostDetail();
      } else if (currentPath === "/ai-plaza") {
        app.innerHTML = renderAIPlaza();
      }
    }

    // Initial render
    const initialPathMatch = window.location.pathname.match(/\\/post\\/(\\d+)/);
    if (initialPathMatch) {
      activePostId = parseInt(initialPathMatch[1]);
    }
    render();
  </script>
</body>
</html>`;
  } else {
    // Admin layout template
    return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Mock Admin Console</title>
  <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100 text-gray-800 min-h-screen">
  <div id="app"></div>

  <script>
    const INITIAL_DB = ${JSON.stringify(INITIAL_DB_STATE)};
    
    function getDB() {
      const val = localStorage.getItem("ai_forum_db_state");
      if (!val) {
        localStorage.setItem("ai_forum_db_state", JSON.stringify(INITIAL_DB));
        return INITIAL_DB;
      }
      try {
        return JSON.parse(val);
      } catch (e) {
        return INITIAL_DB;
      }
    }

    function saveDB(db) {
      localStorage.setItem("ai_forum_db_state", JSON.stringify(db));
      window.dispatchEvent(new Event("storage_updated"));
    }

    let activeTab = "dashboard"; // dashboard, agents, tasks, logs
    let selectedAgentId = null;
    let selectedTaskId = null;
    let logSearchQuery = "";
    
    // Services mock status
    let services = {
      "api-server": "active",
      "worker-service": "active",
      "outbox-publisher": "active"
    };

    function navigateTab(tab) {
      activeTab = tab;
      render();
    }

    function toggleAgentActive(agentId) {
      const db = getDB();
      const agent = db.agents.find(a => a.id === agentId);
      if (agent) {
        agent.active = !agent.active;
        saveDB(db);
        render();
      }
    }

    function openEditAgent(agentId) {
      selectedAgentId = agentId;
      render();
    }

    function closeEditAgent() {
      selectedAgentId = null;
      render();
    }

    function saveAgentConfig(e) {
      e.preventDefault();
      const db = getDB();
      const agent = db.agents.find(a => a.id === selectedAgentId);
      if (agent) {
        agent.replyThreshold = parseFloat(document.getElementById("edit-agent-threshold").value);
        agent.activityLevel = parseFloat(document.getElementById("edit-agent-level").value);
        agent.systemPrompt = document.getElementById("edit-agent-prompt").value;
        saveDB(db);
        selectedAgentId = null;
        render();
      }
    }

    function openTaskDetail(taskId) {
      selectedTaskId = taskId;
      render();
    }

    function closeTaskDetail() {
      selectedTaskId = null;
      render();
    }

    function retryTask(taskId) {
      const db = getDB();
      const task = db.tasks.find(t => t.id === taskId);
      if (task) {
        task.status = "PENDING";
        task.retryCount++;
        saveDB(db);
        render();

        // Simulate background retrying execution
        setTimeout(() => {
          const rdb = getDB();
          const rt = rdb.tasks.find(t => t.id === taskId);
          if (rt) {
            rt.status = "PROCESSING";
            rt.startedAt = new Date().toISOString();
          }
          saveDB(rdb);
          render();

          setTimeout(() => {
            const fdb = getDB();
            const ft = fdb.tasks.find(t => t.id === taskId);
            const post = fdb.posts.find(p => p.id === ft.postId);
            const agent = fdb.agents.find(a => a.id === ft.aiAgentId);
            if (ft && post && agent) {
              const replyText = "Retried reply: Decoupled design passes integration tests.";
              const newComment = {
                id: fdb.comments.length + 1,
                postId: ft.postId,
                parentId: ft.parentCommentId,
                content: replyText,
                author: {
                  username: agent.name,
                  avatar: agent.avatar,
                  isAi: true,
                  aiAgentId: agent.id
                },
                createdAt: new Date().toISOString()
              };

              fdb.comments.push(newComment);
              ft.status = "COMPLETED";
              ft.result = replyText;
              ft.finishedAt = new Date().toISOString();
              saveDB(fdb);
              render();
            }
          }, 1000);
        }, 500);
      }
    }

    // Storage updates listener
    window.addEventListener("storage_updated", () => {
      render();
    });

    function renderDashboard() {
      const db = getDB();
      return \`
        <div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
          <div data-testid="metric-posts" class="bg-white p-6 rounded shadow-sm">
            <h4 class="text-gray-400 text-xs font-bold uppercase mb-1">Total Posts</h4>
            <span class="text-2xl font-bold text-gray-900">\${db.posts.length}</span>
          </div>
          <div data-testid="metric-comments" class="bg-white p-6 rounded shadow-sm">
            <h4 class="text-gray-400 text-xs font-bold uppercase mb-1">Total Comments</h4>
            <span class="text-2xl font-bold text-gray-900">\${db.comments.length}</span>
          </div>
          <div data-testid="metric-tasks" class="bg-white p-6 rounded shadow-sm">
            <h4 class="text-gray-400 text-xs font-bold uppercase mb-1">Total Tasks</h4>
            <span class="text-2xl font-bold text-gray-900">\${db.tasks.length}</span>
          </div>
          <div data-testid="metric-agents" class="bg-white p-6 rounded shadow-sm">
            <h4 class="text-gray-400 text-xs font-bold uppercase mb-1">Active Agents</h4>
            <span class="text-2xl font-bold text-gray-900">\${db.agents.filter(a => a.active).length}</span>
          </div>
        </div>

        <div class="bg-white p-6 rounded shadow-sm mb-8">
          <h3 class="text-lg font-bold mb-4">Service Status Indicators</h3>
          <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div data-testid="service-api-server" class="p-4 border rounded flex justify-between items-center">
              <span>api-server</span>
              <span class="px-2 py-0.5 text-xs bg-green-100 text-green-800 rounded">ACTIVE</span>
            </div>
            <div data-testid="service-worker-service" class="p-4 border rounded flex justify-between items-center">
              <span>worker-service</span>
              <span class="px-2 py-0.5 text-xs bg-green-100 text-green-800 rounded">ACTIVE</span>
            </div>
            <div data-testid="service-outbox-publisher" class="p-4 border rounded flex justify-between items-center">
              <span>outbox-publisher</span>
              <span class="px-2 py-0.5 text-xs bg-green-100 text-green-800 rounded">ACTIVE</span>
            </div>
          </div>
        </div>
      \`;
    }

    function renderAgents() {
      const db = getDB();
      const agentsTableHtml = db.agents.map(agent => \`
        <tr data-testid="agent-row-\${agent.id}" class="border-b">
          <td class="px-6 py-4 flex items-center gap-2">
            <img src="\${agent.avatar}" class="w-8 h-8 rounded-full" />
            <span class="font-bold">\${agent.name}</span>
          </td>
          <td class="px-6 py-4 text-sm">\${agent.replyThreshold}</td>
          <td class="px-6 py-4 text-sm">\${agent.activityLevel}</td>
          <td class="px-6 py-4 text-sm">
            <button 
              data-testid="agent-toggle-active-\${agent.id}" 
              onclick="toggleAgentActive(\${agent.id})" 
              class="px-2 py-1 rounded text-xs \${agent.active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}"
            >
              \${agent.active ? 'Active' : 'Inactive'}
            </button>
          </td>
          <td class="px-6 py-4 text-sm">
            <button data-testid="agent-edit-btn-\${agent.id}" onclick="openEditAgent(\${agent.id})" class="text-indigo-600 hover:underline">Edit</button>
          </td>
        </tr>
      \`).join("");

      return \`
        <div class="bg-white rounded shadow-sm overflow-hidden">
          <table class="w-full text-left border-collapse">
            <thead>
              <tr class="bg-gray-50 border-b">
                <th class="px-6 py-3 text-xs font-bold text-gray-500 uppercase">Agent</th>
                <th class="px-6 py-3 text-xs font-bold text-gray-500 uppercase">Threshold</th>
                <th class="px-6 py-3 text-xs font-bold text-gray-500 uppercase">Activity</th>
                <th class="px-6 py-3 text-xs font-bold text-gray-500 uppercase">Status</th>
                <th class="px-6 py-3 text-xs font-bold text-gray-500 uppercase">Actions</th>
              </tr>
            </thead>
            <tbody>
              \${agentsTableHtml}
            </tbody>
          </table>
        </div>

        \${selectedAgentId ? renderAgentDrawer() : ''}
      \`;
    }

    function renderAgentDrawer() {
      const db = getDB();
      const agent = db.agents.find(a => a.id === selectedAgentId);
      if (!agent) return '';

      return \`
        <div class="fixed inset-y-0 right-0 w-96 bg-white shadow-xl z-50 p-6 flex flex-col border-l">
          <div class="flex justify-between items-center mb-6">
            <h3 class="text-lg font-bold">Edit Agent: \${agent.name}</h3>
            <button onclick="closeEditAgent()" class="text-gray-500 hover:text-gray-700">Close</button>
          </div>
          <form onsubmit="saveAgentConfig(event)" class="space-y-4 flex-1">
            <div>
              <label class="block text-xs font-bold text-gray-500 uppercase mb-1">Reply Threshold</label>
              <input id="edit-agent-threshold" data-testid="drawer-agent-threshold" type="number" step="any" min="0" max="1" value="\${agent.replyThreshold}" required class="w-full border p-2 rounded" />
            </div>
            <div>
              <label class="block text-xs font-bold text-gray-500 uppercase mb-1">Activity Level</label>
              <input id="edit-agent-level" data-testid="drawer-agent-active-level" type="number" step="any" min="0" max="1" value="\${agent.activityLevel}" required class="w-full border p-2 rounded" />
            </div>
            <div>
              <label class="block text-xs font-bold text-gray-500 uppercase mb-1">System Prompt</label>
              <textarea id="edit-agent-prompt" data-testid="drawer-agent-system-prompt" rows="6" required class="w-full border p-2 rounded text-sm">\${agent.systemPrompt}</textarea>
            </div>
            <button type="submit" data-testid="drawer-agent-save-btn" class="w-full bg-indigo-600 text-white py-2 rounded hover:bg-indigo-700 font-bold">Save Configuration</button>
          </form>
        </div>
      \`;
    }

    function renderTasks() {
      const db = getDB();
      const tasksTableHtml = db.tasks.length === 0
        ? \`<tr><td colspan="5" class="text-center py-6 text-gray-500">No tasks executed yet.</td></tr>\`
        : db.tasks.map(task => {
          const agent = db.agents.find(a => a.id === task.aiAgentId);
          return \`
            <tr data-testid="task-row-\${task.id}" class="border-b text-sm">
              <td class="px-6 py-4">Task #\${task.id}</td>
              <td class="px-6 py-4 font-semibold">\${agent ? agent.name : 'Unknown'}</td>
              <td class="px-6 py-4">\${task.triggerType}</td>
              <td class="px-6 py-4">
                <span class="px-2 py-0.5 rounded text-xs font-semibold 
                  \${task.status === 'COMPLETED' ? 'bg-green-100 text-green-800' : ''}
                  \${task.status === 'PROCESSING' ? 'bg-yellow-100 text-yellow-800' : ''}
                  \${task.status === 'PENDING' ? 'bg-blue-100 text-blue-800' : ''}
                  \${task.status === 'FAILED' ? 'bg-red-100 text-red-800' : ''}
                ">
                  \${task.status}
                </span>
              </td>
              <td class="px-6 py-4 space-x-2">
                <button data-testid="task-detail-btn-\${task.id}" onclick="openTaskDetail(\${task.id})" class="text-indigo-600 hover:underline">View Payload</button>
                \${task.status === 'FAILED' ? \`<button data-testid="task-retry-btn-\${task.id}" onclick="retryTask(\${task.id})" class="text-green-600 hover:underline">Retry</button>\` : ''}
              </td>
            </tr>
          \`;
        }).join("");

      return \`
        <div class="bg-white rounded shadow-sm overflow-hidden">
          <table class="w-full text-left border-collapse">
            <thead>
              <tr class="bg-gray-50 border-b">
                <th class="px-6 py-3 text-xs font-bold text-gray-500 uppercase">ID</th>
                <th class="px-6 py-3 text-xs font-bold text-gray-500 uppercase">Agent</th>
                <th class="px-6 py-3 text-xs font-bold text-gray-500 uppercase">Trigger</th>
                <th class="px-6 py-3 text-xs font-bold text-gray-500 uppercase">Status</th>
                <th class="px-6 py-3 text-xs font-bold text-gray-500 uppercase">Actions</th>
              </tr>
            </thead>
            <tbody>
              \${tasksTableHtml}
            </tbody>
          </table>
        </div>

        \${selectedTaskId ? renderTaskDrawer() : ''}
      \`;
    }

    function renderTaskDrawer() {
      const db = getDB();
      const task = db.tasks.find(t => t.id === selectedTaskId);
      if (!task) return '';

      return \`
        <div class="fixed inset-y-0 right-0 w-96 bg-white shadow-xl z-50 p-6 flex flex-col border-l">
          <div class="flex justify-between items-center mb-6">
            <h3 class="text-lg font-bold">Task Payload Detail</h3>
            <button onclick="closeTaskDetail()" class="text-gray-500 hover:text-gray-700">Close</button>
          </div>
          <div class="space-y-4 flex-1 overflow-y-auto">
            <div>
              <h4 class="text-xs font-bold text-gray-400 uppercase mb-1">Prompt Context</h4>
              <pre data-testid="drawer-task-payload" class="bg-gray-50 p-3 rounded text-xs overflow-x-auto whitespace-pre-wrap">\${task.prompt}</pre>
            </div>
            <div>
              <h4 class="text-xs font-bold text-gray-400 uppercase mb-1">Result Summary</h4>
              <p class="bg-gray-50 p-3 rounded text-xs text-gray-800">\${task.result || 'No result yet'}</p>
            </div>
            <div>
              <h4 class="text-xs font-bold text-gray-400 uppercase mb-1">Retry Stats</h4>
              <p class="text-xs">Retry count: \${task.retryCount}</p>
            </div>
          </div>
        </div>
      \`;
    }

    function renderLogs() {
      const db = getDB();
      
      let filteredLogs = db.decisionLogs;
      if (logSearchQuery) {
        const queryVal = parseInt(logSearchQuery);
        if (!isNaN(queryVal)) {
          filteredLogs = filteredLogs.filter(log => log.postId === queryVal);
        }
      }

      const logsHtml = filteredLogs.length === 0
        ? \`<tr><td colspan="6" class="text-center py-6 text-gray-500">No decision logs matches criteria.</td></tr>\`
        : filteredLogs.map(log => \`
          <tr data-testid="log-row-\${log.id}" class="border-b text-xs">
            <td class="px-6 py-4 font-semibold">\${log.aiAgentName}</td>
            <td class="px-6 py-4">Post #\${log.postId}</td>
            <td class="px-6 py-4">\${log.triggerType}</td>
            <td class="px-6 py-4">\${log.willingnessScore.toFixed(2)}</td>
            <td class="px-6 py-4">\${log.thresholdValue}</td>
            <td class="px-6 py-4">
              <span class="px-2 py-0.5 rounded text-[10px] font-bold 
                \${log.decision === 'REPLY' ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'}
              ">
                \${log.decision}
              </span>
            </td>
          </tr>
        \`).join("");

      return \`
        <div class="mb-4">
          <input 
            id="log-search-input"
            data-testid="log-search-post-id"
            type="text" 
            placeholder="Filter by Post ID..." 
            value="\${logSearchQuery}" 
            oninput="handleLogSearch(this.value)"
            class="border p-2 rounded text-sm w-48"
          />
        </div>

        <div class="bg-white rounded shadow-sm overflow-hidden">
          <table class="w-full text-left border-collapse">
            <thead>
              <tr class="bg-gray-50 border-b">
                <th class="px-6 py-3 text-xs font-bold text-gray-500 uppercase">Agent</th>
                <th class="px-6 py-3 text-xs font-bold text-gray-500 uppercase">Post ID</th>
                <th class="px-6 py-3 text-xs font-bold text-gray-500 uppercase">Trigger</th>
                <th class="px-6 py-3 text-xs font-bold text-gray-500 uppercase">Willingness</th>
                <th class="px-6 py-3 text-xs font-bold text-gray-500 uppercase">Threshold</th>
                <th class="px-6 py-3 text-xs font-bold text-gray-500 uppercase">Decision</th>
              </tr>
            </thead>
            <tbody>
              \${logsHtml}
            </tbody>
          </table>
        </div>
      \`;
    }

    window.handleLogSearch = (val) => {
      logSearchQuery = val;
      render();
    };

    function render() {
      const app = document.getElementById("app");
      let activeContent = "";
      if (activeTab === "dashboard") activeContent = renderDashboard();
      else if (activeTab === "agents") activeContent = renderAgents();
      else if (activeTab === "tasks") activeContent = renderTasks();
      else if (activeTab === "logs") activeContent = renderLogs();

      app.innerHTML = \`
        <div class="flex min-h-screen">
          <!-- Sidebar -->
          <div class="w-64 bg-gray-800 text-white p-6 space-y-6">
            <h2 class="text-xl font-bold tracking-tight">Operator Dashboard</h2>
            <nav class="flex flex-col gap-2">
              <button onclick="navigateTab('dashboard')" class="text-left py-2 px-3 rounded hover:bg-gray-700 \${activeTab === 'dashboard' ? 'bg-gray-700 font-bold' : ''}">Dashboard</button>
              <button onclick="navigateTab('agents')" class="text-left py-2 px-3 rounded hover:bg-gray-700 \${activeTab === 'agents' ? 'bg-gray-700 font-bold' : ''}">AI Agent Config</button>
              <button onclick="navigateTab('tasks')" class="text-left py-2 px-3 rounded hover:bg-gray-700 \${activeTab === 'tasks' ? 'bg-gray-700 font-bold' : ''}">AI Task Queue</button>
              <button onclick="navigateTab('logs')" class="text-left py-2 px-3 rounded hover:bg-gray-700 \${activeTab === 'logs' ? 'bg-gray-700 font-bold' : ''}">Decision Logs</button>
            </nav>
          </div>

          <!-- Main Panel -->
          <div class="flex-1 p-8">
            <h1 class="text-2xl font-bold mb-6 capitalize">\${activeTab === 'dashboard' ? 'System Overview' : activeTab}</h1>
            \${activeContent}
          </div>
        </div>
      \`;
    }

    render();
  </script>
</body>
</html>`;
  }
}
