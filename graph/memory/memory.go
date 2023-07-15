package memory

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/gmikhaile/links/graph"
)

var _ graph.Graph = (*Graph)(nil)

type Graph struct {
	mu sync.RWMutex

	links    map[uuid.UUID]*graph.Link
	linksURL map[string]*graph.Link

	edges     map[uuid.UUID]*graph.Edge // edge ID to edge
	linkEdges map[uuid.UUID][]uuid.UUID // link ID to slice of edge IDs
}

func NewMemoryGraph() *Graph {
	return &Graph{
		links:     make(map[uuid.UUID]*graph.Link),
		linksURL:  make(map[string]*graph.Link),
		edges:     make(map[uuid.UUID]*graph.Edge),
		linkEdges: make(map[uuid.UUID][]uuid.UUID),
	}
}

func (g *Graph) UpsertLink(link graph.Link) (graph.Link, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if existing := g.linksURL[link.URL]; existing != nil {
		if link.RetreivedAt.After(existing.RetreivedAt) {
			existing.RetreivedAt = link.RetreivedAt
		}

		return *existing, nil
	}

	newLink := new(graph.Link)
	*newLink = link

	for {
		id, err := uuid.NewRandom()
		if err != nil {
			return graph.Link{}, fmt.Errorf("can't generate uuid: %w", err)
		}

		if _, ok := g.links[id]; ok {
			continue
		}

		newLink.ID = id
		break
	}

	g.links[newLink.ID] = newLink
	g.linksURL[newLink.URL] = newLink

	return *newLink, nil
}

func (g *Graph) FindLink(id uuid.UUID) (graph.Link, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	link, ok := g.links[id]
	if !ok {
		return graph.Link{}, fmt.Errorf("link with id `%v` is not found", id)
	}

	return *link, nil
}

func (g *Graph) Links(
	from, to uuid.UUID, retrievedBefore time.Time,
) (graph.LinkIterator, error) {
	fromStr := from.String()
	toStr := to.String()

	g.mu.RLock()
	defer g.mu.RUnlock()

	var links []*graph.Link
	for _, link := range g.links {
		linkIDstr := link.ID.String()
		isInRangeFromTo := linkIDstr >= fromStr && linkIDstr < toStr
		isBeforeTime := link.RetreivedAt.Before(retrievedBefore)

		if isInRangeFromTo && isBeforeTime {
			links = append(links, link)
		}
	}

	if links == nil {
		return nil, fmt.Errorf(
			"no such range for links from `%s` to `%s` retreived before `%v`",
			fromStr, toStr, retrievedBefore,
		)
	}

	return &linkIterator{
		memoryGraph: g,
		links:       links,
	}, nil
}

func (g *Graph) UpsertEdge(edge graph.Edge) (graph.Edge, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	_, fromOk := g.links[edge.From]
	_, toOk := g.links[edge.To]
	if !fromOk || !toOk {
		return graph.Edge{}, fmt.Errorf("edge is not valid `%v`", edge)
	}

	for _, edgeID := range g.linkEdges[edge.From] {
		existingEdge := g.edges[edgeID]
		if existingEdge.From != edge.From || existingEdge.To != edge.To {
			continue
		}

		existingEdge.UpdatedAt = time.Now()
		return *existingEdge, nil
	}

	newEdge := new(graph.Edge)
	*newEdge = edge
	newEdge.UpdatedAt = time.Now()

	for {
		id, err := uuid.NewRandom()
		if err != nil {
			return graph.Edge{}, fmt.Errorf("can't generate uuid: %w", err)
		}

		if _, ok := g.edges[id]; ok {
			continue
		}

		newEdge.ID = id
		break
	}

	g.edges[newEdge.ID] = newEdge
	g.linkEdges[newEdge.From] = append(g.linkEdges[newEdge.From], newEdge.ID)

	return *newEdge, nil
}

func (g *Graph) Edges(from, to uuid.UUID, updated time.Time) (graph.EdgeIterator, error) {
	fromStr := from.String()
	toStr := to.String()

	g.mu.RLock()
	defer g.mu.RUnlock()

	var edges []*graph.Edge
	for linkID := range g.links {
		linkIDstr := linkID.String()
		isInRangeFromTo := linkIDstr >= fromStr && linkIDstr < toStr
		if !isInRangeFromTo {
			continue
		}

		for _, edgeID := range g.linkEdges[linkID] {
			edge := g.edges[edgeID]
			if edge.UpdatedAt.After(updated) {
				continue
			}

			edges = append(edges, edge)
		}
	}

	if edges == nil {
		return nil, fmt.Errorf("no such range for edges from `%s` to `%s` updated before `%v`", fromStr, toStr, updated)
	}

	return &edgeIterator{
		memoryGraph: g,
		edges:       edges,
	}, nil
}

func (g *Graph) RemoveStaleEdges(
	from uuid.UUID,
	updatedBefore time.Time,
) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	var newEdges []uuid.UUID
	for _, edgeID := range g.linkEdges[from] {
		edge := g.edges[edgeID]
		if edge.UpdatedAt.Before(updatedBefore) {
			delete(g.edges, edge.ID)
			continue
		}

		newEdges = append(newEdges, edge.ID)
	}

	g.linkEdges[from] = newEdges

	return nil
}
