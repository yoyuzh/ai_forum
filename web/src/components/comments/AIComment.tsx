import { Comment } from "../../api/types";
import { formatRelativeTime } from "../../utils/format";
import MaterialIcon from "../ui/MaterialIcon";
import SafeMarkdown from "../ui/SafeMarkdown";

interface AICommentProps {
  comment: Comment;
}

const TRIGGER_LABEL: Record<NonNullable<Comment["triggerType"]>, string> = {
  POST_AUTO: "发帖后自动回复",
  MENTION: "被提及回复",
  FOLLOWUP: "追问回复",
};

/** Distinct green-tinted AI bubble — psychology icon avatar, willingness score,
 *  "继续追问" affordance. Matches the _2/_3 prototype AI comment block. */
export default function AIComment({ comment }: AICommentProps) {
  const score = comment.willingnessScore;
  return (
    <div className="relative flex gap-md group">
      <div className="z-10 flex-shrink-0 relative h-10 w-10 overflow-hidden rounded-ai border-2 border-cohere-surface-lowest shadow-sm">
        <img
          src={comment.author.avatar}
          alt={comment.author.username}
          width={40}
          height={40}
          onError={(e) => {
            e.currentTarget.style.display = "none";
            const fb = e.currentTarget.parentElement?.querySelector(".avatar-fallback");
            if (fb) fb.classList.remove("hidden");
          }}
          className="h-10 w-10 object-cover"
        />
        <div className="avatar-fallback hidden absolute inset-0 flex items-center justify-center bg-cohere-secondary text-cohere-on-secondary">
          <MaterialIcon name="psychology" size={20} />
        </div>
      </div>
      <div className="flex-1 rounded-br-ai rounded-bl-ai rounded-tr-ai border border-[#e0e5e3] bg-[#f5f7f6] p-md transition-all duration-300 ease-cohere hover:border-[#cbdad5] hover:shadow-[0_2px_12px_rgba(0,60,51,0.03)]">
        <div className="mb-md flex flex-wrap items-center gap-1">
          <span className="font-label-mono-bold text-cohere-secondary">
            {comment.author.role ?? comment.author.username}
          </span>
          {comment.triggerType && (
            <span className="rounded bg-cohere-secondary-container px-1 py-0.5 font-label-mono text-[10px] text-cohere-on-secondary-container">
              {TRIGGER_LABEL[comment.triggerType]}
            </span>
          )}
          <div className="ml-auto font-micro text-cohere-muted">
            {formatRelativeTime(comment.createdAt)}
          </div>
        </div>

        <SafeMarkdown content={comment.content} />

        <div className="mt-lg flex items-center justify-between border-t border-[#e0e5e3] pt-md">
          <div className="flex items-center gap-1 font-label-mono text-micro text-cohere-muted">
            <MaterialIcon name="analytics" size={14} />
            意愿分: {score !== undefined ? `${Math.round(score * 100)}/100` : "—"}
          </div>
          <button
            type="button"
            className="flex items-center gap-1 rounded-pill border border-cohere-hairline bg-cohere-surface-lowest px-md py-1 font-label-mono-bold text-micro text-cohere-ink transition-all duration-300 ease-spring hover:border-cohere-secondary hover:text-cohere-secondary hover:-translate-y-[1px] active:translate-y-0 active:scale-[0.98] focus:outline-none focus-visible:ring-2 focus-visible:ring-cohere-focus-blue"
          >
            继续追问 <MaterialIcon name="arrow_forward" size={14} />
          </button>
        </div>
      </div>
    </div>
  );
}
