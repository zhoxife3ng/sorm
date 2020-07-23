package predicate

import (
	"errors"
	"fmt"
)

type In struct {
	identifier string
	values     []interface{}
}

func NewIn(identifier string, values ...interface{}) *In {
	i := &In{}
	if identifier != "" {
		i.SetIdentifier(identifier)
	}
	i.SetValues(values...)
	return i
}

func (i *In) SetIdentifier(identifier string) *In {
	i.identifier = identifier
	return i
}

func (i *In) SetValues(values ...interface{}) *In {
	var v = make([]interface{}, 0)
	for _, value := range values {
		if val, ok := value.([]interface{}); ok {
			v = append(v, val...)
		} else {
			v = append(v, value)
		}
	}
	i.values = v
	return i
}

func (i *In) GetIdentifier() string {
	return i.identifier
}

func (i *In) GetValues() []interface{} {
	return i.values
}

func (i *In) GetExpressionData() ([]interface{}, error) {
	var inData string
	for j := 0; j < len(i.GetValues()); j++ {
		inData += ", " + PlaceHolder
	}
	if inData == "" {
		return nil, errors.New("in: error values")
	}
	return []interface{}{
		NewExpression(
			fmt.Sprintf("%s IN (%s)", QuoteIdentifier(i.GetIdentifier()), inData[2:]),
			i.GetValues()...,
		),
	}, nil
}
