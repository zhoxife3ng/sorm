package builder

type Deleter struct {
	table  string
	where  *Predicate
	params []interface{}
}

func Delete() *Deleter {
	return &Deleter{}
}

func (d *Deleter) addParams(params ...interface{}) {
	if d.params == nil {
		d.params = make([]interface{}, 0)
	}
	d.params = append(d.params, params...)
}

func (d *Deleter) Table(table string) *Deleter {
	d.table = table
	return d
}

func (d *Deleter) Where(where interface{}) *Deleter {
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
