package cdb

import (
	"database/sql"

	"github.com/gmikhaile/links/graph"
)

var _ graph.LinkIterator = (*linkIterator)(nil)

type linkIterator struct {
	latchedLink *graph.Link
	rows        *sql.Rows
	lastErr     error
}

func (it *linkIterator) Link() *graph.Link {
	if !it.next() {
		return nil
	}

	return it.latchedLink
}

func (it *linkIterator) next() bool {
	if it.lastErr != nil || !it.rows.Next() {
		return false
	}

	nextLink := graph.Link{}
	it.lastErr = it.rows.Scan(nextLink.ID, nextLink.URL, nextLink.RetreivedAt)
	if it.lastErr != nil {
		return false
	}

	nextLink.RetreivedAt = nextLink.RetreivedAt.UTC()
	it.latchedLink = &nextLink

	return true
}

var _ graph.EdgeIterator = (*edgeIterator)(nil)

type edgeIterator struct {
	latchedEdge *graph.Edge
	rows        *sql.Rows
	lastErr     error
}

func (it *edgeIterator) Edge() *graph.Edge {
	if !it.next() {
		return nil
	}

	return it.latchedEdge
}

func (it *edgeIterator) next() bool {
	if it.lastErr != nil || !it.rows.Next() {
		return false
	}

	nextEdge := graph.Edge{}
	it.lastErr = it.rows.Scan(nextEdge.ID, nextEdge.From, nextEdge.To, nextEdge.UpdatedAt)
	if it.lastErr != nil {
		return false
	}

	nextEdge.UpdatedAt = nextEdge.UpdatedAt.UTC()
	it.latchedEdge = &nextEdge

	return true
}
