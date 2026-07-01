import { Comment } from "../../api/types";
import { formatRelativeTime } from "../../utils/format";
import SafeMarkdown from "../ui/SafeMarkdown";

interface HumanCommentProps {
  comment: Comment;
}

/** Chat-bubble human comment — rounded on three corners, avatar on the left. */
export default function HumanComment({ comment }: HumanCommentProps) {
  return (
    <div className="group flex gap-md">
      <div className="flex-shrink-0">
        <img
          src={comment.author.avatar}
          alt={comment.author.username}
          width={40}
          height={40}
          className="h-10 w-10 rounded-full border border-cohere-hairline object-cover"
        />
      </div>
      <div className="flex-1 rounded-br-lg rounded-bl-lg rounded-tr-lg border border-cohere-hairline bg-cohere-surface-lowest p-md transition-all duration-300 ease-cohere hover:border-cohere-secondary hover:shadow-[0_2px_12px_rgba(23,23,28,0.02)]">
        <div className="mb-1 flex items-center justify-between">
          <div className="flex items-center gap-1">
            <span className="font-label-mono-bold text-cohere-ink">{comment.author.username}</span>
            {comment.author.role && (
              <span className="font-micro text-cohere-muted">· {comment.author.role}</span>
            )}
          </div>
          <div className="font-micro text-cohere-muted">{formatRelativeTime(comment.createdAt)}</div>
        </div>
        <SafeMarkdown content={comment.content} />
        <div className="mt-md flex gap-md font-label-mono text-micro text-cohere-muted opacity-0 translate-y-[2px] transition-all duration-300 ease-cohere group-hover:opacity-100 group-hover:translate-y-0">
          <button type="button" className="transition-colors hover:text-cohere-ink focus:outline-none focus-visible:underline">回复</button>
          <button type="button" className="transition-colors hover:text-cohere-ink focus:outline-none focus-visible:underline">点赞 ({comment.likeCount})</button>
        </div>
      </div>
    </div>
  );
}
