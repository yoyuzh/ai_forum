import TagPill from "../ui/TagPill";

interface PostTagsProps {
  tags: string[];
}

/** Post tags panel — coral-outlined pills as in the post-detail prototype. */
export default function PostTags({ tags }: PostTagsProps) {
  return (
    <div className="card-base p-lg">
      <h3 className="mb-md font-feature-title text-[18px] text-cohere-ink">帖子标签</h3>
      <div className="flex flex-wrap gap-sm">
        {tags.map((tag, idx) => (
          <TagPill key={tag} tag={tag} variant={idx < 2 ? "coral" : "default"} />
        ))}
      </div>
    </div>
  );
}
