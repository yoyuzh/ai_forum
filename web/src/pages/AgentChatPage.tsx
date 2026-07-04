import { useMemo } from "react";
import { Link, useParams } from "react-router-dom";
import {
  AssistantRuntimeProvider,
  ComposerPrimitive,
  useExternalStoreRuntime,
  type AppendMessage,
  type ThreadMessageLike,
} from "@assistant-ui/react";
import { useAgentChat } from "../hooks/useAgentChat";
import MaterialIcon from "../components/ui/MaterialIcon";
import type { AIChatMessage } from "../api/types";
import { HttpError } from "../api/httpClient";

export default function AgentChatPage() {
  const agentId = Number(useParams().agentId);
  const { history, chat, isLoading, error, sendMessage, isSending, sendError } = useAgentChat(agentId);
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
      if (text) await sendMessage(text);
    },
  });

  if (Number.isNaN(agentId)) return <ChatError title="无效的 AI 角色" />;
  if (isLoading) return <ChatShellSkeleton />;
  if (error instanceof HttpError && error.status === 401) {
    return <ChatError title="请先登录后再开始 AI 对话" actionLabel="去登录" actionTo="/login" />;
  }
  if (error || !chat) return <ChatError title="没有找到这个 AI 角色" />;

  return (
    <AssistantRuntimeProvider runtime={runtime}>
      <main className="mx-auto flex w-full max-w-7xl flex-grow flex-col px-margin-mobile py-lg md:px-margin-desktop">
        <div className="mb-md">
          <Link to="/agents" className="btn-link inline-flex items-center gap-xs">
            <MaterialIcon name="arrow_back" size={18} />
            AI 角色广场
          </Link>
        </div>

        <section className="grid h-[calc(100vh-180px)] overflow-hidden rounded-lg border border-cohere-hairline bg-cohere-surface-lowest lg:grid-cols-[280px_minmax(0,1fr)_280px]">
          <aside className="hidden border-r border-cohere-hairline bg-cohere-surface-low p-lg lg:flex lg:flex-col">
            <img
              src={chat.agent.avatar}
              alt={chat.agent.displayName}
              className="h-56 w-full rounded-lg object-cover object-center"
            />
            <h1 className="mt-lg font-headline-lg-bold text-cohere-primary">
              {chat.agent.displayName}
            </h1>
            <p className="mt-sm font-body-main text-cohere-on-surface-variant">
              {chat.agent.description}
            </p>
            <div className="mt-lg flex flex-wrap gap-xs">
              {chat.agent.traits.slice(0, 3).map((trait) => (
                <span key={trait} className="rounded-sm bg-cohere-surface-highest px-xs py-xxs font-label-mono">
                  {trait}
                </span>
              ))}
            </div>
          </aside>

          <div className="flex min-h-0 flex-col overflow-hidden">
            <header className="flex items-center gap-md border-b border-cohere-hairline px-md py-sm">
              <img
                src={chat.agent.avatar}
                alt=""
                className="h-10 w-10 rounded-full object-cover"
              />
              <div className="min-w-0">
                <div className="truncate font-feature-title text-cohere-primary">
                  {chat.agent.displayName}
                </div>
                <div className="font-label-mono text-cohere-muted">一对一对话</div>
              </div>
            </header>

            <div className="min-h-0 flex-1 overflow-y-auto bg-cohere-surface-lowest px-md py-lg md:px-xl">
              {chat.messages.length === 0 ? (
                <div className="mx-auto flex h-full max-w-xl flex-col items-center justify-center gap-md text-center">
                  <MaterialIcon name={chat.agent.icon || "smart_toy"} size={48} className="text-cohere-muted" />
                  <h2 className="font-feature-title text-cohere-primary">开始和 {chat.agent.displayName} 对话</h2>
                  <p className="font-body-main text-cohere-on-surface-variant">
                    直接输入问题，回复会按这个角色的人设生成并保存。
                  </p>
                </div>
              ) : (
                <div className="mx-auto flex max-w-3xl flex-col gap-md">
                  {chat.messages.map((message) => (
                    <ChatBubble key={message.id} message={message} avatar={chat.agent.avatar} />
                  ))}
                  {isSending && <PendingBubble avatar={chat.agent.avatar} />}
                </div>
              )}
            </div>

            {sendError && (
              <div className="border-t border-cohere-hairline bg-cohere-error-container px-md py-sm font-caption text-cohere-error">
                回复生成失败，用户消息已保存。
              </div>
            )}

            <ComposerPrimitive.Root className="border-t border-cohere-hairline bg-cohere-surface-low p-md">
              <div className="mx-auto flex max-w-3xl items-end gap-sm rounded-lg border border-cohere-hairline bg-white p-sm focus-within:border-cohere-secondary">
                <ComposerPrimitive.Input
                  rows={1}
                  submitMode="enter"
                  placeholder={`和 ${chat.agent.displayName} 说点什么...`}
                  className="min-h-11 flex-1 resize-none bg-transparent px-sm py-sm font-body-main text-cohere-on-surface outline-none placeholder:text-cohere-muted"
                />
                <ComposerPrimitive.Send className="btn-primary inline-flex min-h-11 items-center gap-xs">
                  <MaterialIcon name="send" size={18} />
                  发送
                </ComposerPrimitive.Send>
              </div>
            </ComposerPrimitive.Root>
          </div>

          <aside className="hidden min-h-0 border-l border-cohere-hairline bg-cohere-surface-low p-md lg:flex lg:flex-col">
            <div className="mb-md flex items-center justify-between">
              <h2 className="font-label-mono-bold text-cohere-primary">历史记录</h2>
              <MaterialIcon name="history" size={18} className="text-cohere-muted" />
            </div>
            {history.length === 0 ? (
              <p className="font-caption text-cohere-muted">发送第一条消息后，这里会出现真实保存的历史。</p>
            ) : (
              <div className="min-h-0 flex-1 overflow-y-auto pr-xs flex flex-col gap-xs">
                {history.map((item) => (
                  <Link
                    key={item.session.id}
                    to={`/agents/${item.session.aiAgentId}/chat`}
                    className={`rounded-lg border p-sm transition-colors ${
                      item.session.aiAgentId === agentId
                        ? "border-cohere-secondary bg-cohere-secondary-container text-cohere-on-secondary-container"
                        : "border-cohere-hairline bg-white text-cohere-on-surface hover:border-cohere-slate"
                    }`}
                  >
                    <div className="truncate font-body-main">{item.agent.displayName}</div>
                    <div className="mt-xxs line-clamp-2 font-caption text-cohere-muted">
                      {item.lastMessage || "还没有消息"}
                    </div>
                    <div className="mt-xs font-micro text-cohere-muted">
                      {item.messageCount} 条消息
                    </div>
                  </Link>
                ))}
              </div>
            )}
          </aside>
        </section>
      </main>
    </AssistantRuntimeProvider>
  );
}

function ChatBubble({ message, avatar }: { message: AIChatMessage; avatar: string }) {
  const isUser = message.role === "user";
  return (
    <div className={`flex items-start gap-sm ${isUser ? "justify-end" : "justify-start"}`}>
      {!isUser && <img src={avatar} alt="" className="mt-xs h-8 w-8 rounded-full object-cover" />}
      <div
        className={`max-w-[82%] rounded-lg px-md py-sm font-body-main ${
          isUser
            ? "bg-cohere-surface-low text-cohere-on-surface"
            : "border border-cohere-hairline bg-white text-cohere-on-surface"
        }`}
      >
        <p className="whitespace-pre-wrap">{message.content}</p>
        <div className="mt-xs font-micro text-cohere-muted">{formatTime(message.createdAt)}</div>
      </div>
    </div>
  );
}

function PendingBubble({ avatar }: { avatar: string }) {
  return (
    <div className="flex items-start gap-sm">
      <img src={avatar} alt="" className="mt-xs h-8 w-8 rounded-full object-cover" />
      <div className="rounded-lg border border-cohere-hairline bg-white px-md py-sm font-body-main text-cohere-muted">
        正在生成回复...
      </div>
    </div>
  );
}

function ChatError({
  title,
  actionLabel = "返回角色广场",
  actionTo = "/agents",
}: {
  title: string;
  actionLabel?: string;
  actionTo?: string;
}) {
  return (
    <main className="mx-auto flex w-full max-w-3xl flex-grow flex-col items-center justify-center gap-md px-margin-mobile py-section text-center">
      <MaterialIcon name="error" size={48} className="text-cohere-muted" />
      <h1 className="font-feature-title text-cohere-primary">{title}</h1>
      <Link to={actionTo} className="btn-primary">
        {actionLabel}
      </Link>
    </main>
  );
}

function ChatShellSkeleton() {
  return (
    <main className="mx-auto flex w-full max-w-7xl flex-grow flex-col px-margin-mobile py-lg md:px-margin-desktop">
      <div className="h-[calc(100vh-180px)] animate-pulse-soft rounded-lg border border-cohere-hairline bg-cohere-surface-low" />
    </main>
  );
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
