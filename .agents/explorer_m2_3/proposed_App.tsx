import React from "react";
import { BrowserRouter, Routes, Route, Link, useLocation } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useUserStore } from "./stores/useUserStore";
import { useConnectionStore } from "./stores/useConnectionStore";
import { FeedPage } from "./pages/FeedPage";
import { PostDetailPage } from "./pages/PostDetailPage";
import { AgentPlazaPage } from "./pages/AgentPlazaPage";
import { CreatePostPage } from "./pages/CreatePostPage";
import "./styles/index.css";

const queryClient = new QueryClient();

function HeaderAndLayout() {
  const { currentUser } = useUserStore();
  const { sseStatus } = useConnectionStore();
  const location = useLocation();

  const isLinkActive = (path: string) => {
    if (path === "/") {
      return location.pathname === "/" || location.pathname === "/home";
    }
    return location.pathname.startsWith(path);
  };

  return (
    <div className="min-h-screen bg-background text-on-surface antialiased flex flex-col">
      {/* Announcement Bar */}
      <div className="announcement-bar">
        <span>
          Mock API System Active — SSE Connection State: <strong>{sseStatus}</strong>
        </span>
      </div>

      {/* Top Navigation Bar */}
      <header className="bg-surface border-b border-hairline flex justify-between items-center w-full px-margin-desktop h-16 sticky top-0 z-50">
        <div className="flex items-center gap-xl">
          <Link to="/" className="text-headline-lg font-headline-lg font-black text-primary hover:opacity-80 transition-opacity">
            AI Forum
          </Link>
          <nav className="hidden md:flex items-center gap-lg">
            <Link
              to="/"
              className={`font-body-main text-body-main pb-1 transition-all ${
                isLinkActive("/")
                  ? "text-primary border-b-2 border-primary font-semibold"
                  : "text-on-surface-variant hover:text-primary"
              }`}
            >
              首页
            </Link>
            <Link
              to="/"
              className={`font-body-main text-body-main pb-1 transition-all ${
                isLinkActive("/posts")
                  ? "text-primary border-b-2 border-primary font-semibold"
                  : "text-on-surface-variant hover:text-primary"
              }`}
            >
              讨论列表
            </Link>
            <Link
              to="/ai-agents"
              className={`font-body-main text-body-main pb-1 transition-all ${
                isLinkActive("/ai-agents")
                  ? "text-primary border-b-2 border-primary font-semibold"
                  : "text-on-surface-variant hover:text-primary"
              }`}
            >
              AI 角色广场
            </Link>
          </nav>
        </div>
        <div className="flex items-center gap-md">
          <button className="text-on-surface-variant hover:text-primary transition-colors flex items-center justify-center p-sm rounded-full hover:bg-surface-container">
            <span className="material-symbols-outlined">notifications</span>
          </button>
          <div className="flex items-center gap-3">
            <img src={currentUser.avatar} alt={currentUser.username} className="w-8 h-8 rounded-full border" />
            <span className="font-label-mono-bold text-label-mono-bold text-primary">{currentUser.username}</span>
          </div>
        </div>
      </header>

      {/* Main Outlet */}
      <Routes>
        <Route path="/" element={<FeedPage />} />
        <Route path="/home" element={<FeedPage />} />
        <Route path="/posts" element={<FeedPage />} />
        <Route path="/posts/:id" element={<PostDetailPage />} />
        <Route path="/ai-agents" element={<AgentPlazaPage />} />
        <Route path="/create-post" element={<CreatePostPage />} />
      </Routes>

      {/* Shared Footer */}
      <footer className="w-full py-section border-t border-hairline bg-surface mt-auto">
        <div className="max-w-7xl mx-auto px-margin-desktop flex flex-col md:flex-row justify-between items-center gap-lg">
          <span className="font-headline-lg text-headline-lg text-primary">AI Forum</span>
          <div className="flex gap-md font-caption text-caption flex-wrap justify-center">
            <a className="text-muted hover:text-primary transition-colors" href="#terms">
              Terms of Service
            </a>
            <a className="text-muted hover:text-primary transition-colors" href="#privacy">
              Privacy Policy
            </a>
            <a className="text-muted hover:text-primary transition-colors" href="#api">
              API Documentation
            </a>
            <a className="text-muted hover:text-primary transition-colors" href="#contact">
              Contact
            </a>
          </div>
          <span className="font-caption text-caption text-muted">© 2024 AI Forum Research Lab</span>
        </div>
      </footer>
    </div>
  );
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <HeaderAndLayout />
      </BrowserRouter>
    </QueryClientProvider>
  );
}
