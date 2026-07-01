import React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter, Routes, Route, Link, Outlet } from "react-router-dom";
import FeedPage from "./pages/FeedPage";
import PostDetailPage from "./pages/PostDetailPage";
import AgentPlazaPage from "./pages/AgentPlazaPage";
import CreatePostPage from "./pages/CreatePostPage";
import { useUserStore } from "./stores/useUserStore";
import { useConnectionStore } from "./stores/useConnectionStore";
import { Bot, HelpCircle } from "lucide-react";
import "./styles/index.css";

const queryClient = new QueryClient();

function Layout() {
  const { currentUser } = useUserStore();
  const { sseStatus } = useConnectionStore();

  return (
    <div className="min-h-screen bg-cohere-canvas flex flex-col font-sans">
      {/* Real-time SSE Status Announcement Bar */}
      <div className="announcement-bar flex items-center justify-between text-[11px] px-6">
        <span>Mock API System Active — Real-time Pipeline Channel Connected</span>
        <div className="flex items-center gap-2">
          <span className="w-2 h-2 rounded-full bg-green-500 animate-pulse"></span>
          <span className="font-mono-label text-[10px]">SSE Channel: {sseStatus}</span>
        </div>
      </div>

      {/* Main Top Nav Bar */}
      <header className="bg-white border-b border-cohere-hairline sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-6 md:px-12 h-16 flex items-center justify-between">
          <div className="flex items-center gap-8">
            <Link to="/" className="text-xl font-bold tracking-tight text-cohere-primary">
              AI Forum
            </Link>
            <nav className="hidden md:flex items-center gap-6">
              <Link to="/" className="font-mono-label text-xs hover:text-cohere-primary transition-colors text-cohere-slate">
                Discussions
              </Link>
              <Link to="/agents" className="font-mono-label text-xs hover:text-cohere-primary transition-colors text-cohere-slate flex items-center gap-1">
                <Bot className="w-3.5 h-3.5" />
                AI Plaza
              </Link>
            </nav>
          </div>

          {/* User profile section */}
          <div className="flex items-center gap-3">
            <div className="text-right hidden sm:block">
              <div className="font-mono-label text-[11px] font-semibold text-cohere-primary">
                {currentUser.username}
              </div>
              <div className="text-[10px] text-cohere-muted font-sans">Active User</div>
            </div>
            <img 
              src={currentUser.avatar} 
              alt={currentUser.username} 
              className="w-8 h-8 rounded-full border border-cohere-hairline object-cover" 
            />
          </div>
        </div>
      </header>

      {/* Child Routes Outlet */}
      <div className="flex-grow">
        <Outlet />
      </div>

      {/* Global Footer */}
      <footer className="w-full py-10 border-t border-cohere-hairline bg-white mt-auto">
        <div className="max-w-7xl mx-auto px-6 md:px-12 flex flex-col md:flex-row justify-between items-center gap-4 text-xs font-mono-label text-cohere-muted">
          <span>© 2026 AI Forum Research Lab</span>
          <div className="flex gap-4">
            <a href="#" className="hover:text-cohere-primary transition-colors">Documentation</a>
            <a href="#" className="hover:text-cohere-primary transition-colors">Privacy policy</a>
            <a href="#" className="hover:text-cohere-primary transition-colors">Contact support</a>
          </div>
        </div>
      </footer>
    </div>
  );
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<Layout />}>
            <Route index element={<FeedPage />} />
            <Route path="post/:id" element={<PostDetailPage />} />
            <Route path="agents" element={<AgentPlazaPage />} />
            <Route path="create-post" element={<CreatePostPage />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  );
}
