package builder

type baseSelect struct {
	limit      int
	offset     int
	quantifier string
	table      string
	forceIndex string
	tail       string
	columns    []string
	order      []string
	group      []string
	where      *Predicate
	having     *Predicate
	join       *join
	params     []interface{}
}

func Select() *baseSelect {
	return &baseSelect{
		limit:  -1,
		offset: -1,
	}
}

func (s *baseSelect) addParams(params ...interface{}) {
	if s.params == nil {
		s.params = make([]interface{}, 0)
	}
	s.params = append(s.params, params...)
}

func (s *baseSelect) Table(table string) *baseSelect {
	s.table = table
	return s
}

func (s *baseSelect) Quantifier(quantifier string) *baseSelect {
	s.quantifier = quantifier
	return s
}

func (s *baseSelect) Columns(columns ...string) *baseSelect {
	s.columns = columns
	return s
}

func (s *baseSelect) InnerJoin(name string, on []string, columns ...string) *baseSelect {
	if s.join == nil {
		s.join = newJoin()
	}
	s.join.join(name, on, JoinInner, columns...)
	return s
}

func (s *baseSelect) OuterJoin(name string, on []string, columns ...string) *baseSelect {
	if s.join == nil {
		s.join = newJoin()
	}
	s.join.join(name, on, JoinOuter, columns...)
	return s
}

func (s *baseSelect) LeftJoin(name string, on []string, columns ...string) *baseSelect {
	if s.join == nil {
		s.join = newJoin()
	}
	s.join.join(name, on, JoinLeft, columns...)
	return s
}

func (s *baseSelect) RightJoin(name string, on []string, columns ...string) *baseSelect {
	if s.join == nil {
		s.join = newJoin()
	}
	s.join.join(name, on, JoinRight, columns...)
	return s
}

func (s *baseSelect) LeftOuterJoin(name string, on []string, columns ...string) *baseSelect {
	if s.join == nil {
		s.join = newJoin()
	}
	s.join.join(name, on, JoinLeftOuter, columns...)
	return s

}

func (s *baseSelect) RightOuterJoin(name string, on []string, columns ...string) *baseSelect {
	if s.join == nil {
		s.join = newJoin()
	}
	s.join.join(name, on, JoinRightOuter, columns...)
	return s
}

func (s *baseSelect) Where(where interface{}) *baseSelect {
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

func (s *baseSelect) Group(group ...string) *baseSelect {
	s.group = group
	return s
}

func (s *baseSelect) Having(having interface{}) *baseSelect {
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

func (s *baseSelect) Order(order ...string) *baseSelect {
	s.order = order
	return s
}

func (s *baseSelect) Limit(limit int) *baseSelect {
	if limit >= 0 {
		s.limit = limit
	}
	return s
}

func (s *baseSelect) Offset(offset int) *baseSelect {
	if offset >= 0 {
		s.offset = offset
	}
	return s
}

func (s *baseSelect) ForceIndex(index string) *baseSelect {
	s.forceIndex = index
	return s
}

func (s *baseSelect) Tail(tail string) *baseSelect {
	s.tail = tail
	return s
}
