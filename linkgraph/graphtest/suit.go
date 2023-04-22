package graphtest

import (
	"time"

	"github.com/google/uuid"
	gc "gopkg.in/check.v1"

	"github.com/gmikhaile/links/linkgraph"
)

type SuiteBase struct {
	g linkgraph.Graph
}

func (s *SuiteBase) SetGraph(g linkgraph.Graph) {
	s.g = g
}

func (s *SuiteBase) TestUpserNewLink(check *gc.C) {
	actual, err := s.g.UpsertLink(linkgraph.Link{
		URL:         "test upsert new link",
		RetreivedAt: time.Now(),
	})
	check.Assert(err, gc.IsNil, gc.Commentf("failed to upsert link: %d", actual))
	check.Assert(actual.ID, gc.Not(gc.Equals), uuid.Nil, gc.Commentf("actual link id is nil"))
}

func (s *SuiteBase) TestUpdateLink(check *gc.C) {
	actual, err := s.g.UpsertLink(linkgraph.Link{
		URL:         "test update link",
		RetreivedAt: time.Now(),
	})
	check.Assert(err, gc.IsNil, gc.Commentf("failed to upsert link: %d", actual))
	check.Assert(actual.ID, gc.Not(gc.Equals), uuid.Nil, gc.Commentf("actual link id is nil"))

	updatedLink, err := s.g.UpsertLink(actual)
	check.Assert(err, gc.IsNil, gc.Commentf("failed to update link: %d", actual))
	check.Assert(updatedLink.ID, gc.Equals, actual.ID, gc.Commentf("link ID changed while upserting"))
}
