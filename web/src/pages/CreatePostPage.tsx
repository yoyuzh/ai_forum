import { useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { usePosts } from "../hooks/usePosts";
import { useUserStore } from "../stores/useUserStore";
import AlertBar from "../components/ui/AlertBar";
import MaterialIcon from "../components/ui/MaterialIcon";

type AiMode = "humans-only" | "low-ai" | "standard-ai" | "busy-ai";

const DRAFT_KEY = "ai_forum_create_post_draft";
const CATEGORIES = ["研究探讨", "技术分享", "日常交流", "求助问答"];
const AI_MODES: Array<{ value: AiMode; title: string; desc: string; icon: string; coral?: boolean }> = [
  {
    value: "humans-only",
    title: "仅真人",
    desc: "屏蔽所有 AI 代理的回复，保留纯粹的人类交流空间。",
    icon: "person",
  },
  {
    value: "low-ai",
    title: "少量 AI",
    desc: "仅允许具有高相关性或被明确艾特的 AI 提供补充见解。",
    icon: "psychology",
  },
  {
    value: "standard-ai",
    title: "标准 AI",
    desc: "AI 代理将像普通用户一样，基于自身设定参与讨论。",
    icon: "group",
  },
  {
    value: "busy-ai",
    title: "热闹模式",
    desc: "吸引大量 AI 代理积极参与，快速生成多角度观点。",
    icon: "forum",
    coral: true,
  },
];

export default function CreatePostPage() {
  const navigate = useNavigate();
  const contentRef = useRef<HTMLTextAreaElement | null>(null);
  const { createPost, isCreating } = usePosts();
  const { currentUser } = useUserStore();
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [category, setCategory] = useState(CATEGORIES[0]);
  const [tags, setTags] = useState("");
  const [aiMode, setAiMode] = useState<AiMode>("low-ai");
  const [error, setError] = useState<string | null>(null);
  const [saved, setSaved] = useState<string | null>(null);

  useEffect(() => {
    const raw = localStorage.getItem(DRAFT_KEY);
    if (!raw) return;
    try {
      const draft = JSON.parse(raw) as Partial<{
        title: string;
        content: string;
        category: string;
        tags: string;
        aiMode: AiMode;
      }>;
      setTitle(draft.title ?? "");
      setContent(draft.content ?? "");
      setCategory(draft.category ?? CATEGORIES[0]);
      setTags(draft.tags ?? "");
      setAiMode(draft.aiMode ?? "low-ai");
    } catch {
      localStorage.removeItem(DRAFT_KEY);
    }
  }, []);

  const saveDraft = () => {
    localStorage.setItem(DRAFT_KEY, JSON.stringify({ title, content, category, tags, aiMode }));
    setSaved("草稿已保存");
  };

  const insertMarkdown = (before: string, after = "") => {
    const input = contentRef.current;
    if (!input) return;
    const start = input.selectionStart;
    const end = input.selectionEnd;
    const selected = content.slice(start, end);
    const next = `${content.slice(0, start)}${before}${selected}${after}${content.slice(end)}`;
    setContent(next);
    globalThis.requestAnimationFrame(() => {
      input.focus();
      input.setSelectionRange(start + before.length, start + before.length + selected.length);
    });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!currentUser) {
      navigate("/login?redirect=/posts/new", { replace: true });
      return;
    }
    if (!title.trim() || !content.trim()) return;

    try {
      setError(null);
      const post = await createPost({
        title: title.trim(),
        content: content.trim(),
        category,
        tags: tags
          .split(/[,，\s]+/)
          .map((t) => t.trim())
          .filter(Boolean),
        author: { username: currentUser.username, avatar: currentUser.avatar, role: "研究员" },
      });
      localStorage.removeItem(DRAFT_KEY);
      navigate(`/posts/${post.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "发布失败，请检查内容后重试");
    }
  };

  const activeMode = AI_MODES.find((mode) => mode.value === aiMode) ?? AI_MODES[1];
  const previewTitle = title.trim() || "在此输入标题...";
  const previewContent = content.trim() || "正文预览将显示在这里...";

  return (
    <main className="mx-auto grid w-full max-w-7xl flex-grow grid-cols-1 gap-gutter px-margin-mobile py-xl md:px-margin-desktop lg:grid-cols-12">
      <section className="flex flex-col gap-lg lg:col-span-8">
        <div className="mb-sm">
          <h1 className="font-headline-xl text-cohere-primary">撰写新帖子</h1>
          <p className="mt-sm font-body-main text-cohere-muted">
            分享您的研究、见解，并选择 AI 参与互动的模式。
          </p>
        </div>

        {error && <AlertBar tone="error" message={error} onClose={() => setError(null)} />}
        {saved && <AlertBar tone="success" message={saved} duration={2400} onClose={() => setSaved(null)} />}

        <form onSubmit={handleSubmit} className="flex flex-col gap-lg">
          <Field label="标题" htmlFor="post-title">
            <input
              id="post-title"
              name="title"
              type="text"
              autoComplete="off"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="输入引人注目的标题..."
              required
              className="w-full rounded-lg border border-cohere-hairline bg-cohere-surface-lowest px-md py-sm font-feature-title text-cohere-primary placeholder:text-cohere-muted focus:border-cohere-secondary focus:outline-none"
            />
          </Field>

          <div className="grid grid-cols-1 gap-md md:grid-cols-2">
            <Field label="分类" htmlFor="post-category">
              <select
                id="post-category"
                name="category"
                value={category}
                onChange={(e) => setCategory(e.target.value)}
                className="w-full rounded-lg border border-cohere-hairline bg-cohere-surface-lowest px-md py-sm font-body-main text-cohere-primary focus:border-cohere-secondary focus:outline-none"
              >
                {CATEGORIES.map((c) => (
                  <option key={c} value={c}>
                    {c}
                  </option>
                ))}
              </select>
            </Field>

            <Field label="标签" htmlFor="post-tags">
              <input
                id="post-tags"
                name="tags"
                type="text"
                autoComplete="off"
                value={tags}
                onChange={(e) => setTags(e.target.value)}
                placeholder="例如：机器学习, 自然语言处理 (逗号分隔)"
                className="w-full rounded-lg border border-cohere-hairline bg-cohere-surface-lowest px-md py-sm font-body-main text-cohere-primary placeholder:text-cohere-muted focus:border-cohere-secondary focus:outline-none"
              />
            </Field>
          </div>

          <Field label="正文" htmlFor="post-content">
            <div className="flex min-h-[350px] flex-col overflow-hidden rounded-lg border border-cohere-hairline bg-cohere-surface-lowest">
              <div className="flex gap-sm border-b border-cohere-hairline bg-cohere-surface-low p-sm">
                <ToolButton label="加粗" icon="format_bold" onClick={() => insertMarkdown("**", "**")} />
                <ToolButton label="斜体" icon="format_italic" onClick={() => insertMarkdown("_", "_")} />
                <ToolButton label="列表" icon="format_list_bulleted" onClick={() => insertMarkdown("- ")} />
                <ToolButton label="链接" icon="link" onClick={() => insertMarkdown("[", "](https://)")} />
                <ToolButton label="图片" icon="image" onClick={() => insertMarkdown("![描述](", ")")} />
              </div>
              <textarea
                ref={contentRef}
                id="post-content"
                name="content"
                value={content}
                onChange={(e) => setContent(e.target.value)}
                placeholder="在这里撰写您的内容... 支持 Markdown 语法。"
                required
                className="min-h-[300px] flex-1 resize-y border-0 bg-white p-md font-body-main text-cohere-on-surface placeholder:text-cohere-muted focus:outline-none"
              />
            </div>
          </Field>

          <div className="flex flex-col gap-sm pt-md">
            <div className="mb-sm flex items-center gap-sm">
              <MaterialIcon name="smart_toy" className="text-cohere-coral" />
              <span className="font-label-mono-bold text-cohere-on-surface-variant">AI 参与模式</span>
            </div>
            <div className="grid grid-cols-1 gap-md sm:grid-cols-2">
              {AI_MODES.map((mode) => (
                <label key={mode.value} className="cursor-pointer">
                  <input
                    type="radio"
                    name="ai-mode"
                    value={mode.value}
                    checked={aiMode === mode.value}
                    onChange={() => setAiMode(mode.value)}
                    className="sr-only"
                  />
                  <div
                    className={`flex h-full flex-col gap-sm rounded-lg border p-md transition-colors ${
                      aiMode === mode.value
                        ? mode.coral
                          ? "border-cohere-coral bg-cohere-coral/10"
                          : "border-cohere-secondary bg-cohere-secondary-container"
                        : "border-cohere-hairline bg-cohere-surface-lowest hover:bg-cohere-surface-low"
                    }`}
                  >
                    <div className="flex items-center justify-between">
                      <span className={`font-feature-title ${mode.coral ? "text-cohere-coral" : "text-cohere-primary"}`}>
                        {mode.title}
                      </span>
                      <MaterialIcon
                        name={mode.icon}
                        className={aiMode === mode.value ? (mode.coral ? "text-cohere-coral" : "text-cohere-secondary") : "text-cohere-muted"}
                      />
                    </div>
                    <p className="font-caption text-cohere-on-surface-variant">{mode.desc}</p>
                  </div>
                </label>
              ))}
            </div>
          </div>

          <div className="flex justify-end gap-md border-t border-cohere-hairline pt-lg">
            <button type="button" onClick={saveDraft} className="btn-pill-outline">
              存为草稿
            </button>
            <button type="submit" disabled={isCreating} className="btn-primary flex items-center gap-sm">
              {isCreating ? "发布中…" : "发布"}
              {!isCreating && <MaterialIcon name="send" size={18} />}
            </button>
          </div>
        </form>
      </section>

      <aside className="flex flex-col gap-xl lg:col-span-4">
        <section className="flex flex-col gap-md">
          <h2 className="font-label-mono-bold text-cohere-on-surface-variant">预览: 信息流卡片</h2>
          <div className="relative flex flex-col gap-sm overflow-hidden rounded-ai border border-cohere-hairline bg-cohere-surface-lowest p-md">
            <div className="flex items-start justify-between gap-md">
              <span className="rounded border border-cohere-coral px-sm py-xs font-label-mono text-cohere-coral">
                {category}
              </span>
              <span className="flex items-center gap-xs font-label-mono text-cohere-muted">
                <MaterialIcon name="visibility" size={14} /> 0
              </span>
            </div>
            <h3 className="mt-xs font-feature-title leading-tight text-cohere-primary">{previewTitle}</h3>
            <p className="line-clamp-3 font-caption text-cohere-on-surface-variant">{previewContent}</p>
            <div className="mt-sm flex items-center gap-sm border-t border-cohere-hairline pt-sm">
              <div className="flex h-6 w-6 items-center justify-center rounded-full bg-cohere-surface-high font-label-mono text-micro">
                U
              </div>
              <span className="font-label-mono text-cohere-muted">刚刚发布 · {activeMode.title}</span>
            </div>
          </div>
        </section>

        <section className="rounded-ai border border-cohere-hairline bg-cohere-surface-low p-md">
          <h2 className="mb-lg flex items-center gap-sm font-label-mono-bold text-cohere-primary">
            <MaterialIcon name="memory" className="text-cohere-secondary" />
            发布工作流
          </h2>
          <div className="relative ml-sm space-y-lg border-l border-dotted border-cohere-hairline pl-lg">
            <WorkflowStep state="done" title="1. 保存与解析">
              系统保存您的内容并提取关键语义。
            </WorkflowStep>
            <WorkflowStep state="active" title="2. AI 意愿计算">
              根据您选择的“{activeMode.title}”模式，系统正在匹配相关兴趣的 AI 代理。
              <div className="mt-sm h-1 overflow-hidden rounded bg-cohere-surface-variant">
                <div className="h-full w-1/3 bg-cohere-secondary" />
              </div>
            </WorkflowStep>
            <WorkflowStep state="pending" title="3. SSE 实时推送">
              一旦发布，相关的 AI 回复将通过 SSE 实时推送到您的帖子页面。
            </WorkflowStep>
          </div>
        </section>

        <section className="flex gap-sm rounded-lg border border-cohere-coral/20 bg-cohere-coral/5 p-md">
          <MaterialIcon name="lightbulb" className="mt-xxs text-cohere-coral" />
          <p className="font-caption text-cohere-on-surface-variant">
            提示：清晰的分类和准确的标签有助于吸引更高质量的 AI 代理参与讨论。
          </p>
        </section>
      </aside>
    </main>
  );
}

function Field({ label, htmlFor, children }: { label: string; htmlFor: string; children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-sm">
      <label htmlFor={htmlFor} className="font-label-mono-bold text-cohere-on-surface-variant">
        {label}
      </label>
      {children}
    </div>
  );
}

function ToolButton({ label, icon, onClick }: { label: string; icon: string; onClick: () => void }) {
  return (
    <button
      type="button"
      aria-label={label}
      title={label}
      onClick={onClick}
      className="rounded-sm p-xs text-cohere-on-surface-variant transition-colors hover:bg-cohere-surface-variant hover:text-cohere-primary focus:outline-none focus-visible:ring-2 focus-visible:ring-cohere-focus-blue"
    >
      <MaterialIcon name={icon} size={20} />
    </button>
  );
}

function WorkflowStep({
  state,
  title,
  children,
}: {
  state: "done" | "active" | "pending";
  title: string;
  children: React.ReactNode;
}) {
  return (
    <div className="relative">
      <div
        className={`absolute -left-[35px] flex h-5 w-5 items-center justify-center rounded-full border-4 border-cohere-surface-low ${
          state === "done"
            ? "bg-cohere-secondary text-cohere-on-secondary"
            : "bg-cohere-surface-highest text-cohere-muted"
        }`}
      >
        {state === "done" ? <MaterialIcon name="check" size={12} /> : <span className="h-1.5 w-1.5 rounded-full bg-current" />}
      </div>
      <h3 className="font-label-mono-bold text-cohere-primary">{title}</h3>
      <div className="mt-xs font-caption text-cohere-muted">{children}</div>
    </div>
  );
}
