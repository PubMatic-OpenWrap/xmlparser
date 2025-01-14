package fastxml

import (
	"bytes"
	"fmt"
	"strings"
)

type compare[T any] func(string, T) bool

type treeNode[T any] struct {
	data                     T
	index, first, last, next int
}

func (n treeNode[T]) Data() T {
	return n.data
}

func (n treeNode[T]) Index() int {
	return n.index
}

func (n treeNode[T]) IsLeaf() bool {
	return n.first == -1
}

type tree[T any] struct {
	nodes []treeNode[T] //first node will be always last node
	match compare[T]
}

/*
insert function to insert node n in parent node
NOTE: always re-fetch parent object everytime when inserting new object
*/
func (t *tree[T]) insert(parent *treeNode[T], n treeNode[T]) {
	if len(t.nodes) == 0 {
		t.nodes = append(t.nodes, treeNode[T]{index: 0, first: -1, last: -1, next: -1})
	}

	n.index = len(t.nodes)
	if parent == nil {
		parent = &t.nodes[0]
	}

	if parent.first == -1 {
		//first node
		parent.first = n.index
	} else {
		//subsequent node
		t.nodes[parent.last].next = n.index
	}
	parent.last = n.index

	t.nodes = append(t.nodes, n)
}

func (t *tree[T]) reset() {
	t.nodes = t.nodes[:0]
}

func (t *tree[T]) getChild(parent *treeNode[T], child string) (result *treeNode[T]) {
	parentIndex := 0
	if parent != nil {
		parentIndex = parent.index
	}

	if parentIndex >= len(t.nodes) {
		return nil
	}

	for i := t.nodes[parentIndex].first; i != -1; i = t.nodes[i].next {
		if t.match != nil && t.match(child, t.nodes[i].data) {
			return &t.nodes[i]
		}
	}
	return nil
}

func (t *tree[T]) getAllChild(parent *treeNode[T], child string) (result []*treeNode[T]) {
	parentIndex := 0
	if parent != nil {
		parentIndex = parent.index
	}

	if parentIndex >= len(t.nodes) {
		return nil
	}

	for i := t.nodes[parentIndex].first; i != -1; i = t.nodes[i].next {
		if t.match != nil && t.match(child, t.nodes[i].data) {
			result = append(result, &t.nodes[i])
		}
	}
	return
}

func (t *tree[T]) getChilds(parent *treeNode[T]) (result []*treeNode[T]) {
	parentIndex := 0
	if parent != nil {
		parentIndex = parent.index
	}

	if parentIndex >= len(t.nodes) {
		return nil
	}

	for i := t.nodes[parentIndex].first; i != -1; i = t.nodes[i].next {
		result = append(result, &t.nodes[i])
	}
	return
}

func (t *tree[T]) _getPath(parent int, result *[]*treeNode[T], path ...string) {
	for i := t.nodes[parent].first; i != -1; i = t.nodes[i].next {
		if t.match != nil && t.match(path[0], t.nodes[i].data) {
			if len(path) == 1 {
				(*result) = append((*result), &t.nodes[i])
			} else {
				t._getPath(i, result, path[1:]...)
			}
		}
	}
}

func (t *tree[T]) getPathNodes(parent *treeNode[T], path ...string) (result []*treeNode[T]) {
	parentIndex := 0
	if parent != nil {
		parentIndex = parent.index
	}
	if parentIndex >= len(t.nodes) {
		return nil
	}
	t._getPath(parentIndex, &result, path...)
	return
}

func (t *tree[T]) getPathNode(parent *treeNode[T], path ...string) (result *treeNode[T]) {
	parentIndex := 0
	if parent != nil {
		parentIndex = parent.index
	}

	if parentIndex >= len(t.nodes) {
		return nil
	}

	for iPath := 0; iPath < len(path); iPath++ {
		j := t.nodes[parentIndex].first
		for j != -1 {
			if t.match != nil && t.match(path[iPath], t.nodes[j].data) {
				//found
				break
			}
			j = t.nodes[j].next
		}
		if j == -1 {
			//not found
			return nil
		}
		parentIndex = j
	}

	return &t.nodes[parentIndex]
}

func (t *tree[T]) iterate(f func(*treeNode[T])) {
	for i := range t.nodes {
		f(&t.nodes[i])
	}
}

func (t *tree[T]) _traverse(index int, f func(*treeNode[T])) {
	f(&t.nodes[index])
	for i := t.nodes[index].first; i != -1; i = t.nodes[i].next {
		t._traverse(i, f)
	}
}

func (t *tree[T]) traverse(node *treeNode[T], f func(*treeNode[T])) {
	parent := 0
	if node != nil {
		parent = node.index
	}
	t._traverse(parent, f)
}

/* Printing Function */
func (t *tree[T]) _print(buf *bytes.Buffer, index, indent int, f func(T) string) {
	buf.WriteByte('\n')
	buf.WriteString(strings.Repeat("\t", indent))
	buf.WriteByte('|')
	buf.WriteString(f(t.nodes[index].data))
	indent++
	for i := t.nodes[index].first; i != -1; i = t.nodes[i].next {
		t._print(buf, i, indent, f)
	}
}

func (t *tree[T]) print(f func(T) string) string {
	root := len(t.nodes) - 1
	buf := bytes.Buffer{}
	t._print(&buf, root, 0, f)
	return buf.String()
}

func (t *tree[T]) printRaw(f func(T) string) string {
	buf := bytes.Buffer{}
	for i, node := range t.nodes {
		buf.WriteString(fmt.Sprintf("\n%d:<%d,%d,%d>:%s", i, node.first, node.last, node.next, f(node.data)))
	}
	return buf.String()
}
