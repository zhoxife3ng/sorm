package builder

type Selector struct {
	limit      int
	offset     int
	quantifier string
	table      string
	forceIndex string
	tail       string
	fColumns   map[string]string
	columns    []string
	order      []string
	group      []string
	where      *Predicate
	having     *Predicate
	join       *join
	params     []interface{}
}

func Select() *Selector {
	return &Selector{
		limit:  -1,
		offset: -1,
	}
}

func (s *Selector) addParams(params ...interface{}) {
	if s.params == nil {
		s.params = make([]interface{}, 0)
	}
	s.params = append(s.params, params...)
}

func (s *Selector) Table(table string) *Selector {
	s.table = table
	return s
}

func (s *Selector) Quantifier(quantifier string) *Selector {
	s.quantifier = quantifier
	return s
}

func (s *Selector) FuncColumns(fColumns map[string]string) *Selector {
	s.fColumns = fColumns
	return s
}

func (s *Selector) Columns(columns ...string) *Selector {
	s.columns = columns
	return s
}

func (s *Selector) InnerJoin(name string, on []string, columns ...string) *Selector {
	if s.join == nil {
		s.join = newJoin()
	}
	s.join.join(name, on, JoinInner, columns...)
	return s
}

func (s *Selector) OuterJoin(name string, on []string, columns ...string) *Selector {
	if s.join == nil {
		s.join = newJoin()
	}
	s.join.join(name, on, JoinOuter, columns...)
	return s
}

func (s *Selector) LeftJoin(name string, on []string, columns ...string) *Selector {
	if s.join == nil {
		s.join = newJoin()
	}
	s.join.join(name, on, JoinLeft, columns...)
	return s
}

func (s *Selector) RightJoin(name string, on []string, columns ...string) *Selector {
	if s.join == nil {
		s.join = newJoin()
	}
	s.join.join(name, on, JoinRight, columns...)
	return s
}

func (s *Selector) LeftOuterJoin(name string, on []string, columns ...string) *Selector {
	if s.join == nil {
		s.join = newJoin()
	}
	s.join.join(name, on, JoinLeftOuter, columns...)
	return s

}

func (s *Selector) RightOuterJoin(name string, on []string, columns ...string) *Selector {
	if s.join == nil {
		s.join = newJoin()
	}
	s.join.join(name, on, JoinRightOuter, columns...)
	return s
}

func (s *Selector) Where(where interface{}) *Selector {
	var wherePredicate *Predicate
	if where == nil {
		return s
	} else if w, ok := where.(*clause); ok {
		if w == nil {
			return s
		}
		wherePredicate = w.predicate
	} else {
		wherePredicate = Clause(where).predicate
	}
	if s.where == nil {
		s.where = wherePredicate
	} else {
		s.where.AddPredicate(wherePredicate, CombinedByAnd)
	}
	return s
}

func (s *Selector) Group(group ...string) *Selector {
	s.group = group
	return s
}

func (s *Selector) Having(having interface{}) *Selector {
	var havingPredicate *Predicate
	if w, ok := having.(*clause); ok {
		havingPredicate = w.predicate
	} else {
		havingPredicate = Clause(having).predicate
	}
	if s.having == nil {
		s.having = havingPredicate
	} else {
		s.having.AddPredicate(havingPredicate, CombinedByAnd)
	}
	return s
}

func (s *Selector) Order(order ...string) *Selector {
	s.order = order
	return s
}

func (s *Selector) Limit(limit int) *Selector {
	if limit >= 0 {
		s.limit = limit
	}
	return s
}

func (s *Selector) Offset(offset int) *Selector {
	if offset >= 0 {
		s.offset = offset
	}
	return s
}

func (s *Selector) ForceIndex(index string) *Selector {
	s.forceIndex = index
	return s
}

func (s *Selector) Tail(tail string) *Selector {
	s.tail = tail
	return s
}
