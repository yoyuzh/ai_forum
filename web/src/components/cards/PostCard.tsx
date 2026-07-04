import { Link } from "react-router-dom";
import { AIResponder, Post } from "../../api/types";
import { formatRelativeTime, formatCount } from "../../utils/format";
import CategoryBadge from "../ui/CategoryBadge";
import StatusBadge from "../ui/StatusBadge";
import TagPill from "../ui/TagPill";
import MaterialIcon from "../ui/MaterialIcon";

interface PostCardProps {
  post: Post;
}

/** Feed card — matches the Stitch ai_forum_5 prototype post article. */
export default function PostCard({ post }: PostCardProps) {
  const responders: AIResponder[] =
    post.aiResponders && post.aiResponders.length > 0
      ? post.aiResponders
      : post.aiAvatars.map((avatar, idx) => ({
          name: `AI ${idx + 1}`,
          avatar,
        }));
  const tagAccent = responders[0]?.accentColor;

  return (
    <Link
      to={`/posts/${post.id}`}
      className="card-base block p-lg hover:border-cohere-secondary hover:-translate-y-[2px] hover:shadow-sm transition-all duration-300 ease-cohere focus:outline-none focus-visible:ring-2 focus-visible:ring-cohere-focus-blue"
    >
      <div className="flex items-start justify-between gap-md">
        <h2 className="font-feature-title text-cohere-primary">
          {post.title}
        </h2>
        <CategoryBadge category={post.category} />
      </div>

      <p className="mt-1 line-clamp-2 font-body-main text-cohere-on-surface-variant">
        {post.content}
      </p>

      <div className="mt-sm flex flex-wrap items-center gap-sm">
        <div className="mr-md flex items-center gap-xs">
          <img
            src={post.author.avatar}
            alt={post.author.username}
            width={24}
            height={24}
            className="h-6 w-6 rounded-full bg-cohere-surface-variant"
          />
          <span className="font-caption text-cohere-on-surface-variant">
            {post.author.role ? `${post.author.role} · ` : ""}
            {post.author.username}
          </span>
        </div>

        {post.tags.slice(0, 3).map((tag) => (
          <TagPill key={tag} tag={tag} accentColor={tagAccent} />
        ))}

        <div className="flex-grow" />

        <div className="flex items-center gap-xs">
          <span className="hidden items-center gap-1 font-micro text-cohere-on-surface-variant sm:flex">
            <MaterialIcon name="visibility" size={14} />
            {formatCount(post.viewCount)}
          </span>
          <span className="hidden items-center gap-1 font-micro text-cohere-on-surface-variant sm:flex">
            <MaterialIcon name="forum" size={14} />
            {post.commentCount}
          </span>

          <StatusBadge status={post.aiStatus} responsesCount={post.aiResponsesCount} responders={responders} />
        </div>
      </div>

      <div className="mt-2 font-micro text-cohere-on-surface-variant">{formatRelativeTime(post.createdAt)}</div>
    </Link>
  );
}
