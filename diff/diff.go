package diff

import (
	"strings"

	"github.com/tomo3110/gerbera"
)

// OpType represents the type of a DOM patch operation.
type OpType string

const (
	OpSetText    OpType = "text"    // change text content (Value)
	OpSetHTML    OpType = "html"    // set innerHTML (raw HTML content)
	OpSetAttr    OpType = "attr"    // add/change attribute
	OpRemoveAttr OpType = "rattr"   // remove attribute
	OpSetClass   OpType = "class"   // change class attribute
	OpInsert     OpType = "insert"  // insert child node
	OpRemove     OpType = "remove"  // remove child node
	OpReplace    OpType = "replace" // replace entire node
)

// Patch represents a single DOM mutation.
type Patch struct {
	Op    OpType `json:"op"`
	Path  []int  `json:"path"`
	Key   string `json:"key,omitempty"`
	Value string `json:"val,omitempty"`
	Index int    `json:"idx,omitempty"`
	HTML  string `json:"html,omitempty"`
}

// Diff compares two Node trees and returns the list of patches
// needed to transform oldNode into newNode.
// Operates on the Node interface, making it platform-independent.
func Diff(oldNode, newNode gerbera.Node) []Patch {
	var patches []Patch
	diffRecursive(oldNode, newNode, nil, &patches)
	return patches
}

func diffRecursive(oldNode, newNode gerbera.Node, path []int, patches *[]Patch) {
	// Different tag → replace the entire subtree
	if oldNode.Tag() != newNode.Tag() {
		html, _ := RenderFragment(newNode)
		*patches = append(*patches, Patch{
			Op:   OpReplace,
			Path: copyPath(path),
			HTML: html,
		})
		return
	}

	// Same tag but different keys → replace entire subtree
	if (oldNode.NodeKey() != "" || newNode.NodeKey() != "") && oldNode.NodeKey() != newNode.NodeKey() {
		html, _ := RenderFragment(newNode)
		*patches = append(*patches, Patch{
			Op:   OpReplace,
			Path: copyPath(path),
			HTML: html,
		})
		return
	}

	// Compare text content
	if oldNode.Text() != newNode.Text() {
		op := OpSetText
		if strings.Contains(newNode.Text(), "<") {
			op = OpSetHTML
		}
		*patches = append(*patches, Patch{
			Op:    op,
			Path:  copyPath(path),
			Value: newNode.Text(),
		})
	}

	// Compare attributes (includes class)
	diffAttributes(oldNode.Attributes(), newNode.Attributes(), path, patches)

	// Compare children
	diffChildren(oldNode.Children(), newNode.Children(), path, patches)
}

func diffAttributes(oldAttrs, newAttrs []gerbera.Attribute, path []int, patches *[]Patch) {
	oldMap := attrMap(oldAttrs)
	newMap := attrMap(newAttrs)

	// Attributes added or changed
	for _, a := range newAttrs {
		if oldVal, ok := oldMap[a.Key]; !ok || oldVal != a.Value {
			if a.Key == "class" {
				*patches = append(*patches, Patch{
					Op:    OpSetClass,
					Path:  copyPath(path),
					Value: a.Value,
				})
			} else {
				*patches = append(*patches, Patch{
					Op:    OpSetAttr,
					Path:  copyPath(path),
					Key:   a.Key,
					Value: a.Value,
				})
			}
		}
	}

	// Attributes removed
	for _, a := range oldAttrs {
		if _, ok := newMap[a.Key]; !ok {
			if a.Key == "class" {
				*patches = append(*patches, Patch{
					Op:    OpSetClass,
					Path:  copyPath(path),
					Value: "",
				})
			} else {
				*patches = append(*patches, Patch{
					Op:   OpRemoveAttr,
					Path: copyPath(path),
					Key:  a.Key,
				})
			}
		}
	}
}

func attrMap(attrs []gerbera.Attribute) map[string]string {
	m := make(map[string]string, len(attrs))
	for _, a := range attrs {
		m[a.Key] = a.Value
	}
	return m
}

func diffChildren(oldChildren, newChildren []gerbera.Node, path []int, patches *[]Patch) {
	minLen := len(oldChildren)
	if len(newChildren) < minLen {
		minLen = len(newChildren)
	}

	// Recurse into common children
	for i := 0; i < minLen; i++ {
		childPath := append(copyPath(path), i)
		diffRecursive(oldChildren[i], newChildren[i], childPath, patches)
	}

	// New children added
	for i := minLen; i < len(newChildren); i++ {
		html, _ := RenderFragment(newChildren[i])
		*patches = append(*patches, Patch{
			Op:    OpInsert,
			Path:  copyPath(path),
			Index: i,
			HTML:  html,
		})
	}

	// Old children removed (reverse order to preserve indices)
	for i := len(oldChildren) - 1; i >= minLen; i-- {
		*patches = append(*patches, Patch{
			Op:    OpRemove,
			Path:  copyPath(path),
			Index: i,
		})
	}
}

func copyPath(path []int) []int {
	if path == nil {
		return []int{}
	}
	cp := make([]int, len(path))
	copy(cp, path)
	return cp
}
