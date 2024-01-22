package hierarchy

/*
 Hierarchy pattern library helper
*/

type StringNode struct {
	// Value for current node
	Value string

	// Childrens for the node
	Children []*StringNode
}

func NewStringHierarchy(value string, children ...*StringNode) *StringNode {
	return &StringNode{
		Value:    value,
		Children: children,
	}
}

// AddChild adds a child to the current node
func (n *StringNode) AddChild(child *StringNode) *StringNode {
	// Returns newly added child
	n.Children = append(n.Children, child)
	return child
}

// FindChild finds the first child of the node matching the value
func (n *StringNode) FindChild(value string) *StringNode {
	for _, child := range n.Children {
		if child.Value == value {
			return child
		}
	}
	return nil
}
