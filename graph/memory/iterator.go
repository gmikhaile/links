package memory

import graph "github.com/gmikhaile/links/graph"

var _ graph.LinkIterator = (*linkIterator)(nil)

type linkIterator struct {
	memoryGraph *Graph
	links       []*graph.Link
	curIndex    int
}

func (it *linkIterator) Link() *graph.Link {
	if !it.next() {
		return nil
	}

	it.memoryGraph.mu.RLock()
	defer it.memoryGraph.mu.RUnlock()

	link := new(graph.Link)
	*link = *it.links[it.curIndex-1]

	return link
}

func (it *linkIterator) next() bool {
	if it.curIndex >= len(it.links) {
		return false
	}

	it.curIndex++
	return true
}

type edgeIterator struct {
	memoryGraph *Graph
	edges       []*graph.Edge
	curIndex    int
}

func (it *edgeIterator) Edge() *graph.Edge {
	if !it.next() {
		return nil
	}

	it.memoryGraph.mu.RLock()
	defer it.memoryGraph.mu.RUnlock()

	edge := new(graph.Edge)
	*edge = *it.edges[it.curIndex-1]

	return edge
}

func (it *edgeIterator) next() bool {
	if it.curIndex >= len(it.edges) {
		return false
	}

	it.curIndex++
	return true
}
