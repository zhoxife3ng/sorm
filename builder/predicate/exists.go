package predicate

import (
	"errors"
	"fmt"
	"strings"
)

type Exists struct {
	specification string
	values        []interface{}
}

func NewExists(specification string, values ...interface{}) *Exists {
	e := &Exists{}
	if specification != "" {
		e.SetSpecification(specification)
	}
	e.SetValues(values...)
	return e
}

func (e *Exists) SetSpecification(specification string) *Exists {
	e.specification = specification
	return e
}

func (e *Exists) SetValues(values ...interface{}) *Exists {
	e.values = values
	return e
}

func (e *Exists) GetSpecification() string {
	return e.specification
}

func (e *Exists) GetValues() []interface{} {
	return e.values
}

func (e *Exists) GetExpressionData() ([]interface{}, error) {
	placeHolderCount := strings.Count(e.GetSpecification(), PlaceHolder)
	if placeHolderCount > len(e.values) {
		return nil, errors.New("exists: error value num")
	}
	return []interface{}{
		NewExpression(
			fmt.Sprintf("EXISTS (%s)", e.GetSpecification()),
			e.GetValues()[:placeHolderCount]...,
		),
	}, nil
}
