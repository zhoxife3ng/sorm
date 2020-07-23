package builder

import (
	"github.com/x554462/sorm/builder/predicate"
	"strings"
)

const (
	JoinInner      = "INNER"
	JoinOuter      = "OUTER"
	JoinLeft       = "LEFT"
	JoinRight      = "RIGHT"
	JoinLeftOuter  = "LEFT OUTER"
	JoinRightOuter = "RIGHT OUTER"
)

type joinAttr struct {
	name    string
	on      string
	columns []string
	typo    string
}

type join struct {
	joins []joinAttr
}

func newJoin() *join {
	return &join{joins: make([]joinAttr, 0)}
}

func (j *join) GetJoins() []joinAttr {
	return j.joins
}

func (j *join) join(name string, on []string, joinType string, columns ...string) *join {
	for i, onv := range on {
		on[i] = predicate.QuoteIdentifier(onv)
	}
	j.joins = append(j.joins, joinAttr{
		name:    name,
		on:      strings.Join(on, " = "),
		columns: columns,
		typo:    joinType,
	})
	return j
}

func (j *join) count() int {
	return len(j.joins)
}
