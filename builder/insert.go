package builder

type baseInsert struct {
	table   string
	columns []string
	params  [][]interface{}
}

func Insert() *baseInsert {
	return &baseInsert{}
}

func (i *baseInsert) addParams(params ...[]interface{}) {
	if i.params == nil {
		i.params = make([][]interface{}, 0)
	}
	i.params = append(i.params, params...)
}

func (i *baseInsert) Table(table string) *baseInsert {
	i.table = table
	return i
}

func (i *baseInsert) Values(values ...map[string]interface{}) *baseInsert {
	if i.columns == nil {
		i.columns = make([]string, 0)
		i.params = make([][]interface{}, 0)
	} else {
		i.columns = i.columns[0:0]
		i.params = i.params[0:0]
	}
	for _, value := range values {
		var params = make([]interface{}, 0)
		if len(i.columns) == 0 {
			for k, v := range value {
				i.columns = append(i.columns, k)
				params = append(params, v)
			}
		} else {
			for _, v := range i.columns {
				if val, ok := value[v]; ok {
					params = append(params, val)
					delete(value, v)
				} else {
					params = append(params, nil)
				}
			}
			for k, v := range value {
				i.columns = append(i.columns, k)
				params = append(params, v)
			}
		}
		i.addParams(params)
	}
	return i
}
