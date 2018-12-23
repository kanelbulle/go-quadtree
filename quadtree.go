// Quadtree provides a data structure allowing efficient spatial queries on 2D points.
package quadtree

import (
	"fmt"
)

// Quadtree is a dynamically resizable struct that offers
// performant access to items that can be assigned a 2d position. Internally,
// it represents the data with a tree by segmenting each level into four quadrants.
// The Quadtree has two attributes that may be tweaked depending on your use case:
// Maximum depth: To get the spatial resolution you want, the number of levels
// can be adjusted. For a depth D, the space will be subdivided 2^D times. The
// default is 10.
// The coordinate system of a Quadtree has origo at the lower left corner, with X
// and Y growing positively to the upper right corner. X is the horizontal axis,
// Y is the vertical axis.
//
// Quadtree is not thread-safe.
type Quadtree struct {
	maxDepth        int
	maxItemsPerNode int
	root            *node
	// The total number of items in this Quadtree.
	size            int
	debugAssertions bool
}

type Point struct {
	X float64
	Y float64
}

type Rect struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

const quadrantNone = -1
const quadrantUpperLeft = 0
const quadrantUpperRight = 1
const quadrantLowerLeft = 2
const quadrantLowerRight = 3

// A single node in the tree.
type node struct {
	// The bounds that define this node
	bounds Rect
	// The depth this node is at. Root node is at depth 0.
	depth int
	// The entries that are inside this node
	items []treeEntry
	// The four child nodes of this node (one node per quadrant).
	ul, ur, ll, lr *node
}

// An individual entry in the tree.
type treeEntry struct {
	position Point
	data     interface{}
}

type consumer func(treeEntry) bool

func NewQuadtree(bounds Rect, maxDepth, maxItemsPerNode int) (*Quadtree, error) {
	if maxDepth <= 0 {
		return nil, fmt.Errorf("Creating tree failed: maxDepth must be larger than 1")
	}
	if maxItemsPerNode <= 0 {
		return nil, fmt.Errorf("Creating tree failed: maxItemsPerNode must be larger than 0")
	}
	return &Quadtree{
		maxDepth:        maxDepth,
		maxItemsPerNode: maxItemsPerNode,
		root: &node{
			bounds: bounds,
			depth:  0,
			items:  make([]treeEntry, 0, 4),
		},
	}, nil
}

func (qt *Quadtree) Size() int {
	return qt.size
}

// Returns the objects within the given bounds
func (qt *Quadtree) Query(bounds Rect) []interface{} {
	items := make([]interface{}, 0, 10)
	queryInternal(qt.root, bounds, func(item treeEntry) bool {
		items = append(items, item.data)
		return true
	})
	return items
}

// This will recurse down the tree, removing the nodes that
// have no overlap with the given bounds. When all overlapping
// nodes are found, their items are returned.
func queryInternal(node *node, bounds Rect, consumer consumer) {
	if overlaps(node.bounds, bounds) {
		if node.items == nil {
			// This node has no items, but it has children. Keep recursing.
			queryInternal(node.ul, bounds, consumer)
			queryInternal(node.ur, bounds, consumer)
			queryInternal(node.ll, bounds, consumer)
			queryInternal(node.lr, bounds, consumer)
		} else {
			// We reached an end node. Since this node may only be partially
			// overlapping, ensure each item is inside bounds before consuming.
			for _, e := range node.items {
				if bounds.Contains(e.position) {
					// TODO: handle the return value and exit
					consumer(e)
				}
			}
		}
	}
}

// Adds the data to the tree with the given position.
func (qt *Quadtree) Add(data interface{}, position Point) (err error) {
	if !qt.root.bounds.Contains(position) {
		// If the root node can't contain this data, signal error.
		return fmt.Errorf("Add failed: position outside bounds of tree.")
	}
	item := treeEntry{position, data}
	addInternal(qt, qt.root, item)
	qt.size += 1
	return err
}

// Recurses down the tree to find the correct node for the item.
// 1) Finds an empty node and adds the data to that node
// 2) Finds a node that need to be split into child nodes
// 3)
func addInternal(qt *Quadtree, node *node, item treeEntry) {
	quadrant := whichQuadrant(node.bounds, item.position)
	if quadrant == quadrantNone {
		// The position doesn't belong to this node at all - exit.
		return
	}

	if node.items != nil && node.depth >= qt.maxDepth {
		// We've reached the max depth of the tree. The item must be stored
		// inside this node, regardless of maxItemsPerNode.
		node.items = append(node.items, item)
	} else if len(node.items) >= qt.maxItemsPerNode {
		// This node is already at max capacity, so we need to split it into
		// child nodes.
		ul, ur, ll, lr := node.bounds.quadrants()
		node.ul = newNode(node, ul)
		node.ur = newNode(node, ur)
		node.ll = newNode(node, ll)
		node.lr = newNode(node, lr)
		items := node.items
		node.items = nil
		for _, i := range items {
			addInternal(qt, node, i)
		}
		addInternal(qt, node, item)
	} else if node.items != nil {
		// This node still has an items array which means it does not have
		// any child nodes. Just append to this node.
		node.items = append(node.items, item)
	} else {
		// Recurse down the tree to find the correct node.
		switch quadrant {
		case quadrantUpperLeft:
			addInternal(qt, node.ul, item)
		case quadrantUpperRight:
			addInternal(qt, node.ur, item)
		case quadrantLowerLeft:
			addInternal(qt, node.ll, item)
		case quadrantLowerRight:
			addInternal(qt, node.lr, item)
		}
	}
}

func newNode(parent *node, bounds Rect) *node {
	return &node{
		bounds: bounds,
		depth:  parent.depth + 1,
		items:  make([]treeEntry, 0, 4),
	}
}

// Returns which quadrant the Point p is inside Rect r
func whichQuadrant(r Rect, p Point) int {
	if !r.Contains(p) {
		return quadrantNone
	}
	midX := r.X + r.Width/2.0
	midY := r.Y + r.Height/2.0
	if p.X < midX {
		// Left half
		if p.Y < midY {
			return quadrantLowerLeft
		} else {
			return quadrantUpperLeft
		}
	} else {
		// Right half
		if p.Y < midY {
			return quadrantLowerRight
		} else {
			return quadrantUpperRight
		}
	}
}

// Returns whether two rectangles overlap.
func overlaps(r1, r2 Rect) bool {
	return r1.X < r2.X+r2.Width &&
		r1.X+r1.Width > r2.X &&
		r1.Y+r1.Height > r2.Y &&
		r1.Y < r2.Y+r2.Height
}

// Splits the given rect into quadrants
func (r Rect) quadrants() (ul, ur, ll, lr Rect) {
	w := r.Width / 2.0
	h := r.Height / 2.0
	ll = Rect{r.X, r.Y, w, h}
	ul = Rect{r.X, r.Y + h, w, h}
	ur = Rect{r.X + w, r.Y + h, w, h}
	lr = Rect{r.X + w, r.Y, w, h}
	return
}

// Returns if this rectangle contains the given point.
func (r *Rect) Contains(point Point) bool {
	return point.X >= r.X &&
		point.Y >= r.Y &&
		point.X < r.X+r.Width &&
		point.Y < r.Y+r.Height
}
