package builder

import (
	"fmt"
)

const (
	CombinedByAnd      = "AND"
	CombinedByOr       = "OR"
	defaultCombination = CombinedByAnd
)

type Predicator interface {
	GetExpressionData() ([]interface{}, error)
}

type predicateExpression struct {
	combination    string
	expressionData Predicator
}

type Predicate struct {
	nextPredicateCombineOperator string
	predicates                   []predicateExpression
}

func NewPredicate() *Predicate {
	return &Predicate{
		nextPredicateCombineOperator: CombinedByAnd,
		predicates:                   make([]predicateExpression, 0),
	}
}

func (p *Predicate) AddPredicate(predicate Predicator, combination string) *Predicate {
	if predicate != nil {
		if specs, _ := predicate.GetExpressionData(); len(specs) > 0 {
			if combination == CombinedByOr {
				p.OrPredicate(predicate)
			} else {
				p.AndPredicate(predicate)
			}
		}
		p.nextPredicateCombineOperator = defaultCombination
	}
	return p
}

func (p *Predicate) AndPredicate(predicate Predicator) *Predicate {
	p.predicates = append(p.predicates, predicateExpression{
		combination:    CombinedByAnd,
		expressionData: predicate,
	})
	return p
}

func (p *Predicate) OrPredicate(predicate Predicator) *Predicate {
	p.predicates = append(p.predicates, predicateExpression{
		combination:    CombinedByOr,
		expressionData: predicate,
	})
	return p
}

func (p *Predicate) GetExpressionData() ([]interface{}, error) {
	var parts = make([]interface{}, 0)
	for i, pe := range p.predicates {
		if i > 0 {
			parts = append(parts, fmt.Sprintf(" %s ", pe.combination))
		}
		p, expIsPredicate := pe.expressionData.(*Predicate)
		if expIsPredicate && p.count() > 1 {
			parts = append(parts, "(")
		}
		ed, err := pe.expressionData.GetExpressionData()
		if err != nil {
			return nil, err
		}
		parts = append(parts, ed...)
		if expIsPredicate && p.count() > 1 {
			parts = append(parts, ")")
		}
	}
	return parts, nil
}

func (p *Predicate) count() int {
	return len(p.predicates)
}
