package builder

type baseDelete struct {
	table  string
	where  *Predicate
	params []interface{}
}

func Delete() *baseDelete {
	return &baseDelete{}
}

func (d *baseDelete) addParams(params ...interface{}) {
	if d.params == nil {
		d.params = make([]interface{}, 0)
	}
	d.params = append(d.params, params...)
}

func (d *baseDelete) Table(table string) *baseDelete {
	d.table = table
	return d
}

func (d *baseDelete) Where(where interface{}) *baseDelete {
	var wherePredicate *Predicate
	if where == nil {
		return d
	} else if w, ok := where.(*clause); ok {
		if w == nil {
			return d
		}
		wherePredicate = w.predicate
	} else {
		wherePredicate = Clause(where).predicate
	}
	if d.where == nil {
		d.where = wherePredicate
	} else {
		d.where.AddPredicate(wherePredicate, CombinedByAnd)
	}
	return d
}
