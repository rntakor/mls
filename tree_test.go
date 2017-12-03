package mls

import (
	"reflect"
	"testing"
)

var stringNodeDefn = &nodeDefinition{
	valid: func(x Node) bool {
		_, ok := x.(string)
		return ok
	},

	equal: func(x, y Node) bool {
		xs, okx := x.(string)
		ys, oky := y.(string)
		return okx && oky && (xs == ys)
	},

	create: func(d []byte) Node {
		return string(d)
	},

	combine: func(x, y Node) ([]byte, error) {
		xs, okx := x.(string)
		ys, oky := y.(string)
		if !okx || !oky {
			return nil, InvalidNodeError
		}

		return []byte(xs + ys), nil
	},
}

func TestNewTree(t *testing.T) {
	aDefn := stringNodeDefn
	aSize := uint(3)
	aNodes := map[uint]Node{
		0: "a",
		1: "ab",
		2: "b",
		3: "abc",
		4: "c",
	}

	leaves := []Node{"a", "b", "c"}
	tree, err := newTreeFromLeaves(stringNodeDefn, leaves)

	if err != nil {
		t.Fatalf("Error constructing tree: %v", err)
	}

	if tree.defn != aDefn {
		t.Fatalf("Incorrect tree node definition: %v != %v", tree.defn, aDefn)
	}

	if tree.size != aSize {
		t.Fatalf("Incorrect computed tree size: %v != %v", tree.size, aSize)
	}

	if !reflect.DeepEqual(tree.nodes, aNodes) {
		t.Fatalf("Incorrect computed tree nodes: %v != %v", tree.nodes, aNodes)
	}

	// Test equality
	if !tree.Equal(tree) {
		t.Fatalf("Tree does not equal itself")
	}
}

func TestTreeAdd(t *testing.T) {
	aDefn := stringNodeDefn
	aSize := uint(5)
	aLeaves := []Node{"a", "b", "c", "d", "e"}
	aNodes := map[uint]Node{
		0: "a",
		1: "ab",
		2: "b",
		3: "abcd",
		4: "c",
		5: "cd",
		6: "d",
		7: "abcde",
		8: "e",
	}
	aFrontier := &Frontier{
		Entries: []FrontierEntry{
			{Value: "abcd", Size: 4},
			{Value: "e", Size: 1},
		},
	}

	// Build tree by additions
	tree := newTree(stringNodeDefn)
	for _, leaf := range aLeaves {
		if err := tree.Add(leaf); err != nil {
			t.Fatalf("Error adding leaf: %v", err)
		}
	}

	// Verify contents directly
	if tree.size != aSize {
		t.Fatalf("Incorrect computed tree size: %v != %v", tree.size, aSize)
	}

	if !reflect.DeepEqual(tree.nodes, aNodes) {
		t.Fatalf("Incorrect computed tree nodes: %v != %v", tree.nodes, aNodes)
	}

	// Verify that it's the same as a tree built directly
	aTree, _ := newTreeFromLeaves(aDefn, aLeaves)
	if !aTree.Equal(tree) {
		t.Fatalf("Add-built tree does not equal leaf-built tree: %v != %v", aTree, tree)
	}

	// Verify that it has all its leaves
	if !tree.HasAllLeaves() {
		t.Fatalf("Add-built tree does not have all leaves: %v", tree)
	}

	// Verify that its leaves are as expected
	if !reflect.DeepEqual(tree.Leaves(), aLeaves) {
		t.Fatalf("Add-built tree does not expected leaves: %v != %v", tree.Leaves(), aLeaves)
	}

	// Verify that the Frontier is as expected
	if !reflect.DeepEqual(tree.Frontier(), aFrontier) {
		t.Fatalf("Add-built tree does not expected frontier: %v != %v", tree.Frontier(), aFrontier)
	}

	// Verify that Copaths have expected values
	for i := uint(0); i < tree.size; i += 1 {
		c := copath(2*i, tree.size)
		C := tree.Copath(i)

		if C.Index != i {
			t.Fatalf("Copath has wrong index @ %v: %v != %v", i, C.Index, i)
		}

		if C.Size != tree.size {
			t.Fatalf("Copath has wrong size @ %v: %v != %v", i, C.Size, tree.size)
		}

		if len(C.Nodes) != len(c) {
			t.Fatalf("Copath has wrong path length @ %v: %v != %v", i, len(C.Nodes), len(c))
		}
	}
}

func TestTreeUpdate(t *testing.T) {
	aSize := uint(5)
	aLeaves := []Node{"a", "b", "c", "d", "e"}

	aIndex1 := uint(3)
	aNewLeaf1 := "x"
	aNodes1 := map[uint]Node{
		0: "a",
		1: "ab",
		2: "b",
		3: "abcx",
		4: "c",
		5: "cx",
		6: "x",
		7: "abcxe",
		8: "e",
	}

	aIndex2 := uint(1)
	aUpdatePath2 := []Node{"y", "ay", "aycx"}
	aNodes2 := map[uint]Node{
		0: "a",
		1: "ay",
		2: "y",
		3: "aycx",
		4: "c",
		5: "cx",
		6: "x",
		7: "aycxe",
		8: "e",
	}

	// Build tree, then update leaf
	tree, _ := newTreeFromLeaves(stringNodeDefn, aLeaves)
	if err := tree.Update(aIndex1, aNewLeaf1); err != nil {
		t.Fatalf("Error updating leaf: %v", err)
	}

	if tree.size != aSize {
		t.Fatalf("Incorrect computed tree size: %v != %v", tree.size, aSize)
	}

	if !reflect.DeepEqual(tree.nodes, aNodes1) {
		t.Fatalf("Incorrect computed tree nodes: %v != %v", tree.nodes, aNodes1)
	}

	// Update another leaf with a full path
	if err := tree.UpdateWithPath(aIndex2, aUpdatePath2); err != nil {
		t.Fatalf("Error updating leaf: %v", err)
	}

	if tree.size != aSize {
		t.Fatalf("Incorrect computed tree size: %v != %v", tree.size, aSize)
	}

	if !reflect.DeepEqual(tree.nodes, aNodes2) {
		t.Fatalf("Incorrect computed tree nodes: %v != %v", tree.nodes, aNodes2)
	}
}

func TestTreeUpdatePath(t *testing.T) {
	aLeaves := []Node{"a", "b", "c", "d", "e"}
	aIndex := uint(3)
	aNewLeaf := "x"
	aPath := []Node{"abcx", "cx", "x"}

	// Build tree, then generate update path
	tree, _ := newTreeFromLeaves(stringNodeDefn, aLeaves)

	path, err := tree.UpdatePath(aIndex, aNewLeaf)
	if err != nil {
		t.Fatalf("Error creating update path: %v", err)
	}

	if !reflect.DeepEqual(path, aPath) {
		t.Fatalf("Incorrect computed update path: %v != %v", path, aPath)
	}
}
