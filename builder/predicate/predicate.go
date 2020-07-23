package predicate

import (
	"bytes"
	"strings"
)

func QuoteTable(table string) string {
	tableName, aliasTableName := ResolveIdentifier(table)
	str := strings.Builder{}
	str.WriteString(QuoteIdentifier(tableName))
	if tableName != aliasTableName {
		str.WriteString(" AS ")
		str.WriteString(QuoteIdentifier(aliasTableName))
	}
	return str.String()
}

func ResolveIdentifier(identifier string) (name string, alias string) {
	var s []string
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
	name, alias := ResolveIdentifier(identifier)
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

func QuoteValue(value string) string {
	return "\"" + value + "\""
}
