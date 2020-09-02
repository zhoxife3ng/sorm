package builder

type Updater struct {
	table  string
	set    []string
	where  *Predicate
	join   *join
	limit  int
	offset int
	params []interface{}
}

func Update() *Updater {
	return &Updater{
		limit:  -1,
		offset: -1,
	}
}

func (u *Updater) addParams(params ...interface{}) {
	if u.params == nil {
		u.params = make([]interface{}, 0)
	}
	u.params = append(u.params, params...)
}

func (u *Updater) Table(table string) *Updater {
	u.table = table
	return u
}

func (u *Updater) Set(set map[string]interface{}) *Updater {
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

func (u *Updater) Where(where interface{}) *Updater {
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

func (u *Updater) InnerJoin(name string, on []string, columns ...string) *Updater {
	if u.join == nil {
		u.join = newJoin()
	}
	u.join.join(name, on, JoinInner, columns...)
	return u
}

func (u *Updater) OuterJoin(name string, on []string, columns ...string) *Updater {
	if u.join == nil {
		u.join = newJoin()
	}
	u.join.join(name, on, JoinOuter, columns...)
	return u
}

func (u *Updater) LeftJoin(name string, on []string, columns ...string) *Updater {
	if u.join == nil {
		u.join = newJoin()
	}
	u.join.join(name, on, JoinLeft, columns...)
	return u
}

func (u *Updater) RightJoin(name string, on []string, columns ...string) *Updater {
	if u.join == nil {
		u.join = newJoin()
	}
	u.join.join(name, on, JoinRight, columns...)
	return u
}

func (u *Updater) LeftOuterJoin(name string, on []string, columns ...string) *Updater {
	if u.join == nil {
		u.join = newJoin()
	}
	u.join.join(name, on, JoinLeftOuter, columns...)
	return u

}

func (u *Updater) RightOuterJoin(name string, on []string, columns ...string) *Updater {
	if u.join == nil {
		u.join = newJoin()
	}
	u.join.join(name, on, JoinRightOuter, columns...)
	return u
}
