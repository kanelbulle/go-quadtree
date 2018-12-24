# goquadtree
A [Quadtree](https://en.wikipedia.org/wiki/Quadtree) implementation in go. This datastructure allows efficient spatial queries of objects associated with a 2D coordinate.

You can create one like this:

```go
import (
    quadtree "github.com/kanelbulle/goquadtree"
)

func main() {
    qt, err := quadtree.NewQuadtree(quadtree.Rect{0, 0, 1, 1}, 10, 10)
    if err != nil {
        panic("I did bad!")
    }

    // Add a node to the tree
    qt.Add(7, quadtree.Point{0.3, 0.14})

    // Find all the items within the given rectangle
    items := qt.Query(quadtree.Rect{0.1, 0.1, 0.5, 0.5})
}
```

This implementation has two tweakable parameters:
* Tree depth (`maxDepth`)
* Items per node before splitting (`maxItemsPerNode`)

Use `maxDepth` to control the depth of the tree. To choose a value, it may help to see that the bounds you specify for the tree (say it's a Rect of width and height 10), will be subdivided `2^maxDepth` times. I.e, for a Rect of size 10, the smallest segment in that tree will be a rectangle of size `10/2^maxDepth`. For most uses, this can be left at 10.

Use `maxItemsPerNode` to control how many items can be stored in a single node before it must be split into child nodes. This attribute lets you trade off query speed for a sparser tree. If memory is what you optimize for, this attribute can be made larger. However, keep in mind that the larger this value is, the performance of the tree will approach `O(N)`.

This Quadtree implementation uses the cartesian coordinate system.
This Quadtree implementation is not thread safe.

## To run the tests

You will need one dependency ([testify](github.com/stretchr/testify/assert)) to run the tests.

```bash
go test -v github.com/kanelbulle/goquadtree
```

You can also run the benchmarks:

```bash
go test -v -bench=. github.com/kanelbulle/goquadtree
```
