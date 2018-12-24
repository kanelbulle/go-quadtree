package quadtree

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

type TestData struct {
	id int
	position Point
}

func NewData(id int) *TestData {
	return &TestData{id, Point{0, 0}}
}

func NewDataWithPosition(id int, position Point) *TestData {
	return &TestData{id, position}
}

func ToData(d interface{}) *TestData {
	data, ok := d.(*TestData)
	if (!ok) {
		panic("it's not a TestData")
	}
 	return data
}

func BenchmarkAdd(b *testing.B) {
	bounds := Rect{0, 0, 1, 1}
	qt, _ := NewQuadtree(bounds, 10, 10)

	data := make([]*TestData, b.N, b.N)
	positions := make([]Point, b.N, b.N)
	for i := 0; i < b.N; i++ {
		data[i] = NewData(i)
		positions[i] = Point{rand.Float64(), rand.Float64()}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qt.Add(data[i], positions[i])
	}
}

func BenchmarkQuery(b *testing.B) {
	qt, _ := NewQuadtree(Rect{0, 0, 1, 1}, 10, 10)

	rectSize := 0.02
	numItems := 100000
	for i := 0; i < numItems; i++ {
		qt.Add(NewData(i), Point{rand.Float64(), rand.Float64()})
	}

	bounds := make([]Rect, b.N, b.N)
	for i := 0; i < b.N; i++ {
		bounds[i] = Rect{rand.Float64(), rand.Float64(),
			rand.Float64() * rectSize, rand.Float64() * rectSize}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qt.Query(bounds[i])
	}
}

func TestQueryIterative_returnsCorrectPosition(t *testing.T) {
	qt, _ := NewQuadtree(Rect{-1, -1, 1, 1}, 2, 2)

	d1 := NewDataWithPosition(1, Point{-0.5, -0.1})
	d2 := NewDataWithPosition(2, Point{-0.1, -0.3})
	d3 := NewDataWithPosition(3, Point{-0.7, -0.7})
	d4 := NewDataWithPosition(4, Point{-0.3, -0.9})
	qt.Add(d1, d1.position)
	qt.Add(d2, d2.position)
	qt.Add(d3, d3.position)
	qt.Add(d4, d4.position)

	found := make([]interface{}, 0)
	qt.QueryIterative(Rect{-1, -1, 1, 1}, func(data interface{}, pos Point) bool {
		found = append(found, data)
		d := ToData(data)
		assert.Equal(t, d.position, pos)
		return true
	})
	assert.ElementsMatch(t, [4]*TestData{d1, d2, d3, d4}, found)
}

func TestQueryIterative_stopsWhenReturnFalse(t *testing.T) {
	qt, _ := NewQuadtree(Rect{-1, -1, 1, 1}, 2, 2)

	d1, d2, d3, d4 := NewData(1), NewData(2), NewData(3), NewData(4)
	qt.Add(d1, Point{-0.5, -0.1})
	qt.Add(d2, Point{-0.1, -0.3})
	qt.Add(d3, Point{-0.7, -0.7})
	qt.Add(d4, Point{-0.3, -0.9})

	count := 0
	qt.QueryIterative(Rect{-1, -1, 1, 1}, func(data interface{}, pos Point) bool {
		count += 1
		if (count > 1) {
			return false
		}
		return true
	})
	assert.Equal(t, 2, count)
}

func TestQuery_stressTest_returnsAllAddedPoints(t *testing.T) {
	bounds := Rect{-1, -1, 2, 2}
	qt, _ := NewQuadtree(bounds, 4, 4)

	numItems := 1000
	datas := make([]*TestData, 0, numItems)
	for i := 0; i < numItems; i++ {
		x := (2.0 * rand.Float64()) - 1.0
		y := (2.0 * rand.Float64()) - 1.0
		d := NewData(i)
		datas = append(datas, d)
		qt.Add(d, Point{x, y})
	}

	items := qt.Query(bounds)
	assert.ElementsMatch(t, datas, items)
}

func TestQuery_withNegativeCoordinates_returnsAllPoints(t *testing.T) {
	qt, _ := NewQuadtree(Rect{-1, -1, 1, 1}, 2, 2)

	d1, d2, d3, d4 := NewData(1), NewData(2), NewData(3), NewData(4)
	qt.Add(d1, Point{-0.5, -0.1})
	qt.Add(d2, Point{-0.1, -0.3})
	qt.Add(d3, Point{-0.7, -0.7})
	qt.Add(d4, Point{-0.3, -0.9})

	items := qt.Query(Rect{-1, -1, 1, 1})
	assert.ElementsMatch(t, [4]*TestData{d1, d2, d3, d4}, items)
}

func TestQuery_withPointRect_doesNotreturnItemAtPosition(t *testing.T) {
	bounds := Rect{0, 0, 5, 5}
	qt, _ := NewQuadtree(bounds, 2, 2)

	qt.Add(NewData(1), Point{1, 1})
	items := qt.Query(Rect{1, 1, 0, 0})

	assert.Len(t, items, 0)
}

func TestQuery_whenSplittingToMaxDepth_returnsAllItems(t *testing.T) {
	bounds := Rect{0, 0, 5, 5}
	// With max depth = 2 and max items per node = 2, we expect the tree
	// to split
	qt, _ := NewQuadtree(bounds, 2, 2)

	d1, d2, d3, d4, d5, d6 := NewData(1), NewData(2), NewData(3), NewData(4), NewData(5), NewData(6)
	qt.Add(d1, Point{1, 1})
	qt.Add(d2, Point{1, 1})
	qt.Add(d3, Point{1, 1})
	qt.Add(d4, Point{1, 1})
	qt.Add(d5, Point{1, 1})
	qt.Add(d6, Point{1, 1})

	items := qt.Query(bounds)
	assert.Len(t, items, 6)
	assert.ElementsMatch(t, [6]*TestData{d1, d2, d3, d4, d5, d6}, items)
}

func TestQuery_whenNoInternalSplitting_returnsAllItems(t *testing.T) {
	bounds := Rect{0, 0, 5, 5}

	qt, _ := NewQuadtree(bounds, 5, 5)

	d1, d2, d3, d4 := NewData(1), NewData(2), NewData(3), NewData(4)
	qt.Add(d1, Point{1, 1})
	qt.Add(d2, Point{1, 4})
	qt.Add(d3, Point{4, 1})
	qt.Add(d4, Point{4, 4})

	items := qt.Query(bounds)
	assert.Len(t, items, 4)
	assert.ElementsMatch(t, [4]*TestData{d1, d2, d3, d4}, items)
}

func TestAdd_outsideBoundsGeneratesError(t *testing.T) {
	qt, _ := NewQuadtree(Rect{0, 0, 1, 1}, 5, 5)

	// These points are outside the bounds (N, S, W, E of it)
	// They should all fail
	err := qt.Add(NewData(1), Point{0, 5})
	assert.NotNil(t, err)
	err = qt.Add(NewData(1), Point{0, -5})
	assert.NotNil(t, err)
	err = qt.Add(NewData(1), Point{5, 0})
	assert.NotNil(t, err)
	err = qt.Add(NewData(1), Point{-5, 0})
	assert.NotNil(t, err)
}

func TestAdd_insideBoundsGeneratesNoError(t *testing.T) {
	qt, _ := NewQuadtree(Rect{0, 0, 1, 1}, 5, 1)

	err := qt.Add(NewData(7), Point{0.5, 0.5})
	assert.Nil(t, err)
}

func TestSize_sizeIncrements(t *testing.T) {
	qt, _ := NewQuadtree(Rect{0, 0, 100, 100}, 5, 1)

	assert.Equal(t, 0, qt.Size())
	qt.Add(NewData(7), Point{20, 20})
	assert.Equal(t, 1, qt.Size())
	qt.Add(NewData(7), Point{20, 20})
	assert.Equal(t, 2, qt.Size())
	qt.Add(NewData(2), Point{10, 20})
	qt.Add(NewData(3), Point{20, 90})
	assert.Equal(t, 4, qt.Size())
}

func TestNewQuadTree_failsWhenDepthLessThan1(t *testing.T) {
	qt, err := NewQuadtree(Rect{0, 0, 1, 1}, 0, 5)
	assert.Nil(t, qt)
	assert.NotNil(t, err)
}

func TestNewQuadTree_failsWhenItemsPerNodeLessThan1(t *testing.T) {
	qt, err := NewQuadtree(Rect{0, 0, 1, 1}, 5, 0)
	assert.Nil(t, qt)
	assert.NotNil(t, err)
}

func TestNewQuadTree_succeedsWhenValidConfiguration(t *testing.T) {
	qt, err := NewQuadtree(Rect{0, 0, 1, 1}, 5, 10)
	assert.NotNil(t, qt)
	assert.Nil(t, err)
}
