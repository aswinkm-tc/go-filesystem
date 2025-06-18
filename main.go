package main

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"time"
)

var (
	// ErrActionNotAllowed is returned when an action is not allowed on the INode.
	ErrActionNotAllowed = fmt.Errorf("action not allowed on this INode")
)

// INode is an interface that defines the methods for an INode in a filesystem-like structure.
type INode interface {
	AddChild(child INode)
	PrintInfo(args ...string)
	Walk(fn func(INode))
	Find(path string) (INode, error)
}

// iNode is a struct that implements the INode interface.
type iNode struct {
	name             string
	parent           INode
	children         []INode
	creationTime     int64
	modificationTime int64
	isDirectory      bool
	data             []byte
	deleted          bool
}

// Create the root INode with no parent and no children.
func Root() INode {
	return &iNode{
		name:             "/",
		parent:           nil,
		children:         make([]INode, 0),
		creationTime:     time.Now().Unix(),
		modificationTime: time.Now().Unix(),
		isDirectory:      true,
		data:             nil,
		deleted:          false,
	}
}

// AddChild adds a child INode to the current INode.
func (n *iNode) AddChild(child INode) {
	if childNode, ok := child.(*iNode); ok {
		childNode.parent = n
		n.children = append(n.children, childNode)
	}
}

// CreateDirectory creates a new directory INode with the given name and parent.
func CreateDirectory(name string, parent INode) INode {
	n := &iNode{
		name:             name,
		parent:           parent,
		children:         make([]INode, 0),
		creationTime:     time.Now().Unix(),
		modificationTime: time.Now().Unix(),
		isDirectory:      true,
		data:             nil,
	}
	parent.AddChild(n)
	return n
}

// CreateFile creates a new file INode with the given name, parent, and data.
func CreateFile(name string, parent INode, data []byte) INode {
	n := &iNode{
		name:             name,
		parent:           parent,
		children:         nil,
		creationTime:     time.Now().Unix(),
		modificationTime: time.Now().Unix(),
		data:             data,
	}
	parent.AddChild(n)
	return n
}

// Delete removes the INode from its parent's children.
func (n *iNode) Delete(args ...string) error {
	softDelete := !slices.Contains(args, "--force") || !slices.Contains(args, "-f")
	if n.parent == nil {
		return ErrActionNotAllowed
	}

	if softDelete {
		n.deleted = true
		n.modificationTime = time.Now().Unix()
		return nil
	}

	if parentNode, ok := n.parent.(*iNode); ok {
		for i, child := range parentNode.children {
			if child == n {
				parentNode.children = append(parentNode.children[:i], parentNode.children[i+1:]...)
				break
			}
		}
		n.parent = nil
	}
	return nil
}

// Walk traverses the INode and its children, applying the given function to each INode.
func (n *iNode) Walk(fn func(INode)) {
	fn(n)
	for _, child := range n.children {
		child.Walk(fn)
	}
}

// PrintInfo prints the information of the INode in ls -l format.
func (n *iNode) PrintInfo(args ...string) {
	if n.isDirectory {
		fmt.Println("d", n.name, time.Unix(n.creationTime, 0).Format(time.RFC3339), time.Unix(n.modificationTime, 0).Format(time.RFC3339))
	} else {
		fmt.Println("-", n.name, time.Unix(n.creationTime, 0).Format(time.RFC3339), time.Unix(n.modificationTime, 0).Format(time.RFC3339), len(n.data), "bytes")
	}
}

// Find finds an INode by name in the current INode and its children.
func (n *iNode) Find(path string) (INode, error) {
	if path == "" || path == "/" {
		return n, nil
	}

	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")

	node := n.findRecursive(parts)
	if node == nil {
		return nil, os.ErrNotExist
	}
	return node, nil
}

// findRecursive is a helper function to recursively find an INode by its path.
func (n *iNode) findRecursive(parts []string) INode {
	if len(parts) == 0 {
		return n
	}

	for _, child := range n.children {
		if childNode, ok := child.(*iNode); ok && childNode.name == parts[0] {
			return childNode.findRecursive(parts[1:])
		}
	}
	return nil
}

func main() {
	root := Root()
	dir1 := CreateDirectory("dir1", root)
	file1 := CreateFile("file1.txt", dir1, []byte("Hello, World!"))
	dir2 := CreateDirectory("dir2", root)
	file2 := CreateFile("file2.txt", dir2, []byte("Another file content."))

	root.Walk(func(n INode) {
		n.PrintInfo()
	})
	root.PrintInfo()
	file1.PrintInfo()
	file2.PrintInfo()
	fmt.Println("Finding dir1/file1.txt:")
	node, err := root.Find("dir1/file1.txt")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		node.PrintInfo()
	}
	fmt.Println("Finding file2.txt in dir2:")
	node, err = dir2.Find("file2.txt")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		node.PrintInfo()
	}
}
