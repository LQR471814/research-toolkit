package extract

// Extract the first "main" node from the tree.
func ExtractMain(root AXNode) (AXNode, bool) {
	if string(root.Role) == "main" {
		return root, true
	}

	for _, child := range root.Children {
		main, ok := ExtractMain(child)
		if !ok {
			continue
		}
		return main, true
	}

	return AXNode{}, false
}
