import { RelatedDiscussion } from "../../api/types";
import MaterialIcon from "../ui/MaterialIcon";
import { Link } from "react-router-dom";

interface RelatedDiscussionsProps {
  discussions: RelatedDiscussion[];
}

/** Related-discussion list — rule-separated links. */
export default function RelatedDiscussions({ discussions }: RelatedDiscussionsProps) {
  return (
    <div className="card-base p-lg">
      <h3 className="mb-md font-feature-title text-[18px] text-cohere-ink">相关讨论</h3>
      {discussions.length === 0 && (
        <p className="font-micro text-cohere-on-surface-variant">暂无相关讨论。</p>
      )}
      <ul className="space-y-md font-body-main text-[14px] text-cohere-ink">
        {discussions.map((d, idx) => (
          <li
            key={d.id}
            className={`flex items-start gap-xs transition-colors hover:text-cohere-action-blue ${
              idx < discussions.length - 1 ? "border-b border-cohere-hairline pb-sm" : ""
            }`}
          >
            <MaterialIcon name="link" size={16} className="mt-[2px] text-cohere-on-surface-variant" />
            <Link to={`/posts/${d.id}`} className="hover:text-cohere-action-blue">
              {d.title}
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
}
