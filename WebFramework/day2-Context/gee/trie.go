package gee

import "strings"

type node struct {
	pattern  string  // 待匹配的路由
	part     string  // 路由中的一部分
	children []*node // 子节点
	isWild   bool    // 是否模糊匹配, ':' or '*'
}

// 用于插入
func (n *node) mathChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

// 用于查到所有匹配的子节点
func (n *node) mathChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

func (n *node) insert(pattern string, parts []string, height int) {
	if len(parts) == height { // 达到叶子节点

		n.pattern = pattern
		return
	}

	part := parts[height]      // 当前要匹配的 part
	child := n.mathChild(part) // 看是否已经存在于当前节点的的孩子节点中

	if child == nil {
		child = &node{part: part, isWild: part[0] == ':' || part[0] == '*'}
		n.children = append(n.children, child)
	}
	child.insert(pattern, parts, height+1)
}

func (n *node) search(parts []string, height int) *node {
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		// 达到匹配终点了, 若 pattern 为空说明当前节点不是叶子结点, 返回空, 否则返回当前节点.
		if n.pattern == "" {
			return nil
		} else {
			return n
		}
	}

	part := parts[height]            // 当前要匹配的 part
	children := n.mathChildren(part) // part 所对应的所有可能的孩子节点

	for _, child := range children {
		result := child.search(parts, height+1)
		if result != nil {
			return result
		}
	}
	return nil
}
