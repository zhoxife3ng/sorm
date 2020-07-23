package builder

type baseUpdate struct {
	table  string
	set    []string
	where  *Predicate
	join   *join
	limit  int
	offset int
	params []interface{}
}

func Update() *baseUpdate {
	return &baseUpdate{
		limit:  -1,
		offset: -1,
	}
}

func (u *baseUpdate) addParams(params ...interface{}) {
	if u.params == nil {
		u.params = make([]interface{}, 0)
	}
	u.params = append(u.params, params...)
}

func (u *baseUpdate) Table(table string) *baseUpdate {
	u.table = table
	return u
}

func (u *baseUpdate) Set(set map[string]interface{}) *baseUpdate {
	if u.set == nil {
		u.set = make([]string, 0)
		u.params = make([]interface{}, 0)
	} else {
		u.set = u.set[0:0]
		u.params = u.params[0:0]
	}
	for k, v := range set {
		u.set = append(u.set, k)
		u.addParams(v)
	}
	return u
}

func (u *baseUpdate) Where(where interface{}) *baseUpdate {
	var wherePredicate *Predicate
	if where == nil {
		return u
	} else if w, ok := where.(*clause); ok {
		if w == nil {
			return u
		}
		wherePredicate = w.predicate
	} else {
		wherePredicate = Clause(where).predicate
	}
	if u.where == nil {
		u.where = wherePredicate
	} else {
		u.where.AddPredicate(wherePredicate, CombinedByAnd)
	}
	return u
}

func (u *baseUpdate) InnerJoin(name string, on []string, columns ...string) *baseUpdate {
	if u.join == nil {
		u.join = newJoin()
	}
	u.join.join(name, on, JoinInner, columns...)
	return u
}

func (u *baseUpdate) OuterJoin(name string, on []string, columns ...string) *baseUpdate {
	if u.join == nil {
		u.join = newJoin()
	}
	u.join.join(name, on, JoinOuter, columns...)
	return u
}

func (u *baseUpdate) LeftJoin(name string, on []string, columns ...string) *baseUpdate {
	if u.join == nil {
		u.join = newJoin()
	}
	u.join.join(name, on, JoinLeft, columns...)
	return u
}

func (u *baseUpdate) RightJoin(name string, on []string, columns ...string) *baseUpdate {
	if u.join == nil {
		u.join = newJoin()
	}
	u.join.join(name, on, JoinRight, columns...)
	return u
}

func (u *baseUpdate) LeftOuterJoin(name string, on []string, columns ...string) *baseUpdate {
	if u.join == nil {
		u.join = newJoin()
	}
	u.join.join(name, on, JoinLeftOuter, columns...)
	return u

}

func (u *baseUpdate) RightOuterJoin(name string, on []string, columns ...string) *baseUpdate {
	if u.join == nil {
		u.join = newJoin()
	}
	u.join.join(name, on, JoinRightOuter, columns...)
	return u
}
