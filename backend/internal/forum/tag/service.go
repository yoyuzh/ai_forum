package tag

func GroupByType(tags []Tag) map[string][]string {
	grouped := make(map[string][]string)
	for _, t := range tags {
		grouped[t.Type] = append(grouped[t.Type], t.Name)
	}
	return grouped
}
