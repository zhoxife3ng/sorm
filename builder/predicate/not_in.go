package predicate

import (
	"errors"
	"fmt"
)

type NotIn struct {
	identifier string
	values     []interface{}
}

func NewNotIn(identifier string, values ...interface{}) *NotIn {
	i := &NotIn{}
	if identifier != "" {
		i.SetIdentifier(identifier)
	}
	i.SetValues(values...)
	return i
}

func (i *NotIn) SetIdentifier(identifier string) *NotIn {
	i.identifier = identifier
	return i
}

func (i *NotIn) SetValues(values ...interface{}) *NotIn {
	var v = make([]interface{}, 0)
	for _, value := range values {
		if val, ok := value.([]interface{}); ok {
			v = append(v, val...)
		}
	}
	i.values = v
	return i
}

func (i *NotIn) GetIdentifier() string {
	return i.identifier
}

func (i *NotIn) GetValues() []interface{} {
	return i.values
}

func (i *NotIn) GetExpressionData() ([]interface{}, error) {
	var inData string
	for j := 0; j < len(i.GetValues()); j++ {
		inData += "," + PlaceHolder
	}
	if inData == "" {
		return nil, errors.New("not in: error values")
	}
	return []interface{}{
		NewExpression(
			fmt.Sprintf("%s NOT IN (%s)", QuoteIdentifier(i.GetIdentifier()), inData[1:]),
			i.GetValues()...,
		),
	}, nil
}
