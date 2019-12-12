package mls

import (
	"reflect"
	"testing"
)

// Precomputed answers for the tree on ten elements:
//
//                                              X
//                      X
//          X                       X                       X
//    X           X           X           X           X
// X     X     X     X     X     X     X     X     X     X     X
// 0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f 10 11 12 13 14
var (
	aRoot = []nodeIndex{0x00, 0x01, 0x03, 0x03, 0x07, 0x07, 0x07, 0x07, 0x0f, 0x0f, 0x0f}

	aN       = leafCount(0x0b)
	index    = []nodeIndex{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14}
	aLog2    = []nodeIndex{0x00, 0x00, 0x01, 0x01, 0x02, 0x02, 0x02, 0x02, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x04, 0x04, 0x04, 0x04, 0x04}
	aLevel   = []nodeIndex{0x00, 0x01, 0x00, 0x02, 0x00, 0x01, 0x00, 0x03, 0x00, 0x01, 0x00, 0x02, 0x00, 0x01, 0x00, 0x04, 0x00, 0x01, 0x00, 0x02, 0x00}
	aLeft    = []nodeIndex{0x00, 0x00, 0x02, 0x01, 0x04, 0x04, 0x06, 0x03, 0x08, 0x08, 0x0a, 0x09, 0x0c, 0x0c, 0x0e, 0x07, 0x10, 0x10, 0x12, 0x11, 0x14}
	aRight   = []nodeIndex{0x00, 0x02, 0x02, 0x05, 0x04, 0x06, 0x06, 0x0b, 0x08, 0x0a, 0x0a, 0x0d, 0x0c, 0x0e, 0x0e, 0x13, 0x10, 0x12, 0x12, 0x14, 0x14}
	aParent  = []nodeIndex{0x01, 0x03, 0x01, 0x07, 0x05, 0x03, 0x05, 0x0f, 0x09, 0x0b, 0x09, 0x07, 0x0d, 0x0b, 0x0d, 0x0f, 0x11, 0x13, 0x11, 0x0f, 0x13}
	aSibling = []nodeIndex{0x02, 0x05, 0x00, 0x0b, 0x06, 0x01, 0x04, 0x13, 0x0a, 0x0d, 0x08, 0x03, 0x0e, 0x09, 0x0c, 0x0f, 0x12, 0x14, 0x10, 0x07, 0x11}

	aDirpath = [][]nodeIndex{
		{0x00, 0x01, 0x03, 0x07, 0x0f},
		{0x01, 0x03, 0x07, 0x0f},
		{0x02, 0x01, 0x03, 0x07, 0x0f},
		{0x03, 0x07, 0x0f},
		{0x04, 0x05, 0x03, 0x07, 0x0f},
		{0x05, 0x03, 0x07, 0x0f},
		{0x06, 0x05, 0x03, 0x07, 0x0f},
		{0x07, 0x0f},
		{0x08, 0x09, 0x0b, 0x07, 0x0f},
		{0x09, 0x0b, 0x07, 0x0f},
		{0x0a, 0x09, 0x0b, 0x07, 0x0f},
		{0x0b, 0x07, 0x0f},
		{0x0c, 0x0d, 0x0b, 0x07, 0x0f},
		{0x0d, 0x0b, 0x07, 0x0f},
		{0x0e, 0x0d, 0x0b, 0x07, 0x0f},
		{0x0f},
		{0x10, 0x11, 0x13, 0x0f},
		{0x11, 0x13, 0x0f},
		{0x12, 0x11, 0x13, 0x0f},
		{0x13, 0x0f},
		{0x14, 0x13, 0x0f},
	}
	aCopath = [][]nodeIndex{
		{0x02, 0x05, 0x0b, 0x13},
		{0x05, 0x0b, 0x13},
		{0x00, 0x05, 0x0b, 0x13},
		{0x0b, 0x13},
		{0x06, 0x01, 0x0b, 0x13},
		{0x01, 0x0b, 0x13},
		{0x04, 0x01, 0x0b, 0x13},
		{0x13},
		{0x0a, 0x0d, 0x03, 0x13},
		{0x0d, 0x03, 0x13},
		{0x08, 0x0d, 0x03, 0x13},
		{0x03, 0x13},
		{0x0e, 0x09, 0x03, 0x13},
		{0x09, 0x03, 0x13},
		{0x0c, 0x09, 0x03, 0x13},
		{},
		{0x12, 0x14, 0x07},
		{0x14, 0x07},
		{0x10, 0x14, 0x07},
		{0x07},
		{0x11, 0x07},
	}
)

func TestSizeProperties(t *testing.T) {
	for n := leafCount(1); n < aN; n += 1 {
		if root(n) != aRoot[n-1] {
			t.Fatalf("Root mismatch: %v != %v", root(n), aRoot[n-1])
		}
	}
}

func TestNodeRelations(t *testing.T) {
	run := func(label string, f func(x nodeIndex) nodeIndex, a []nodeIndex) {
		for i, x := range index {
			if f(x) != a[i] {
				t.Fatalf("Relation test failure: %s @ 0x%02x: %v != %v", label, x, f(x), a[i])
			}
		}
	}

	run("log2", func(x nodeIndex) nodeIndex { return nodeIndex(log2(nodeCount(x))) }, aLog2)
	run("level", func(x nodeIndex) nodeIndex { return nodeIndex(level(x)) }, aLevel)
	run("left", left, aLeft)
	run("right", func(x nodeIndex) nodeIndex { return right(x, aN) }, aRight)
	run("parent", func(x nodeIndex) nodeIndex { return parent(x, aN) }, aParent)
	run("sibling", func(x nodeIndex) nodeIndex { return sibling(x, aN) }, aSibling)
}

func TestPaths(t *testing.T) {
	run := func(label string, f func(x nodeIndex, n leafCount) []nodeIndex, a [][]nodeIndex) {
		for i, x := range index {
			if !reflect.DeepEqual(f(x, aN), a[i]) {
				t.Fatalf("Path test failure: %s @ 0x%02x: %v != %v", label, x, f(x, aN), a[i])
			}
		}
	}

	run("dirpath", dirpath, aDirpath)
	run("copath", copath, aCopath)
}
