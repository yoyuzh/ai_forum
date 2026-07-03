import { useState } from "react";
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
  const [content, setContent] = useState("");

  const wrap = (token: string) => {
    setContent((prev) => `${prev}${token}`);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = content.trim();
    if (!trimmed) return;
    await onSubmit(trimmed);
    setContent("");
  };

  return (
    <form
      onSubmit={handleSubmit}
      className="mb-xl flex flex-col rounded-lg border border-cohere-hairline bg-cohere-surface-lowest p-md transition-[border-color,box-shadow] focus-within:border-cohere-primary focus-within:shadow-[0_0_0_1px_rgba(0,0,0,1)]"
    >
      <textarea
        id="comment-content"
        name="comment"
        value={content}
        onChange={(e) => setContent(e.target.value)}
        placeholder={placeholder}
        aria-label="评论内容"
        className="min-h-[80px] w-full resize-none border-none bg-transparent p-0 font-body-main text-cohere-ink placeholder:text-cohere-on-surface-variant focus:outline-none focus:ring-0"
      />
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
            onClick={() => wrap("@ArchTechLead ")}
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
