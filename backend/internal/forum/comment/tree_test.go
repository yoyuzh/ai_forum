package comment

import "testing"

func TestBuildTreeAssemblesParentChildComments(t *testing.T) {
	parent := int64(1)
	tree := BuildTree([]Comment{
		{ID: 2, ParentCommentID: &parent, Content: "child"},
		{ID: 1, Content: "root"},
	})

	if len(tree) != 1 {
		t.Fatalf("roots = %d, want 1", len(tree))
	}
	if tree[0].ID != 1 || len(tree[0].Children) != 1 || tree[0].Children[0].ID != 2 {
		t.Fatalf("tree = %#v", tree)
	}
}
