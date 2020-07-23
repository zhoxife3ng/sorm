package predicate

import (
	"fmt"
)

const (
	OpEq      = "="
	OpNe      = "!="
	OpLt      = "<"
	OpLte     = "<="
	OpGt      = ">"
	OpGte     = ">="
	OpLike    = "LIKE"
	OpNotLike = "NOT LIKE"
)

type Operator struct {
	left     string
	right    interface{}
	operator string
}

func NewOperator(left, operator string, right interface{}) *Operator {
	o := &Operator{}
	if left != "" {
		o.SetLeft(left)
	}
	if right != "" {
		o.SetRight(right)
	}
	if operator != "" {
		o.SetOperator(operator)
	}
	return o
}

func (o *Operator) SetLeft(left string) *Operator {
	o.left = left
	return o
}

func (o *Operator) SetRight(right interface{}) *Operator {
	o.right = right
	return o
}

func (o *Operator) SetOperator(operator string) *Operator {
	o.operator = operator
	return o
}

func (o *Operator) GetLeft() string {
	return o.left
}

func (o *Operator) GetOperator() string {
	return o.operator
}

func (o *Operator) GetRight() interface{} {
	return o.right
}

func (o *Operator) GetExpressionData() ([]interface{}, error) {
	return []interface{}{
		NewExpression(
			fmt.Sprintf("%s %s ?", QuoteIdentifier(o.GetLeft()), o.GetOperator()),
			o.GetRight(),
		),
	}, nil
}
