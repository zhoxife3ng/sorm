package predicate

import (
	"errors"
	"fmt"
	"strings"
)

type NotExists struct {
	specification string
	values        []interface{}
}

func NewNotExists(specification string, values ...interface{}) *NotExists {
	e := &NotExists{}
	if specification != "" {
		e.SetSpecification(specification)
	}
	e.SetValues(values...)
	return e
}

func (e *NotExists) SetSpecification(specification string) *NotExists {
	e.specification = specification
	return e
}

func (e *NotExists) SetValues(values ...interface{}) *NotExists {
	e.values = values
	return e
}

func (e *NotExists) GetSpecification() string {
	return e.specification
}

func (e *NotExists) GetValues() []interface{} {
	return e.values
}

func (e *NotExists) GetExpressionData() ([]interface{}, error) {
	placeHolderCount := strings.Count(e.GetSpecification(), PlaceHolder)
	if placeHolderCount > len(e.values) {
		return nil, errors.New("exists: error value num")
	}
	return []interface{}{
		NewExpression(
			fmt.Sprintf("NOT EXISTS (%s)", e.GetSpecification()),
			e.GetValues()[:placeHolderCount]...,
		),
	}, nil
}
