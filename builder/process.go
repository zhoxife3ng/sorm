package builder

import (
	"errors"
	"github.com/x554462/sorm/builder/predicate"
	"strings"
)

func (s *baseSelect) processSelect() (string, error) {
	var (
		selectQuantifier = s.quantifier
		selectTable      = s.table
		selectColumns    = s.columns
		selectJoin       = s.join
	)

	var str = getStrB()
	defer putStrB(str)
	str.WriteString("SELECT ")
	if selectQuantifier != "" {
		str.WriteString(selectQuantifier)
		str.WriteString(" ")
	}
	_, aliasTableName := predicate.ResolveIdentifier(selectTable)
	for i, c := range selectColumns {
		if i > 0 {
			str.WriteString(", ")
		}
		if !strings.Contains(c, ".") {
			str.WriteString(predicate.QuoteTable(aliasTableName))
			str.WriteString(".")
			str.WriteString(predicate.QuoteIdentifier(c))
		} else {
			str.WriteString(predicate.QuoteIdentifier(c))
		}
	}
	if len(selectColumns) == 0 {
		str.WriteString(predicate.QuoteTable(aliasTableName))
		str.WriteString(".*")
	}
	if selectJoin != nil {
		for _, joinAttr := range selectJoin.GetJoins() {
			_, aliasTableName := predicate.ResolveIdentifier(joinAttr.name)
			for _, c := range joinAttr.columns {
				str.WriteString(", ")
				if !strings.Contains(c, ".") {
					str.WriteString(aliasTableName)
					str.WriteString(".")
					str.WriteString(predicate.QuoteIdentifier(c))
				} else {
					str.WriteString(predicate.QuoteIdentifier(c))
				}
			}
		}

	}
	if selectTable != "" {
		str.WriteString(" FROM ")
		str.WriteString(predicate.QuoteTable(selectTable))
	}
	return str.String(), nil
}

func (s *baseSelect) processForceIndex() (string, error) {
	if s.forceIndex == "" {
		return "", nil
	}
	var str = getStrB()
	defer putStrB(str)
	str.WriteString("FORCE INDEX (")
	str.WriteString(predicate.QuoteIdentifier(s.forceIndex))
	str.WriteString(")")
	return str.String(), nil
}

func (s *baseSelect) processJoins() (string, error) {
	join := s.join
	if join == nil || join.count() == 0 {
		return "", nil
	}
	var str = getStrB()
	defer putStrB(str)
	for _, joinAttr := range join.GetJoins() {
		if str.Len() > 0 {
			str.WriteString(" ")
		}
		str.WriteString(joinAttr.typo)
		str.WriteString(" JOIN ")
		str.WriteString(predicate.QuoteTable(joinAttr.name))
		str.WriteString(" ON ")
		str.WriteString(joinAttr.on)
	}
	return str.String(), nil
}

func (s *baseSelect) processWhere() (string, error) {
	selectWhere := s.where
	if selectWhere == nil || selectWhere.count() == 0 {
		return "", nil
	}
	var (
		parts, _ = selectWhere.GetExpressionData()
		where    = getStrB()
	)
	defer putStrB(where)
	where.WriteString("WHERE ")
	for _, part := range parts {
		switch p := part.(type) {
		case string:
			where.WriteString(p)
		case predicate.Expression:
			where.WriteString(p.GetSpecification())
			s.addParams(p.GetValues()...)
		default:
			return "", errors.New("error")
		}
	}
	return where.String(), nil
}

func (s *baseSelect) processGroup() (string, error) {
	selectGroup := s.group
	if selectGroup == nil || len(selectGroup) == 0 {
		return "", nil
	}
	var str = getStrB()
	defer putStrB(str)
	str.WriteString("GROUP BY ")
	for i, group := range selectGroup {
		if i > 0 {
			str.WriteString(", ")
		}
		str.WriteString(predicate.QuoteIdentifier(group))
	}
	return str.String(), nil
}

func (s *baseSelect) processHaving() (string, error) {
	having := s.having
	if having == nil || having.count() == 0 {
		return "", nil
	}
	var (
		parts, _ = having.GetExpressionData()
		where    = getStrB()
	)
	defer putStrB(where)
	where.WriteString("HAVING ")
	for _, part := range parts {
		switch p := part.(type) {
		case string:
			where.WriteString(p)
		case predicate.Expression:
			where.WriteString(p.GetSpecification())
			s.addParams(p.GetValues()...)
		default:
			return "", errors.New("error")
		}
	}
	return where.String(), nil
}

func (s *baseSelect) processOrder() (string, error) {
	selectOrder := s.order
	if selectOrder == nil || len(selectOrder) == 0 {
		return "", nil
	}
	var str = getStrB()
	defer putStrB(str)
	str.WriteString("ORDER BY ")
	for i, order := range selectOrder {
		if i > 0 {
			str.WriteString(", ")
		}
		o := strings.Split(order, " ")
		if len(o) != 2 {
			return "", errors.New("order error")
		}
		str.WriteString(predicate.QuoteIdentifier(o[0]))
		str.WriteString(" ")
		str.WriteString(strings.Trim(o[1], " "))
	}
	return str.String(), nil
}

func (s *baseSelect) processLimit() (string, error) {
	limit := s.limit
	if limit < 0 {
		return "", nil
	}
	s.addParams(limit)
	return "LIMIT ?", nil
}

func (s *baseSelect) processOffset() (string, error) {
	offset := s.offset
	if offset < 0 {
		return "", nil
	}
	s.addParams(offset)
	return "OFFSET ?", nil
}

func (s *baseSelect) processTail() (string, error) {
	return s.tail, nil
}

func (u *baseUpdate) processUpdate() (string, error) {
	var str = getStrB()
	defer putStrB(str)
	str.WriteString("UPDATE ")
	str.WriteString(predicate.QuoteIdentifier(u.table))
	return str.String(), nil
}

func (u *baseUpdate) processJoins() (string, error) {
	join := u.join
	if join == nil || join.count() == 0 {
		return "", nil
	}
	var str = getStrB()
	defer putStrB(str)
	for _, joinAttr := range join.GetJoins() {
		if str.Len() > 0 {
			str.WriteString(" ")
		}
		str.WriteString(joinAttr.typo)
		str.WriteString(" JOIN ")
		str.WriteString(predicate.QuoteTable(joinAttr.name))
		str.WriteString(" ON ")
		str.WriteString(joinAttr.on)
	}
	return str.String(), nil
}

func (u *baseUpdate) processSet() (string, error) {
	if u.set == nil || len(u.set) == 0 {
		return "", errors.New("process set: empty set")
	}
	var str = getStrB()
	defer putStrB(str)
	str.WriteString("SET ")
	for i, v := range u.set {
		if i > 0 {
			str.WriteString(", ")
		}
		str.WriteString(predicate.QuoteIdentifier(v))
		str.WriteString(" = ?")
	}
	return str.String(), nil
}

func (u *baseUpdate) processWhere() (string, error) {
	if u.where == nil || u.where.count() == 0 {
		return "", nil
	}
	var (
		parts, _ = u.where.GetExpressionData()
		where    = getStrB()
	)
	defer putStrB(where)
	where.WriteString("WHERE ")
	for _, part := range parts {
		switch p := part.(type) {
		case string:
			where.WriteString(p)
		case predicate.Expression:
			where.WriteString(p.GetSpecification())
			u.addParams(p.GetValues()...)
		default:
			return "", errors.New("error")
		}
	}
	return where.String(), nil
}

func (i *baseInsert) processInsert() (string, error) {
	columns := i.columns
	var str = getStrB()
	defer putStrB(str)
	str.WriteString("INSERT INTO ")
	str.WriteString(predicate.QuoteTable(i.table))
	str.WriteString("(")
	for c, v := range columns {
		if c > 0 {
			str.WriteString(", ")
		}
		str.WriteString(predicate.QuoteIdentifier(v))
	}
	str.WriteString(") VALUES")
	for j := 0; j < len(i.params); j += len(columns) {
		if j > 0 {
			str.WriteString(", ")
		}
		str.WriteString("(")
		str.WriteString(strings.Repeat(", ?", len(columns))[2:])
		str.WriteString(")")
	}
	return str.String(), nil
}

func (d *baseDelete) processDelete() (string, error) {
	str := getStrB()
	defer putStrB(str)
	str.WriteString("DELETE FROM ")
	str.WriteString(predicate.QuoteTable(d.table))
	return str.String(), nil
}

func (d *baseDelete) processWhere() (string, error) {
	if d.where == nil || d.where.count() == 0 {
		return "", nil
	}
	var (
		parts, _ = d.where.GetExpressionData()
		where    = getStrB()
	)
	defer putStrB(where)
	where.WriteString("WHERE ")
	for _, part := range parts {
		switch p := part.(type) {
		case string:
			where.WriteString(p)
		case predicate.Expression:
			where.WriteString(p.GetSpecification())
			d.addParams(p.GetValues()...)
		default:
			return "", errors.New("error")
		}
	}
	return where.String(), nil
}
