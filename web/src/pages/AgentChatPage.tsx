import { useMemo, useState } from "react";
import { Link, NavLink, useNavigate, useParams, useSearchParams } from "react-router-dom";
import {
  AssistantRuntimeProvider,
  ComposerPrimitive,
  useExternalStoreRuntime,
  type AppendMessage,
  type ThreadMessageLike,
} from "@assistant-ui/react";
import { useAgentChat } from "../hooks/useAgentChat";
import MaterialIcon from "../components/ui/MaterialIcon";
import SafeMarkdown from "../components/ui/SafeMarkdown";
import type { AIAgent, AIChatMessage } from "../api/types";
import { HttpError } from "../api/httpClient";

export default function AgentChatPage() {
  const agentId = Number(useParams().agentId);
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const rawSessionId = searchParams.get("sessionId");
  const sessionId = rawSessionId ? Number(rawSessionId) : undefined;
  const [historyQuery, setHistoryQuery] = useState("");
  const {
    history,
    chat,
    isLoading,
    error,
    sendMessage,
    isSending,
    sendError,
    retryMessage,
    isRetrying,
    deleteConversation,
    isDeleting,
  } = useAgentChat(agentId, sessionId);
  const runtimeMessages = useMemo(
    () => (chat?.messages ?? []).map(toThreadMessage),
    [chat?.messages],
  );
  const runtime = useExternalStoreRuntime<ThreadMessageLike>({
    messages: runtimeMessages,
    convertMessage: (message) => message,
    isRunning: isSending,
    onNew: async (message) => {
      const text = appendMessageText(message);
      if (!text || isSending) return;
      const result = await sendMessage(text);
      if (!sessionId) navigate(`/agents/${result.session.aiAgentId}/chat?sessionId=${result.session.id}`, { replace: true });
    },
  });
  const filteredHistory = useMemo(() => {
    const q = historyQuery.trim().toLowerCase();
    if (!q) return history;
    return history.filter((item) =>
      `${item.session.title} ${item.agent.displayName} ${item.lastMessage}`.toLowerCase().includes(q),
    );
  }, [history, historyQuery]);

  const startNewChat = async () => {
    if (isSending) return;
    navigate(`/agents/${agentId}/chat`);
  };

  if (Number.isNaN(agentId)) return <ChatState title="无效的 AI 角色" />;
  if (rawSessionId && Number.isNaN(sessionId)) return <ChatState title="无效的对话" />;
  if (isLoading) return <ChatState title="正在加载对话..." loading />;
  if (error instanceof HttpError && error.status === 401) {
    return <ChatState title="请先登录后再开始 AI 对话" actionLabel="去登录" actionTo="/login" />;
  }
  if (error || !chat) return <ChatState title="没有找到这个 AI 角色" />;

  return (
    <AssistantRuntimeProvider runtime={runtime}>
      <div className="flex flex-col overflow-hidden bg-cohere-surface text-cohere-on-surface" style={{ height: "100vh" }}>
        <ChatTopNav />
        <main className="flex min-h-0 flex-1 overflow-hidden">
          <aside
            className="hidden flex-shrink-0 flex-col border-r border-cohere-hairline bg-cohere-surface-lowest md:flex"
            style={{ width: "clamp(320px, 25vw, 400px)" }}
          >
            <div className="flex flex-col gap-sm border-b border-cohere-hairline p-md">
              <button
                type="button"
                onClick={startNewChat}
                disabled={isSending}
                className="flex w-full items-center justify-center gap-sm rounded-pill bg-cohere-primary px-md py-sm font-label-mono-bold text-cohere-on-primary transition-opacity hover:opacity-90"
              >
                <MaterialIcon name="add" size={16} />
                新对话
              </button>
              <div className="relative mt-xs">
                <MaterialIcon
                  name="search"
                  size={18}
                  className="absolute left-3 top-1/2 -translate-y-1/2 text-cohere-muted"
                />
                <input
                  value={historyQuery}
                  onChange={(event) => setHistoryQuery(event.target.value)}
                  className="w-full rounded-sm border border-cohere-hairline bg-cohere-surface py-sm pl-9 pr-sm font-caption text-cohere-on-surface outline-none transition-colors placeholder:text-cohere-muted focus:border-cohere-secondary"
                  placeholder="搜索历史..."
                  type="search"
                />
              </div>
            </div>

            <div className="min-h-0 flex-1 overflow-y-auto p-sm">
              <div className="mb-xs px-sm py-sm font-caption text-cohere-muted">对话历史</div>
              <div className="flex flex-col gap-xs">
                {filteredHistory.length === 0 ? (
                  <p className="px-sm py-md font-caption text-cohere-muted">暂无匹配对话</p>
                ) : (
                  filteredHistory.map((item) => (
                    <Link
                      key={item.session.id}
                      to={`/agents/${item.session.aiAgentId}/chat?sessionId=${item.session.id}`}
                      onClick={(event) => {
                        if (isSending) event.preventDefault();
                      }}
                      className={`rounded-sm border p-sm transition-colors ${
                        item.session.id === chat.session.id && chat.session.id > 0
                          ? "border-cohere-hairline bg-cohere-soft-stone"
                          : "border-transparent hover:bg-cohere-surface-low"
                      } ${isSending ? "pointer-events-none opacity-60" : ""}`}
                    >
                      <div className="truncate font-body-main font-medium text-cohere-on-surface">
                        {item.session.title || item.agent.displayName}
                      </div>
                      <div className="mt-xs flex justify-between gap-md font-caption text-cohere-muted">
                        <span className="truncate">{item.agent.displayName}</span>
                        <span className="flex-shrink-0">{formatHistoryTime(item.session.updatedAt)}</span>
                      </div>
                    </Link>
                  ))
                )}
              </div>
            </div>
          </aside>

          <section className="relative flex min-w-0 flex-1 flex-col bg-cohere-surface-lowest">
            <header
              className="z-10 flex flex-shrink-0 items-center justify-between border-b border-cohere-hairline bg-cohere-surface-lowest/90 px-md backdrop-blur-sm"
              style={{ height: 70 }}
            >
              <div className="flex min-w-0 items-center gap-md">
                <AgentAvatar agent={chat.agent} size="sm" />
                <div className="min-w-0">
                  <div className="truncate font-body-main text-[18px] font-medium text-cohere-primary">
                    {chat.agent.displayName}
                  </div>
                  <div className="flex items-center gap-xs font-caption text-cohere-muted">
                    <span className="h-2 w-2 rounded-full border border-cohere-secondary bg-cohere-success" />
                    在线
                  </div>
                </div>
              </div>
              {chat.session.id > 0 && (
                <button
                  type="button"
                  disabled={isSending || isDeleting}
                  onClick={async () => {
                    await deleteConversation(chat.session.id);
                    navigate(`/agents/${chat.agent.id}/chat`, { replace: true });
                  }}
                  className="flex h-8 w-8 items-center justify-center rounded-full text-cohere-muted hover:text-cohere-primary disabled:opacity-50"
                  aria-label="删除会话"
                >
                  <MaterialIcon name="delete" size={20} />
                </button>
              )}
            </header>

            <div className="min-h-0 flex-1 overflow-y-auto px-md py-xl md:px-section">
              {chat.messages.length === 0 ? (
                <div className="mx-auto flex h-full max-w-xl flex-col items-center justify-center gap-md text-center">
                  <AgentAvatar agent={chat.agent} size="lg" />
                  <h1 className="font-feature-title text-cohere-primary">开始和 {chat.agent.displayName} 对话</h1>
                  <p className="font-body-main text-cohere-on-surface-variant">
                    输入问题后，回复会按这个角色的人设生成并保存。
                  </p>
                </div>
              ) : (
                <div className="mx-auto flex max-w-[620px] flex-col gap-lg">
                  {chat.messages.map((message) => (
                    <ChatBubble
                      key={message.id}
                      message={message}
                      agent={chat.agent}
                      onRetry={() => retryMessage(message.id)}
                      retrying={isRetrying}
                    />
                  ))}
                  {isSending && <PendingBubble agent={chat.agent} />}
                </div>
              )}
            </div>

            {sendError && (
              <div className="border-t border-cohere-hairline bg-cohere-error-container px-md py-sm text-center font-caption text-cohere-error">
                回复生成失败，用户消息已保存。
              </div>
            )}

            <ComposerPrimitive.Root className="flex-shrink-0 border-t border-cohere-hairline bg-cohere-surface-lowest p-md">
              <div className="mx-auto max-w-[760px]">
                <div className="relative">
                  <ComposerPrimitive.Input
                    rows={1}
                    submitMode="enter"
                    disabled={isSending}
                    placeholder="输入消息..."
                    className="min-h-12 max-h-32 w-full resize-none rounded-pill border-none bg-cohere-soft-stone py-3 pl-md pr-14 font-body-main text-cohere-on-surface outline-none placeholder:text-cohere-muted focus:ring-1 focus:ring-cohere-secondary"
                  />
                  <ComposerPrimitive.Send
                    className={`absolute right-2 top-1/2 flex h-8 w-8 -translate-y-1/2 items-center justify-center rounded-full bg-cohere-primary text-cohere-on-primary transition-opacity hover:opacity-80 ${isSending ? "pointer-events-none opacity-50" : ""}`}
                  >
                    <MaterialIcon name="arrow_upward" size={18} />
                  </ComposerPrimitive.Send>
                </div>
                <div className="mt-sm text-center font-micro text-cohere-muted">
                  AI 可能会产生误导性信息，请核实重要内容。
                </div>
              </div>
            </ComposerPrimitive.Root>
          </section>

          <aside
            className="hidden flex-shrink-0 flex-col border-l border-cohere-hairline bg-cohere-surface-lowest lg:flex"
            style={{ width: "clamp(320px, 25vw, 400px)" }}
          >
            <header className="flex flex-shrink-0 items-center border-b border-cohere-hairline px-md" style={{ height: 70 }}>
              <h2 className="font-body-main text-[18px] font-medium text-cohere-primary">角色资料</h2>
            </header>
            <AgentProfile agent={chat.agent} />
          </aside>
        </main>
      </div>
    </AssistantRuntimeProvider>
  );
}

function ChatTopNav() {
  const nav = [
    { to: "/", label: "首页", end: true },
    { to: "/posts", label: "帖子" },
    { to: "/agents", label: "AI 角色" },
  ];

  return (
    <nav className="flex flex-shrink-0 items-center justify-between border-b border-cohere-hairline bg-cohere-surface px-margin-mobile md:px-margin-desktop" style={{ height: 80 }}>
      <div className="flex items-center gap-xl">
        <Link to="/" className="font-headline-lg-bold text-cohere-primary" aria-label="AI Forum 首页">
          AI Forum
        </Link>
        <div className="hidden items-center gap-md md:flex">
          {nav.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.end}
              className={({ isActive }) =>
                `pb-xs font-body-main transition-colors ${
                  isActive
                    ? "border-b-2 border-cohere-primary text-cohere-primary"
                    : "border-b-2 border-transparent text-cohere-on-surface-variant hover:text-cohere-primary"
                }`
              }
            >
              {item.label}
            </NavLink>
          ))}
        </div>
      </div>
      <div className="flex items-center gap-md">
        <button
          type="button"
          className="flex h-10 w-10 items-center justify-center rounded-full text-cohere-on-surface-variant hover:text-cohere-primary"
          aria-label="通知"
        >
          <MaterialIcon name="notifications" size={24} />
        </button>
        <Link
          to="/profile"
          className="rounded-pill bg-cohere-primary px-md py-sm font-mono text-label-mono font-semibold normal-case tracking-[0.02em] text-cohere-on-primary transition-opacity hover:opacity-80"
        >
          Profile
        </Link>
      </div>
    </nav>
  );
}

function AgentProfile({ agent }: { agent: AIAgent }) {
  const abilities = unique([...(agent.specialties ?? []), ...(agent.traits ?? [])]).slice(0, 6);
  const bio = agent.personality || agent.speakingStyle || agent.description;

  return (
    <div className="flex min-h-0 flex-1 flex-col gap-lg overflow-y-auto p-md">
      <div className="flex flex-col items-center gap-sm py-md text-center">
        <ProfileAvatar agent={agent} />
        <div>
          <h3 className="font-body-large font-medium text-cohere-primary">{agent.displayName}</h3>
          <p className="font-caption text-cohere-muted">{agentSubtitle(agent)}</p>
        </div>
      </div>

      <section className="space-y-sm">
        <h4 className="font-label-mono text-cohere-muted">个人简介</h4>
        <p className="font-caption leading-relaxed text-cohere-on-surface-variant">{bio}</p>
      </section>

      <section className="space-y-sm">
        <h4 className="font-label-mono text-cohere-muted">核心能力</h4>
        <div className="flex flex-wrap gap-sm">
          {(abilities.length ? abilities : ["分析", "推理", "讨论"]).map((ability) => (
            <span
              key={ability}
              className="rounded-xs border border-cohere-hairline bg-cohere-soft-stone px-sm py-xs font-label-mono text-cohere-on-surface-variant"
            >
              {ability}
            </span>
          ))}
        </div>
      </section>

      <section className="rounded-sm border border-cohere-hairline bg-cohere-surface-low p-md">
        <h4 className="mb-md font-label-mono text-cohere-muted">模型参数</h4>
        <ParamRow label="基础模型" value="Command R+" />
        <ParamRow label="采样温度" value={`${agent.temperature.toFixed(1)} (分析模式)`} />
        <ParamRow label="系统版本" value="v2.4.1-stable" last />
      </section>

      <div className="mt-auto pt-md">
        <div className="inline-flex items-center gap-sm rounded-pill border border-cohere-success bg-cohere-success/60 px-md py-xs">
          <span className="h-2 w-2 rounded-full bg-cohere-secondary" />
          <span className="font-label-mono text-cohere-on-secondary-container">
            {agent.displayName} 已就绪
          </span>
        </div>
      </div>
    </div>
  );
}

function ProfileAvatar({ agent }: { agent: AIAgent }) {
  if (!agent.avatar) return <AgentInitials agent={agent} size="lg" />;

  return (
    <div className="relative h-32 w-32 overflow-hidden rounded-full bg-cohere-surface">
      <img
        src={agent.avatar}
        alt={agent.displayName}
        className="h-full w-full object-cover object-center"
      />
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_center,rgba(251,249,244,0)_42%,rgba(251,249,244,0.35)_66%,var(--c-surface)_100%)]" />
      <div className="pointer-events-none absolute inset-x-0 bottom-0 h-14 bg-[linear-gradient(to_bottom,rgba(251,249,244,0),var(--c-surface)_92%)]" />
    </div>
  );
}

function ParamRow({ label, value, last = false }: { label: string; value: string; last?: boolean }) {
  return (
    <div className={`flex items-center justify-between gap-md py-sm ${last ? "" : "border-b border-dotted border-cohere-hairline"}`}>
      <span className="font-micro text-cohere-muted">{label}</span>
      <span className="text-right font-mono text-label-mono normal-case tracking-[0.02em] text-cohere-primary">{value}</span>
    </div>
  );
}

function ChatBubble({
  message,
  agent,
  onRetry,
  retrying,
}: {
  message: AIChatMessage;
  agent: AIAgent;
  onRetry: () => void;
  retrying: boolean;
}) {
  const isUser = message.role === "user";
  if (isUser) {
    return (
      <div className="flex justify-end">
        <div className="max-w-[520px] rounded-ai rounded-tr-sm bg-cohere-soft-stone px-lg py-md text-cohere-on-surface">
          <p className="whitespace-pre-wrap font-body-main">{message.content}</p>
          <div className="mt-sm text-right font-micro text-cohere-muted">{formatTime(message.createdAt)}</div>
        </div>
      </div>
    );
  }

  return (
    <div className="flex items-start gap-md">
      <AgentAvatar agent={agent} size="sm" className="mt-xs" />
      <div className="max-w-[560px] rounded-ai rounded-tl-sm border border-cohere-hairline bg-cohere-surface-lowest px-lg py-md text-cohere-on-surface">
        {message.content ? (
          <SafeMarkdown
            content={message.content}
            className="space-y-md font-body-main text-cohere-on-surface [&_h1]:font-body-large [&_h2]:font-body-large [&_h3]:font-body-large [&_li]:ml-md [&_ul]:list-disc [&_p]:text-cohere-on-surface"
          />
        ) : (
          <p className="font-body-main text-cohere-muted">正在生成回复...</p>
        )}
        {(message.status === "FAILED" || message.status === "PARTIAL") && (
          <div className="mt-md flex items-center justify-between gap-md rounded-sm bg-cohere-error-container px-sm py-xs font-caption text-cohere-error">
            <span>{message.errorMessage || "回复生成失败"}</span>
            <button
              type="button"
              disabled={retrying}
              onClick={onRetry}
              className="rounded-pill bg-cohere-primary px-sm py-xs font-label-mono-bold text-cohere-on-primary disabled:opacity-50"
            >
              {retrying ? "重试中" : "重新生成"}
            </button>
          </div>
        )}
        <div className="mt-md border-t border-cohere-hairline pt-sm text-right font-micro text-cohere-muted">
          {formatTime(message.createdAt)}
        </div>
      </div>
    </div>
  );
}

function PendingBubble({ agent }: { agent: AIAgent }) {
  return (
    <div className="flex items-center gap-md">
      <AgentAvatar agent={agent} size="sm" muted />
      <div className="flex items-center gap-sm font-label-mono text-cohere-muted">
        正在生成回复
        <span className="flex gap-xs">
          <span className="h-1.5 w-1.5 animate-bounce rounded-full bg-cohere-muted" />
          <span className="h-1.5 w-1.5 animate-bounce rounded-full bg-cohere-muted [animation-delay:0.2s]" />
          <span className="h-1.5 w-1.5 animate-bounce rounded-full bg-cohere-muted [animation-delay:0.4s]" />
        </span>
      </div>
    </div>
  );
}

function AgentAvatar({
  agent,
  size,
  muted = false,
  className = "",
}: {
  agent: AIAgent;
  size: "sm" | "lg";
  muted?: boolean;
  className?: string;
}) {
  if (!agent.avatar) return <AgentInitials agent={agent} size={size} muted={muted} className={className} />;

  const sizeClass = size === "lg" ? "h-24 w-24 border-4" : "h-8 w-8 border";
  return (
    <img
      src={agent.avatar}
      alt={agent.displayName}
      className={`${sizeClass} ${className} flex-shrink-0 rounded-full border-cohere-surface-highest object-cover object-center ${
        muted ? "opacity-60 grayscale" : ""
      }`}
    />
  );
}

function AgentInitials({
  agent,
  size,
  muted = false,
  className = "",
}: {
  agent: AIAgent;
  size: "sm" | "lg";
  muted?: boolean;
  className?: string;
}) {
  const sizeClass = size === "lg" ? "h-24 w-24 text-headline-lg border-4" : "h-8 w-8 text-label-mono border";
  return (
    <div
      className={`${sizeClass} ${className} flex flex-shrink-0 items-center justify-center rounded-full ${
        muted
          ? "border-cohere-hairline bg-cohere-surface-variant text-cohere-muted"
          : "border-cohere-surface-highest bg-cohere-deep-green text-cohere-on-primary"
      } font-label-mono-bold`}
      aria-hidden="true"
    >
      {initials(agent.displayName || agent.name)}
    </div>
  );
}

function ChatState({
  title,
  actionLabel = "返回角色广场",
  actionTo = "/agents",
  loading = false,
}: {
  title: string;
  actionLabel?: string;
  actionTo?: string;
  loading?: boolean;
}) {
  return (
    <div className="flex flex-col overflow-hidden bg-cohere-surface" style={{ height: "100vh" }}>
      <ChatTopNav />
      <main className="flex flex-1 flex-col items-center justify-center gap-md px-margin-mobile text-center">
        <MaterialIcon name={loading ? "progress_activity" : "error"} size={48} className="text-cohere-muted" />
        <h1 className="font-feature-title text-cohere-primary">{title}</h1>
        {!loading && (
          <Link to={actionTo} className="rounded-pill bg-cohere-primary px-lg py-sm font-label-mono-bold text-cohere-on-primary">
            {actionLabel}
          </Link>
        )}
      </main>
    </div>
  );
}

function initials(value: string): string {
  const ascii = value.match(/[A-Za-z]+/g);
  if (ascii?.length) {
    return ascii
      .slice(0, 2)
      .map((part) => part[0])
      .join("")
      .toUpperCase();
  }
  return value.trim().slice(0, 2) || "AI";
}

function agentSubtitle(agent: AIAgent): string {
  return agent.ageViewpoint.split("·")[1]?.trim() || agent.description || "AI 研究角色";
}

function unique(values: string[]): string[] {
  return Array.from(new Set(values.filter(Boolean)));
}

function toThreadMessage(message: AIChatMessage): ThreadMessageLike {
  return {
    id: String(message.id),
    role: message.role,
    content: message.content,
    createdAt: new Date(message.createdAt),
  };
}

function appendMessageText(message: AppendMessage): string {
  return message.content
    .map((part) => (part.type === "text" ? part.text : ""))
    .join("\n")
    .trim();
}

function formatTime(value: string): string {
  return new Intl.DateTimeFormat("zh-CN", { hour: "2-digit", minute: "2-digit" }).format(
    new Date(value),
  );
}

function formatHistoryTime(value: string): string {
  const date = new Date(value);
  const now = new Date();
  if (date.toDateString() === now.toDateString()) return formatTime(value);
  const yesterday = new Date(now);
  yesterday.setDate(now.getDate() - 1);
  if (date.toDateString() === yesterday.toDateString()) return "昨天";
  return new Intl.DateTimeFormat("zh-CN", { weekday: "short" }).format(date);
}
