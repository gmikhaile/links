package memory

import (
	"testing"

	"gopkg.in/check.v1"

	"github.com/gmikhaile/links/graph/graphtest"
)

var _ = check.Suite(new(InMemoryGraphTestSuite))

func Test(t *testing.T) {
	check.TestingT(t)
}

type InMemoryGraphTestSuite struct {
	graphtest.SuiteBase
}

func (s *InMemoryGraphTestSuite) SetUpTest(c *check.C) {
	s.SetGraph(NewMemoryGraph())
}
