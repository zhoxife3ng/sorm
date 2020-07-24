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
				order:  []string{"age DESC", "score ASC"},
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
