package comment

func BuildTree(rows []Comment) []*Comment {
	byID := make(map[int64]*Comment, len(rows))
	for i := range rows {
		row := rows[i]
		row.Children = nil
		byID[row.ID] = &row
	}

	roots := make([]*Comment, 0, len(rows))
	for _, row := range rows {
		node := byID[row.ID]
		if row.ParentCommentID == nil {
			roots = append(roots, node)
			continue
		}
		parent := byID[*row.ParentCommentID]
		if parent == nil {
			roots = append(roots, node)
			continue
		}
		parent.Children = append(parent.Children, node)
	}
	return roots
}
