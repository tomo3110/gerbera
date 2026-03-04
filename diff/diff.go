package diff

import (
	"sort"
	"strings"

	"github.com/tomo3110/gerbera"
)

// OpType represents the type of a DOM patch operation.
type OpType string

const (
	OpSetText    OpType = "text"    // change text content (Value)
	OpSetHTML    OpType = "html"   // set innerHTML (raw HTML content)
	OpSetAttr    OpType = "attr"    // add/change attribute
	OpRemoveAttr OpType = "rattr"  // remove attribute
	OpSetClass   OpType = "class"  // change class attribute
	OpInsert     OpType = "insert" // insert child node
	OpRemove     OpType = "remove" // remove child node
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

// Diff compares two Element trees and returns the list of patches
// needed to transform oldEl into newEl.
func Diff(oldEl, newEl *gerbera.Element) []Patch {
	var patches []Patch
	diffRecursive(oldEl, newEl, nil, &patches)
	return patches
}

func diffRecursive(oldEl, newEl *gerbera.Element, path []int, patches *[]Patch) {
	// Different tag → replace the entire subtree
	if oldEl.TagName != newEl.TagName {
		html, _ := RenderFragment(newEl)
		*patches = append(*patches, Patch{
			Op:   OpReplace,
			Path: copyPath(path),
			HTML: html,
		})
		return
	}

	// Same tag but different keys → replace entire subtree
	if (oldEl.Key != "" || newEl.Key != "") && oldEl.Key != newEl.Key {
		html, _ := RenderFragment(newEl)
		*patches = append(*patches, Patch{
			Op:   OpReplace,
			Path: copyPath(path),
			HTML: html,
		})
		return
	}

	// Compare Value (text content)
	if oldEl.Value != newEl.Value {
		op := OpSetText
		if strings.Contains(newEl.Value, "<") {
			op = OpSetHTML
		}
		*patches = append(*patches, Patch{
			Op:    op,
			Path:  copyPath(path),
			Value: newEl.Value,
		})
	}

	// Compare Attr maps
	diffAttrs(oldEl.Attr, newEl.Attr, path, patches)

	// Compare ClassNames
	diffClasses(oldEl.ClassNames, newEl.ClassNames, path, patches)

	// Compare Children
	diffChildren(oldEl.Children, newEl.Children, path, patches)
}

func diffAttrs(oldAttr, newAttr gerbera.AttrMap, path []int, patches *[]Patch) {
	// Attributes added or changed
	for key, newVal := range newAttr {
		if oldVal, ok := oldAttr[key]; !ok || oldVal != newVal {
			*patches = append(*patches, Patch{
				Op:    OpSetAttr,
				Path:  copyPath(path),
				Key:   key,
				Value: newVal,
			})
		}
	}
	// Attributes removed
	for key := range oldAttr {
		if _, ok := newAttr[key]; !ok {
			*patches = append(*patches, Patch{
				Op:   OpRemoveAttr,
				Path: copyPath(path),
				Key:  key,
			})
		}
	}
}

func diffClasses(oldClasses, newClasses gerbera.ClassMap, path []int, patches *[]Patch) {
	oldStr := classString(oldClasses)
	newStr := classString(newClasses)
	if oldStr != newStr {
		*patches = append(*patches, Patch{
			Op:    OpSetClass,
			Path:  copyPath(path),
			Value: newStr,
		})
	}
}

func classString(cm gerbera.ClassMap) string {
	if len(cm) == 0 {
		return ""
	}
	names := make([]string, 0, len(cm))
	for name := range cm {
		names = append(names, name)
	}
	sort.Strings(names)
	return strings.Join(names, " ")
}

func diffChildren(oldChildren, newChildren []*gerbera.Element, path []int, patches *[]Patch) {
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
