package builder

import (
	"fmt"
	"github.com/x554462/sorm/builder/predicate"
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

func (p *Predicate) EqualTo(left string, right interface{}) *Predicate {
	p.AddPredicate(predicate.NewOperator(left, predicate.OpEq, right), p.nextPredicateCombineOperator)
	return p
}

func (p *Predicate) Between(identifier string, minValue, maxValue interface{}) *Predicate {
	p.AddPredicate(predicate.NewBetween(identifier, minValue, maxValue), p.nextPredicateCombineOperator)
	return p
}

func (p *Predicate) Exists(specification string, values ...interface{}) *Predicate {
	p.AddPredicate(predicate.NewExists(specification, values...), p.nextPredicateCombineOperator)
	return p
}

func (p *Predicate) Or() *Predicate {
	p.nextPredicateCombineOperator = CombinedByOr
	return p
}

func (p *Predicate) And() *Predicate {
	p.nextPredicateCombineOperator = CombinedByAnd
	return p
}

func (p *Predicate) AddPredicate(predicate Predicator, combination string) *Predicate {
	if predicate != nil {
		if specs, err := predicate.GetExpressionData(); err == nil && len(specs) > 0 {
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
