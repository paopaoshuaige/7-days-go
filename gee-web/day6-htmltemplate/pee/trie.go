package pee

import "strings"

type node struct {
	pattern  string  // 待匹配路由，例如 /p/:lang
	part     string  // 当前节点所占的路由一部分，比如:lang
	children []*node // 当前节点的子节点，例如 [doc, tutorial, intro]
	isWild   bool    // 是否精确匹配，part 含有 : 或 * 时为true，去匹配的时候如果是true不相等也可以匹配
}

// 插入
func (n *node) insert(pattern string, parts []string, height int) {
	// 如果parts里面的后缀已经都插入了让他等于最后递归的后缀
	if len(parts) == height {
		// 比如/hello/:name，到2==2的时候hello.pattern就等于:name了，也就相当于:name是hello的子路由
		n.pattern = pattern
		return
	}

	// 获取第路径名
	part := parts[height]
	// 匹配节点
	child := n.matchChild(part)
	// 如果匹配不到节点就创建一个
	if child == nil {
		// 保存该节点的路径信息，如果路径信息里有：或者*就设置精确匹配
		child = &node{
			part:   part,
			isWild: part[0] == ':' || part[0] == '*',
		}
		// 把刚才创建的节点加入当前节点的子节点
		n.children = append(n.children, child)
	}
	child.insert(pattern, parts, height+1)
}

func (n *node) search(parts []string, height int) *node {
	// 如果搜完了或者检测到*是前缀
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		if n.pattern == "" {
			return nil
		}
		return n // 把上一个节点返回，因为是通过循环递归进来的，当前的函数执行就是上一个n
		// 比如/hello/lzj，现在判断的就是2 == 2，然后n.pattern就是hello.pattern
		// 如果没有待匹配路径，比如GET（根）就没有，只有一个part hello
	}

	part := parts[height]
	children := n.matchChildren(part) // 获取所有和该url取出来的待匹配路由匹配的

	for _, child := range children { // 通过每个匹配到的路由往下搜
		result := child.search(parts, height+1) // 直到搜到合适的路由节点
		if result != nil {
			return result
		}
	}
	return nil
}

// 第一个匹配成功的节点，用于插入
func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

// 所有匹配成功的，用来查找
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild { // 如果是：name，就直接匹配
			nodes = append(nodes, child)
		}
	}
	return nodes
}
