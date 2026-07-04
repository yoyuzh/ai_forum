import { useMemo, useRef, useState, type KeyboardEvent } from "react";
import type { AIAgent } from "../../api/types";
import { useAgents } from "../../hooks/useAgents";
import MaterialIcon from "../ui/MaterialIcon";

interface CommentEditorProps {
  onSubmit: (content: string) => Promise<void>;
  isSubmitting: boolean;
  placeholder?: string;
}

/** Comment composer with a lightweight formatting toolbar and @AI mention hint.
 *  The toolbar buttons are plain-text affordances — they prefix the textarea,
 *  no rich content is produced, so sanitization stays trivial downstream. */
export default function CommentEditor({
  onSubmit,
  isSubmitting,
  placeholder = "输入你的评论，输入 @ 唤醒特定 AI Agent…",
}: CommentEditorProps) {
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const [content, setContent] = useState("");
  const [mentionStart, setMentionStart] = useState<number | null>(null);
  const [activeMentionIndex, setActiveMentionIndex] = useState(0);
  const { agents } = useAgents();

  const mentionAgents = useMemo(
    () => agents.filter((agent) => agent.active && agent.allowMentionReply).slice(0, 12),
    [agents],
  );
  const mentionOpen = mentionStart !== null && mentionAgents.length > 0;

  const wrap = (token: string) => {
    setContent((prev) => `${prev}${token}`);
  };

  const syncMention = (nextContent: string, cursor: number) => {
    const start = currentMentionStart(nextContent, cursor);
    setMentionStart(start);
    if (start !== null) setActiveMentionIndex(0);
  };

  const insertMention = (agent: AIAgent) => {
    const input = textareaRef.current;
    const cursor = input?.selectionStart ?? content.length;
    const start = mentionStart ?? currentMentionStart(content, cursor) ?? cursor;
    const mention = `@${agent.name} `;
    const next = `${content.slice(0, start)}${mention}${content.slice(cursor)}`;
    const nextCursor = start + mention.length;

    setContent(next);
    setMentionStart(null);
    setActiveMentionIndex(0);
    window.requestAnimationFrame(() => {
      input?.focus();
      input?.setSelectionRange(nextCursor, nextCursor);
    });
  };

  const openMentionFromToolbar = () => {
    const input = textareaRef.current;
    const cursor = input?.selectionStart ?? content.length;
    const next = `${content.slice(0, cursor)}@${content.slice(cursor)}`;
    const nextCursor = cursor + 1;

    setContent(next);
    setMentionStart(cursor);
    setActiveMentionIndex(0);
    window.requestAnimationFrame(() => {
      input?.focus();
      input?.setSelectionRange(nextCursor, nextCursor);
    });
  };

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (!mentionOpen) return;
    if (e.key === "ArrowDown") {
      e.preventDefault();
      setActiveMentionIndex((index) => (index + 1) % mentionAgents.length);
      return;
    }
    if (e.key === "ArrowUp") {
      e.preventDefault();
      setActiveMentionIndex((index) => (index - 1 + mentionAgents.length) % mentionAgents.length);
      return;
    }
    if (e.key === "Enter" || e.key === "Tab") {
      e.preventDefault();
      insertMention(mentionAgents[activeMentionIndex]);
      return;
    }
    if (e.key === "Escape") {
      e.preventDefault();
      setMentionStart(null);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = content.trim();
    if (!trimmed) return;
    await onSubmit(trimmed);
    setContent("");
    setMentionStart(null);
  };

  return (
    <form
      onSubmit={handleSubmit}
      className="mb-xl flex flex-col rounded-lg border border-cohere-hairline bg-cohere-surface-lowest p-md transition-[border-color,box-shadow] focus-within:border-cohere-primary focus-within:shadow-[0_0_0_1px_rgba(0,0,0,1)]"
    >
      <div className="relative">
        <textarea
          ref={textareaRef}
          id="comment-content"
          name="comment"
          value={content}
          onChange={(e) => {
            setContent(e.target.value);
            syncMention(e.target.value, e.target.selectionStart);
          }}
          onKeyDown={handleKeyDown}
          placeholder={placeholder}
          aria-label="评论内容"
          aria-controls={mentionOpen ? "mention-dropdown-root" : undefined}
          aria-expanded={mentionOpen}
          data-testid="comment-input"
          className="min-h-[80px] w-full resize-none border-none bg-transparent p-0 font-body-main text-cohere-ink placeholder:text-cohere-on-surface-variant focus:outline-none focus:ring-0"
        />
        {mentionOpen && (
          <div
            id="mention-dropdown-root"
            role="listbox"
            aria-label="选择 AI 角色"
            className="absolute left-0 right-0 top-full z-30 mt-xs max-h-72 overflow-y-auto rounded-sm border border-cohere-hairline bg-cohere-surface-lowest p-xs shadow-sm"
          >
            <div className="grid grid-cols-1 gap-1 sm:grid-cols-2">
              {mentionAgents.map((agent, index) => (
                <button
                  key={agent.id}
                  type="button"
                  role="option"
                  aria-selected={index === activeMentionIndex}
                  data-testid={`mention-item-${agent.name}`}
                  onMouseEnter={() => setActiveMentionIndex(index)}
                  onMouseDown={(e) => e.preventDefault()}
                  onClick={() => insertMention(agent)}
                  className={`flex min-w-0 items-center gap-sm rounded-sm px-sm py-xs text-left transition-colors focus:outline-none ${
                    index === activeMentionIndex
                      ? "bg-cohere-pale-blue text-cohere-action-blue"
                      : "text-cohere-ink hover:bg-cohere-surface-variant"
                  }`}
                >
                  <img
                    src={agent.avatar}
                    alt=""
                    className="h-7 w-7 flex-shrink-0 rounded-full border border-cohere-hairline object-cover"
                  />
                  <span className="min-w-0">
                    <span className="block truncate font-label-mono-bold normal-case tracking-normal">
                      @{agent.name}
                    </span>
                    <span className="block truncate font-micro text-cohere-muted">{agent.ageViewpoint}</span>
                  </span>
                </button>
              ))}
            </div>
          </div>
        )}
      </div>
      <div className="mt-sm flex items-center justify-between border-t border-cohere-surface-variant pt-sm">
        <div className="flex gap-1">
          <button
            type="button"
            onClick={() => wrap("**")}
            className="p-xs text-cohere-on-surface-variant transition-colors hover:text-cohere-ink focus:outline-none focus-visible:ring-2 focus-visible:ring-cohere-focus-blue"
            aria-label="加粗"
            title="加粗"
          >
            <MaterialIcon name="format_bold" />
          </button>
          <button
            type="button"
            onClick={() => wrap("`code`")}
            className="p-xs text-cohere-on-surface-variant transition-colors hover:text-cohere-ink focus:outline-none focus-visible:ring-2 focus-visible:ring-cohere-focus-blue"
            aria-label="代码"
            title="代码"
          >
            <MaterialIcon name="code" />
          </button>
          <button
            type="button"
            onClick={openMentionFromToolbar}
            className="font-label-mono-bold p-xs text-cohere-action-blue transition-colors hover:text-cohere-focus-blue focus:outline-none focus-visible:ring-2 focus-visible:ring-cohere-focus-blue"
            aria-label="提及 AI"
            title="提及 AI"
          >
            @AI
          </button>
        </div>
        <button type="submit" disabled={isSubmitting || !content.trim()} className="btn-primary">
          {isSubmitting ? "发布中…" : "发布评论"}
        </button>
      </div>
    </form>
  );
}

function currentMentionStart(value: string, cursor: number): number | null {
  const start = value.lastIndexOf("@", cursor - 1);
  if (start === -1) return null;
  return /\s/.test(value.slice(start + 1, cursor)) ? null : start;
}
