package builder

import (
	"bytes"
	"strings"
)

func (s *Selector) processSelect() (string, error) {
	var (
		selectQuantifier = s.quantifier
		selectTable      = s.table
		selectColumns    = s.columns
		selectJoin       = s.join
	)

	var str = getStrBuilder()
	defer putStrBuilder(str)
	str.WriteString("SELECT ")
	if selectQuantifier != "" {
		str.WriteString(selectQuantifier)
		str.WriteString(" ")
	}

	var columnStr = bytes.Buffer{}
	_, aliasTableName := resolveIdentifier(selectTable)
	// column
	for i, c := range selectColumns {
		if i > 0 {
			columnStr.WriteString(", ")
		}
		if !strings.Contains(c, ".") {
			columnStr.WriteString(QuoteTable(aliasTableName))
			columnStr.WriteString(".")
			columnStr.WriteString(QuoteIdentifier(c))
		} else {
			columnStr.WriteString(QuoteIdentifier(c))
		}
	}

	// join column
	if selectJoin != nil {
		for _, joinAttr := range selectJoin.GetJoins() {
			_, aliasTableName := resolveIdentifier(joinAttr.name)
			for _, c := range joinAttr.columns {
				if columnStr.Len() > 0 {
					columnStr.WriteString(", ")
				}
				if !strings.Contains(c, ".") {
					columnStr.WriteString(QuoteTable(aliasTableName))
					columnStr.WriteString(".")
					columnStr.WriteString(QuoteIdentifier(c))
				} else {
					columnStr.WriteString(QuoteIdentifier(c))
				}
			}
		}
	}

	// func column
	for fColumns, alias := range s.fColumns {
		if columnStr.Len() > 0 {
			columnStr.WriteString(", ")
		}
		columnStr.WriteString(fColumns)
		columnStr.WriteString(" AS ")
		columnStr.WriteString(QuoteIdentifier(alias))
	}

	if columnStr.Len() == 0 {
		columnStr.WriteString(QuoteTable(aliasTableName))
		columnStr.WriteString(".*")
	}

	str.Write(columnStr.Bytes())

	if selectTable != "" {
		str.WriteString(" FROM ")
		str.WriteString(QuoteTable(selectTable))
	}
	return str.String(), nil
}

func (s *Selector) processForceIndex() (string, error) {
	if s.forceIndex == "" {
		return "", nil
	}
	var str = getStrBuilder()
	defer putStrBuilder(str)
	str.WriteString("FORCE INDEX (")
	str.WriteString(QuoteIdentifier(s.forceIndex))
	str.WriteString(")")
	return str.String(), nil
}

func (s *Selector) processJoins() (string, error) {
	join := s.join
	if join == nil || join.count() == 0 {
		return "", nil
	}
	var str = getStrBuilder()
	defer putStrBuilder(str)
	for _, joinAttr := range join.GetJoins() {
		if str.Len() > 0 {
			str.WriteString(" ")
		}
		str.WriteString(joinAttr.typo)
		str.WriteString(" JOIN ")
		str.WriteString(QuoteTable(joinAttr.name))
		str.WriteString(" ON ")
		str.WriteString(joinAttr.on)
	}
	return str.String(), nil
}

func (s *Selector) processWhere() (string, error) {
	selectWhere := s.where
	if selectWhere == nil || selectWhere.count() == 0 {
		return "", nil
	}
	var parts, err = selectWhere.GetExpressionData()
	if err != nil {
		return "", err
	}
	var where = getStrBuilder()
	defer putStrBuilder(where)
	where.WriteString("WHERE ")
	for _, part := range parts {
		switch p := part.(type) {
		case string:
			where.WriteString(p)
		case *Expression:
			where.WriteString(p.GetSpecification())
			s.addParams(p.GetValues()...)
		default:
			return "", ErrNotSupportProcess
		}
	}
	return where.String(), nil
}

func (s *Selector) processGroup() (string, error) {
	selectGroup := s.group
	if selectGroup == nil || len(selectGroup) == 0 {
		return "", nil
	}
	var str = getStrBuilder()
	defer putStrBuilder(str)
	str.WriteString("GROUP BY ")
	for i, group := range selectGroup {
		if i > 0 {
			str.WriteString(", ")
		}
		str.WriteString(QuoteIdentifier(group))
	}
	return str.String(), nil
}

func (s *Selector) processHaving() (string, error) {
	having := s.having
	if having == nil || having.count() == 0 {
		return "", nil
	}
	var parts, err = having.GetExpressionData()
	if err != nil {
		return "", err
	}
	var where = getStrBuilder()
	defer putStrBuilder(where)
	where.WriteString("HAVING ")
	for _, part := range parts {
		switch p := part.(type) {
		case string:
			where.WriteString(p)
		case *Expression:
			where.WriteString(p.GetSpecification())
			s.addParams(p.GetValues()...)
		default:
			return "", ErrNotSupportProcess
		}
	}
	return where.String(), nil
}

func (s *Selector) processOrder() (string, error) {
	selectOrder := s.order
	if selectOrder == nil || len(selectOrder) == 0 {
		return "", nil
	}
	var str = getStrBuilder()
	defer putStrBuilder(str)
	str.WriteString("ORDER BY ")
	for i, order := range selectOrder {
		if i > 0 {
			str.WriteString(", ")
		}
		o := strings.Split(strings.Trim(order, " "), " ")
		if len(o) > 0 && len(o) < 3 {
			str.WriteString(QuoteIdentifier(o[0]))
			str.WriteString(" ")
			var sort = "ASC"
			if len(o) == 2 && o[1] != "" {
				sort = strings.ToUpper(strings.Trim(o[1], " "))
			}
			if sort == "ASC" || sort == "DESC" {
				str.WriteString(sort)
				continue
			}
		}
		return "", ErrProcessOrder
	}
	return str.String(), nil
}

func (s *Selector) processLimit() (string, error) {
	limit := s.limit
	if limit < 0 {
		return "", nil
	}
	s.addParams(limit)
	return "LIMIT ?", nil
}

func (s *Selector) processOffset() (string, error) {
	offset := s.offset
	if offset < 0 {
		return "", nil
	}
	s.addParams(offset)
	return "OFFSET ?", nil
}

func (s *Selector) processTail() (string, error) {
	return s.tail, nil
}

func (u *Updater) processUpdate() (string, error) {
	var str = getStrBuilder()
	defer putStrBuilder(str)
	str.WriteString("UPDATE ")
	str.WriteString(QuoteTable(u.table))
	return str.String(), nil
}

func (u *Updater) processJoins() (string, error) {
	join := u.join
	if join == nil || join.count() == 0 {
		return "", nil
	}
	var str = getStrBuilder()
	defer putStrBuilder(str)
	for _, joinAttr := range join.GetJoins() {
		if str.Len() > 0 {
			str.WriteString(" ")
		}
		str.WriteString(joinAttr.typo)
		str.WriteString(" JOIN ")
		str.WriteString(QuoteTable(joinAttr.name))
		str.WriteString(" ON ")
		str.WriteString(joinAttr.on)
	}
	return str.String(), nil
}

func (u *Updater) processSet() (string, error) {
	if u.set == nil || len(u.set) == 0 {
		return "", ErrProcessSet
	}
	var str = getStrBuilder()
	defer putStrBuilder(str)
	str.WriteString("SET ")
	for i, v := range u.set {
		if i > 0 {
			str.WriteString(", ")
		}
		if strings.Contains(v, PlaceHolder) {
			str.WriteString(v)
		} else {
			str.WriteString(QuoteIdentifier(v))
			str.WriteString("=?")
		}
	}
	return str.String(), nil
}

func (u *Updater) processWhere() (string, error) {
	if u.where == nil || u.where.count() == 0 {
		return "", nil
	}
	var parts, err = u.where.GetExpressionData()
	if err != nil {
		return "", err
	}
	var where = getStrBuilder()
	defer putStrBuilder(where)
	where.WriteString("WHERE ")
	for _, part := range parts {
		switch p := part.(type) {
		case string:
			where.WriteString(p)
		case *Expression:
			where.WriteString(p.GetSpecification())
			u.addParams(p.GetValues()...)
		default:
			return "", ErrNotSupportProcess
		}
	}
	return where.String(), nil
}

func (i *Inserter) processInsert() (string, error) {
	columns := i.columns
	var str = getStrBuilder()
	defer putStrBuilder(str)
	str.WriteString("INSERT INTO ")
	str.WriteString(QuoteTable(i.table))
	str.WriteString("(")
	for c, v := range columns {
		if c > 0 {
			str.WriteString(", ")
		}
		str.WriteString(QuoteIdentifier(v))
	}
	str.WriteString(") VALUES")
	for j := 0; j < len(i.params); j += 1 {
		if j > 0 {
			str.WriteString(", ")
		}
		str.WriteString("(")
		str.WriteString(strings.Repeat(",?", len(columns))[1:])
		str.WriteString(")")
	}
	return str.String(), nil
}

func (d *Deleter) processDelete() (string, error) {
	str := getStrBuilder()
	defer putStrBuilder(str)
	str.WriteString("DELETE FROM ")
	str.WriteString(QuoteTable(d.table))
	return str.String(), nil
}

func (d *Deleter) processWhere() (string, error) {
	if d.where == nil || d.where.count() == 0 {
		return "", nil
	}
	var parts, err = d.where.GetExpressionData()
	if err != nil {
		return "", err
	}
	var where = getStrBuilder()
	defer putStrBuilder(where)
	where.WriteString("WHERE ")
	for _, part := range parts {
		switch p := part.(type) {
		case string:
			where.WriteString(p)
		case *Expression:
			where.WriteString(p.GetSpecification())
			d.addParams(p.GetValues()...)
		default:
			return "", ErrNotSupportProcess
		}
	}
	return where.String(), nil
}
