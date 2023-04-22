package linkgraph

import (
	"time"

	"github.com/google/uuid"
)

type Link struct {
	ID          uuid.UUID
	URL         string
	RetreivedAt time.Time
}

type Edge struct {
	ID        uuid.UUID
	From      uuid.UUID
	To        uuid.UUID
	UpdatedAt time.Time
}

type LinkIterator interface {
	Link() *Link
}

type EdgeIterator interface {
	Edge() *Edge
}

type Graph interface {
	UpsertLink(link Link) (Link, error)
	FindLink(id uuid.UUID) (Link, error)
	Links(from, to uuid.UUID, retrievedBefore time.Time) (LinkIterator, error)

	UpsertEdge(edge Edge) (Edge, error)
	Edges(from, to uuid.UUID, updatedBefore time.Time) (EdgeIterator, error)
	RemoveStaleEdges(from uuid.UUID, updatedBefore time.Time) error
}
