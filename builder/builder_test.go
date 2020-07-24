package builder

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBaseSelect_Build(t *testing.T) {
	type inStruct struct {
		table  string
		fields []string
		where  map[string]interface{}
		or     []map[string]interface{}
		group  []string
		having map[string]interface{}
		order  []string
		limit  int
		offset int
		tail   string
	}
	type outStruct struct {
		cond   string
		params []interface{}
		err    error
	}
	var data = []struct {
		in  inStruct
		out outStruct
	}{
		{
			in: inStruct{
				table:  "tb",
				fields: []string{"id", "name", "age"},
				where: map[string]interface{}{
					"foo":    "bar",
					"qq":     "tt",
					"age":    []interface{}{1, 3, 5, 7, 9},
					"!faith": "Muslim",
				},
				or: []map[string]interface{}{
					{
						"aa": 11,
						"bb": "xswl",
					},
					{
						"cc": "234",
						"dd": []interface{}{7, 8},
					},
				},
				group: []string{"department"},
				having: map[string]interface{}{
					"`id`>?": 0,
				},
				order:  []string{"age DESC", "score "},
				limit:  0,
				offset: 100,
				tail:   "FOR UPDATE",
			},
			out: outStruct{
				cond:   "SELECT `tb`.`id`, `tb`.`name`, `tb`.`age` FROM `tb` WHERE `foo`=? AND `qq`=? AND `age` IN (?,?,?,?,?) AND `faith`!=? AND ((`aa`=? AND `bb`=?) OR (`cc`=? AND `dd` IN (?,?))) GROUP BY `department` HAVING `id`>? ORDER BY `age` DESC, `score` ASC LIMIT ? OFFSET ? FOR UPDATE",
				params: []interface{}{"bar", "tt", 1, 3, 5, 7, 9, "Muslim", 11, "xswl", "234", 7, 8, 0, 0, 100},
				err:    nil,
			},
		},
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"foo":                    "bar",
					"`foo2` BETWEEN ? AND ?": []interface{}{1, 2},
				},
				tail:   "FOR UPDATE",
				limit:  -1,
				offset: -1,
			},
			out: outStruct{
				cond:   "SELECT `tb`.* FROM `tb` WHERE `foo`=? AND `foo2` BETWEEN ? AND ? FOR UPDATE",
				params: []interface{}{"bar", 1, 2},
				err:    nil,
			},
		},
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"foo":  "bar",
					"foo2": "bar2",
				},
				tail:   "FOR UPDATE",
				limit:  -1,
				offset: -1,
			},
			out: outStruct{
				cond:   "SELECT `tb`.* FROM `tb` WHERE `foo`=? AND `foo2`=? FOR UPDATE",
				params: []interface{}{"bar", "bar2"},
				err:    nil,
			},
		},
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"foo":  "bar",
					"foo2": "bar2",
				},
				tail:   "FOR UPDATE",
				order:  []string{"we DASC"},
				limit:  -1,
				offset: -1,
			},
			out: outStruct{
				cond:   "",
				params: nil,
				err:    ErrProcessOrder,
			},
		},
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"foo":      "bar",
					"`foo2`=?": []interface{}{},
				},
				tail:   "FOR UPDATE",
				limit:  -1,
				offset: -1,
			},
			out: outStruct{
				cond:   "",
				params: nil,
				err:    ErrBuildPlaceHolder,
			},
		},
	}
	ass := assert.New(t)
	for _, tc := range data {
		var where = Clause(tc.in.where)
		clause := EmptyClause()
		for _, o := range tc.in.or {
			clause.Or(Clause(o))
		}
		where.And(clause)
		sel := Select().Table(tc.in.table).
			Columns(tc.in.fields...).
			Where(where).
			Group(tc.in.group...).
			Having(tc.in.having).
			Order(tc.in.order...).
			Limit(tc.in.limit).
			Offset(tc.in.offset).
			Tail(tc.in.tail)
		cond, params, err := sel.Build()
		ass.Equal(tc.out.err, err)
		ass.Equal(tc.out.cond, cond)
		ass.Equal(tc.out.params, params)
	}
}

func TestBaseSelectJoin_Build(t *testing.T) {
	type inStruct struct {
		table     string
		fields    []string
		where     map[string]interface{}
		or        []map[string]interface{}
		group     []string
		having    map[string]interface{}
		order     []string
		limit     int
		offset    int
		tail      string
		innerJoin struct {
			name    string
			on      []string
			columns []string
		}
	}
	type outStruct struct {
		cond   string
		params []interface{}
		err    error
	}
	var data = []struct {
		in  inStruct
		out outStruct
	}{
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"foo":  "bar",
					"foo2": "bar2",
				},
				innerJoin: struct {
					name    string
					on      []string
					columns []string
				}{name: "tb2", on: []string{"tb.id", "tb2.id"}, columns: []string{"*"}},
				tail:   "FOR UPDATE",
				limit:  -1,
				offset: -1,
			},
			out: outStruct{
				cond:   "SELECT `tb`.*, `tb2`.* FROM `tb` INNER JOIN `tb2` ON `tb`.`id`=`tb2`.`id` WHERE `foo`=? AND `foo2`=? FOR UPDATE",
				params: []interface{}{"bar", "bar2"},
				err:    nil,
			},
		},
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"foo":   "bar",
					"foo2":  "bar2",
					"t2.id": 1,
				},
				innerJoin: struct {
					name    string
					on      []string
					columns []string
				}{name: "tb2 AS t2", on: []string{"tb.id", "tb2.id"}, columns: []string{"id"}},
				tail:   "FOR UPDATE",
				limit:  -1,
				offset: -1,
			},
			out: outStruct{
				cond:   "SELECT `tb`.*, `t2`.`id` FROM `tb` INNER JOIN `tb2` AS `t2` ON `tb`.`id`=`tb2`.`id` WHERE `foo`=? AND `foo2`=? AND `t2`.`id`=? FOR UPDATE",
				params: []interface{}{"bar", "bar2", 1},
				err:    nil,
			},
		},
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"foo":   "bar",
					"foo2":  "bar2",
					"t2.id": 1,
				},
				innerJoin: struct {
					name    string
					on      []string
					columns []string
				}{name: "tb2 AS t2", on: []string{"tb.id", "tb2.id"}},
				limit:  -1,
				offset: -1,
			},
			out: outStruct{
				cond:   "SELECT `tb`.* FROM `tb` INNER JOIN `tb2` AS `t2` ON `tb`.`id`=`tb2`.`id` WHERE `foo`=? AND `foo2`=? AND `t2`.`id`=?",
				params: []interface{}{"bar", "bar2", 1},
				err:    nil,
			},
		},
	}
	ass := assert.New(t)
	for _, tc := range data {
		var where = Clause(tc.in.where)
		clause := EmptyClause()
		for _, o := range tc.in.or {
			clause.Or(Clause(o))
		}
		where.And(clause)
		sel := Select().Table(tc.in.table).
			Columns(tc.in.fields...).
			Where(where).
			Group(tc.in.group...).
			Having(tc.in.having).
			Order(tc.in.order...).
			Limit(tc.in.limit).
			Offset(tc.in.offset).
			Tail(tc.in.tail).InnerJoin(tc.in.innerJoin.name, tc.in.innerJoin.on, tc.in.innerJoin.columns...)
		cond, params, err := sel.Build()
		ass.Equal(tc.out.err, err)
		ass.Equal(tc.out.cond, cond)
		ass.Equal(tc.out.params, params)
	}
}
func TestBaseUpdate_Build(t *testing.T) {
	type inStruct struct {
		table string
		set   map[string]interface{}
		where map[string]interface{}
		or    []map[string]interface{}
	}
	type outStruct struct {
		cond   string
		params []interface{}
		err    error
	}
	var data = []struct {
		in  inStruct
		out outStruct
	}{
		{
			in: inStruct{
				table: "tb",
				set: map[string]interface{}{
					"foo":           "foo2",
					"qq":            "qq2",
					"`inc`=`inc`+?": 2,
				},
				where: map[string]interface{}{
					"foo":    "bar",
					"qq":     "tt",
					"age":    []interface{}{1, 3, 5, 7, 9},
					"!faith": "Muslim",
				},
				or: []map[string]interface{}{
					{
						"aa": 11,
						"bb": "xswl",
					},
					{
						"cc": "234",
						"dd": []interface{}{7, 8},
					},
				},
			},
			out: outStruct{
				cond:   "UPDATE `tb` SET `foo`=?, `qq`=?, `inc`=`inc`+? WHERE `foo`=? AND `qq`=? AND `age` IN (?,?,?,?,?) AND `faith`!=? AND ((`aa`=? AND `bb`=?) OR (`cc`=? AND `dd` IN (?,?)))",
				params: []interface{}{"foo2", "qq2", 2, "bar", "tt", 1, 3, 5, 7, 9, "Muslim", 11, "xswl", "234", 7, 8},
				err:    nil,
			},
		},
	}
	ass := assert.New(t)
	for _, tc := range data {
		var where = Clause(tc.in.where)
		clause := EmptyClause()
		for _, o := range tc.in.or {
			clause.Or(Clause(o))
		}
		where.And(clause)
		up := Update().Table(tc.in.table).
			Set(tc.in.set).
			Where(where)
		cond, params, err := up.Build()
		ass.Equal(tc.out.err, err)
		ass.Equal(tc.out.cond, cond)
		ass.Equal(tc.out.params, params)
	}
}

func TestBaseInsert_Build(t *testing.T) {
	type inStruct struct {
		table  string
		values []map[string]interface{}
	}
	type outStruct struct {
		cond   string
		params []interface{}
		err    error
	}
	var data = []struct {
		in  inStruct
		out outStruct
	}{
		{
			in: inStruct{
				table: "tb",
				values: []map[string]interface{}{
					{
						"aa": 11,
						"bb": "xswl",
					},
					{
						"aa": 22,
						"cc": "234",
						"dd": "3",
					},
				},
			},
			out: outStruct{
				cond:   "INSERT INTO `tb`(`aa`, `bb`, `cc`, `dd`) VALUES(?,?,?,?), (?,?,?,?)",
				params: []interface{}{11, "xswl", nil, nil, 22, nil, "234", "3"},
				err:    nil,
			},
		},
	}
	ass := assert.New(t)
	for _, tc := range data {
		i := Insert().Table(tc.in.table).Values(tc.in.values...)
		cond, params, err := i.Build()
		ass.Equal(tc.out.err, err)
		ass.Equal(tc.out.cond, cond)
		ass.Equal(tc.out.params, params)
	}
}

func TestBaseDelete_Build(t *testing.T) {
	type inStruct struct {
		table string
		where map[string]interface{}
		or    []map[string]interface{}
	}
	type outStruct struct {
		cond   string
		params []interface{}
		err    error
	}
	var data = []struct {
		in  inStruct
		out outStruct
	}{
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"foo":    "bar",
					"qq":     "tt",
					"age":    []interface{}{1, 3, 5, 7, 9},
					"!faith": "Muslim",
				},
				or: []map[string]interface{}{
					{
						"aa": 11,
						"bb": "xswl",
					},
					{
						"cc": "234",
						"dd": []interface{}{7, 8},
					},
				},
			},
			out: outStruct{
				cond:   "DELETE FROM `tb` WHERE `foo`=? AND `qq`=? AND `age` IN (?,?,?,?,?) AND `faith`!=? AND ((`aa`=? AND `bb`=?) OR (`cc`=? AND `dd` IN (?,?)))",
				params: []interface{}{"bar", "tt", 1, 3, 5, 7, 9, "Muslim", 11, "xswl", "234", 7, 8},
				err:    nil,
			},
		},
	}
	ass := assert.New(t)
	for _, tc := range data {
		var where = Clause(tc.in.where)
		clause := EmptyClause()
		for _, o := range tc.in.or {
			clause.Or(Clause(o))
		}
		where.And(clause)
		del := Delete().Table(tc.in.table).Where(where)
		cond, params, err := del.Build()
		ass.Equal(tc.out.err, err)
		ass.Equal(tc.out.cond, cond)
		ass.Equal(tc.out.params, params)
	}
}
