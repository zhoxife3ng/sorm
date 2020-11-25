package builder

import (
	"bytes"
	"strings"
)

const (
	OpEq      = "="
	OpNe      = "!="
	OpLt      = "<"
	OpLte     = "<="
	OpGt      = ">"
	OpGte     = ">="
	OpLike    = " LIKE "
	OpNotLike = " NOT LIKE "
)

func QuoteTable(table string) string {
	tableName, aliasTableName := resolveIdentifier(table)
	var str = getStrBuilder()
	defer putStrBuilder(str)
	str.WriteString(QuoteIdentifier(tableName))
	if tableName != aliasTableName {
		str.WriteString(" AS ")
		str.WriteString(QuoteIdentifier(aliasTableName))
	}
	return str.String()
}
func resolveIdentifier(identifier string) (name string, alias string) {
	var s []string
	identifier = strings.Trim(identifier, " ")
	lowerTable := strings.ToLower(identifier)
	if strings.Contains(lowerTable, " as ") {
		asIndex := strings.Index(lowerTable, " as ")
		s = []string{
			identifier[:asIndex],
			identifier[asIndex+3:],
		}
	} else if strings.Contains(identifier, " ") {
		s = strings.Split(identifier, " ")
	} else {
		s = []string{
			identifier,
			identifier,
		}
	}
	return strings.Trim(s[0], " "), strings.Trim(s[1], " ")
}

func QuoteIdentifier(identifier string) string {
	name, alias := resolveIdentifier(identifier)
	var ids []string
	if name == alias {
		ids = []string{name}
	} else {
		ids = []string{name, alias}
	}
	var str = strings.Builder{}
	for _, id := range ids {
		if str.Len() > 0 {
			str.WriteString(" AS ")
		}
		bb := bytes.Buffer{}
		for _, v := range strings.Split(id, ".") {
			v = strings.Trim(v, " ")
			if bb.Len() > 0 {
				bb.WriteString(".")
			}
			if v != "*" {
				bb.WriteString("`")
			}
			bb.WriteString(v)
			if v != "*" {
				bb.WriteString("`")
			}
		}
		str.Write(bb.Bytes())
	}
	return str.String()
}

func operate(left, operator string, right interface{}) *Expression {
	var str = getStrBuilder()
	defer putStrBuilder(str)
	str.WriteString(QuoteIdentifier(left))
	str.WriteString(operator)
	str.WriteString("?")
	return NewExpression(str.String(), right)
}

func between(identifier string, minValue, maxValue interface{}) *Expression {
	var str = getStrBuilder()
	defer putStrBuilder(str)
	str.WriteString(QuoteIdentifier(identifier))
	str.WriteString(" BETWEEN ? AND ?")
	return NewExpression(str.String(), minValue, maxValue)
}

func notBetween(identifier string, minValue, maxValue interface{}) *Expression {
	var str = getStrBuilder()
	defer putStrBuilder(str)
	str.WriteString(QuoteIdentifier(identifier))
	str.WriteString(" NOT BETWEEN ? AND ?")
	return NewExpression(str.String(), minValue, maxValue)
}

func exists(specification string, values ...interface{}) *Expression {
	placeHolderCount := strings.Count(specification, PlaceHolder)
	if placeHolderCount > len(values) {
		return ErrExpression(ErrBuildPlaceHolder)
	}
	var str = getStrBuilder()
	defer putStrBuilder(str)
	str.WriteString("EXISTS(")
	str.WriteString(specification)
	str.WriteString(")")
	return NewExpression(str.String(), values...)
}

func notExists(specification string, values ...interface{}) *Expression {
	placeHolderCount := strings.Count(specification, PlaceHolder)
	if placeHolderCount > len(values) {
		return ErrExpression(ErrBuildPlaceHolder)
	}
	var str = getStrBuilder()
	defer putStrBuilder(str)
	str.WriteString("NOT EXISTS(")
	str.WriteString(specification)
	str.WriteString(")")
	return NewExpression(str.String(), values...)
}

func in(identifier string, values ...interface{}) *Expression {
	var str = getStrBuilder()
	defer putStrBuilder(str)
	str.WriteString(QuoteIdentifier(identifier))
	str.WriteString(" IN (")
	for j := 0; j < len(values); j++ {
		if j != 0 {
			str.WriteString(",")
		}
		str.WriteString(PlaceHolder)
	}
	str.WriteString(")")
	return NewExpression(str.String(), values...)
}

func notIn(identifier string, values ...interface{}) *Expression {
	var str = getStrBuilder()
	defer putStrBuilder(str)
	str.WriteString(QuoteIdentifier(identifier))
	str.WriteString(" NOT IN (")
	for j := 0; j < len(values); j++ {
		if j != 0 {
			str.WriteString(",")
		}
		str.WriteString(PlaceHolder)
	}
	str.WriteString(")")
	return NewExpression(str.String(), values...)
}

func isNull(identifier string) *Expression {
	var str = getStrBuilder()
	defer putStrBuilder(str)
	str.WriteString(QuoteIdentifier(identifier))
	str.WriteString(" IS NULL")
	return NewExpression(str.String())
}

func isNotNull(identifier string) *Expression {
	var str = getStrBuilder()
	defer putStrBuilder(str)
	str.WriteString(QuoteIdentifier(identifier))
	str.WriteString(" IS NOT NULL")
	return NewExpression(str.String())
}
