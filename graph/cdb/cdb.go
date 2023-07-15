package cdb

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/gmikhaile/links/graph"
	"github.com/google/uuid"
)

const (
	upsertLinkQuery = `
INSERT INTO links (url, retrieved_at) VALUES ($1, $2) 
ON CONFLICT (url) DO UPDATE SET retrieved_at=GREATEST(links.retrieved_at, $2)
RETURNING id, retrieved_at
`
	findLinkQuery         = "SELECT url, retrieved_at FROM links WHERE id=$1"
	linksInPartitionQuery = "SELECT id, url, retrieved_at FROM links WHERE id >= $1 AND id < $2 AND retrieved_at < $3"

	upsertEdgeQuery = `
INSERT INTO edges (src, dst, updated_at) VALUES ($1, $2, NOW())
ON CONFLICT (src,dst) DO UPDATE SET updated_at=NOW()
RETURNING id, updated_at
`
	edgesInPartitionQuery = "SELECT id, src, dst, updated_at FROM edges WHERE src >= $1 AND src < $2 AND updated_at < $3"
	removeStaleEdgesQuery = "DELETE FROM edges WHERE src=$1 AND updated_at < $2"
)

var _ graph.Graph = (*CockroachDB)(nil)

type CockroachDB struct {
	db *sql.DB
}

func NewCockroachDB(dsn string) (CockroachDB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return CockroachDB{}, fmt.Errorf("can't open sql connection: %w", err)
	}

	return CockroachDB{db: db}, nil
}

func (d *CockroachDB) Close() error {
	err := d.db.Close()
	if err != nil {
		return fmt.Errorf("can't close connection: %w", err)
	}

	return nil
}

func (d *CockroachDB) UpsertLink(link graph.Link) (graph.Link, error) {
	rows, err := d.db.Query(upsertLinkQuery, link.URL, link.RetreivedAt.UTC())
	if err != nil {
		return graph.Link{}, fmt.Errorf("failed to upsert link: %w", err)
	}

	var upsertedLink graph.Link
	if err := rows.Scan(upsertedLink.ID, upsertedLink.RetreivedAt); err != nil {
		return graph.Link{}, fmt.Errorf("failed to scan upserted link: %w", err)
	}
	return upsertedLink, nil
}

func (d *CockroachDB) FindLink(id uuid.UUID) (graph.Link, error) {
	rows, err := d.db.Query(findLinkQuery, id)
	if err != nil {
		return graph.Link{}, fmt.Errorf("failed to find link: %w", err)
	}

	link := graph.Link{ID: id}
	if err := rows.Scan(link.URL, link.RetreivedAt); err != nil {
		return graph.Link{}, fmt.Errorf("failed to scan founded link: %w", err)
	}

	link.RetreivedAt.UTC()

	return link, nil
}
func (d *CockroachDB) Links(from, to uuid.UUID, retrievedBefore time.Time) (graph.LinkIterator, error) {
	rows, err := d.db.Query(linksInPartitionQuery, from, to, retrievedBefore)
	if err != nil {
		return nil, fmt.Errorf("failed to find links: %w", err)
	}

	return &linkIterator{
		rows: rows,
	}, nil
}
func (d *CockroachDB) UpsertEdge(edge graph.Edge) (graph.Edge, error) {
	rows, err := d.db.Query(upsertEdgeQuery, edge.From, edge.To, edge.UpdatedAt.UTC())
	if err != nil {
		return graph.Edge{}, fmt.Errorf("failed to upsert edge: %w", err)
	}

	var upserteEdge graph.Edge
	if err := rows.Scan(upserteEdge.ID, upserteEdge.UpdatedAt); err != nil {
		return graph.Edge{}, fmt.Errorf("failed to scan upserted edge: %w", err)
	}
	return upserteEdge, nil
}

func (d *CockroachDB) Edges(from, to uuid.UUID, updatedBefore time.Time) (graph.EdgeIterator, error) {
	rows, err := d.db.Query(edgesInPartitionQuery, from, to, updatedBefore.UTC())
	if err != nil {
		return nil, fmt.Errorf("failed to upsert edge: %w", err)
	}

	return &edgeIterator{rows: rows}, nil
}

func (d *CockroachDB) RemoveStaleEdges(from uuid.UUID, updatedBefore time.Time) error {
	_, err := d.db.Exec(edgesInPartitionQuery, from, updatedBefore.UTC())
	if err != nil {
		return fmt.Errorf("failed to upsert edge: %w", err)
	}

	return nil
}
