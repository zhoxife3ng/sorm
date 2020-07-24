package builder

import (
	"strings"
)

type clause struct {
	predicate *Predicate
}

func buildPredicate(where interface{}, values ...interface{}) *Predicate {
	var (
		wherePredicate = NewPredicate()
		combination    = defaultCombination
	)
	switch w := where.(type) {
	case string:
		var predicateIfe Predicator
		if strings.Contains(w, PlaceHolder) {
			if strings.Count(w, PlaceHolder) != len(values) {
				predicateIfe = ErrExpression(ErrBuildPlaceHolder)
			} else {
				predicateIfe = NewExpression(w, values...)
			}
		} else if len(values) > 0 {
			if len(values) == 1 {
				if values[0] == nil {
					if strings.Index(w, "!") == 0 {
						predicateIfe = isNotNull(w[1:])
					} else {
						predicateIfe = isNull(w)
					}
				} else {
					if strings.Index(w, "!") == 0 {
						predicateIfe = operate(w[1:], OpNe, values[0])
					} else {
						predicateIfe = operate(w, OpEq, values[0])
					}
				}
			} else {
				if strings.Index(w, "!") == 0 {
					predicateIfe = notIn(w[1:], values...)
				} else {
					predicateIfe = in(w, values...)
				}
			}
		}
		wherePredicate.AddPredicate(predicateIfe, combination)
	case map[string]interface{}:
		for key, value := range w {
			var predicateIfe Predicator
			if strings.Contains(key, PlaceHolder) {
				if v, ok := value.([]interface{}); ok {
					if strings.Count(key, PlaceHolder) != len(v) {
						predicateIfe = ErrExpression(ErrBuildPlaceHolder)
					} else {
						predicateIfe = NewExpression(key, v...)
					}
				} else {
					predicateIfe = NewExpression(key, value)
				}
			} else if value == nil {
				if strings.Index(key, "!") == 0 {
					predicateIfe = isNotNull(key[1:])
				} else {
					predicateIfe = isNull(key)
				}
			} else if v, ok := value.([]interface{}); ok {
				if strings.Index(key, "!") == 0 {
					predicateIfe = notIn(key[1:], v...)
				} else {
					predicateIfe = in(key, v...)
				}
			} else {
				if strings.Index(key, "!") == 0 {
					predicateIfe = operate(key[1:], OpNe, value)
				} else {
					predicateIfe = operate(key, OpEq, value)
				}
			}
			wherePredicate.AddPredicate(predicateIfe, combination)
		}
	case Predicator:
		if w != nil {
			wherePredicate.AddPredicate(w, combination)
		}
	case *clause:
		if w != nil {
			wherePredicate.AddPredicate(w.predicate, combination)
		}
	}
	return wherePredicate
}

func Clause(where interface{}, values ...interface{}) *clause {
	return &clause{buildPredicate(where, values...)}
}

func EmptyClause() *clause {
	return &clause{buildPredicate(nil)}
}

func (c *clause) Or(where interface{}, values ...interface{}) *clause {
	c.predicate.AddPredicate(buildPredicate(where, values...), CombinedByOr)
	return c
}

func (c *clause) And(where interface{}, values ...interface{}) *clause {
	c.predicate.AddPredicate(buildPredicate(where, values...), CombinedByAnd)
	return c
}

func (c *clause) EqualTo(identifier string, value interface{}) *clause {
	c.predicate.AddPredicate(operate(identifier, OpEq, value), defaultCombination)
	return c
}

func (c *clause) NotEqualTo(identifier string, value interface{}) *clause {
	c.predicate.AddPredicate(operate(identifier, OpNe, value), defaultCombination)
	return c
}

func (c *clause) Like(identifier, value string) *clause {
	c.predicate.AddPredicate(operate(identifier, OpLike, value), defaultCombination)
	return c
}

func (c *clause) NotLike(identifier, value string) *clause {
	c.predicate.AddPredicate(operate(identifier, OpNotLike, value), defaultCombination)
	return c
}

func (c *clause) LessThan(identifier string, value interface{}) *clause {
	c.predicate.AddPredicate(operate(identifier, OpLt, value), defaultCombination)
	return c
}

func (c *clause) LessThanOrEqualTo(identifier string, value interface{}) *clause {
	c.predicate.AddPredicate(operate(identifier, OpLte, value), defaultCombination)
	return c
}

func (c *clause) GreaterThan(identifier string, value interface{}) *clause {
	c.predicate.AddPredicate(operate(identifier, OpGt, value), defaultCombination)
	return c
}

func (c *clause) GreaterThanOrEqualTo(identifier string, value interface{}) *clause {
	c.predicate.AddPredicate(operate(identifier, OpGte, value), defaultCombination)
	return c
}

func (c *clause) Between(identifier string, minValue, maxValue interface{}) *clause {
	c.predicate.AddPredicate(between(identifier, minValue, maxValue), defaultCombination)
	return c
}

func (c *clause) NotBetween(identifier string, minValue, maxValue interface{}) *clause {
	c.predicate.AddPredicate(notBetween(identifier, minValue, maxValue), defaultCombination)
	return c
}

func (c *clause) Exists(specification string, values ...interface{}) *clause {
	c.predicate.AddPredicate(exists(specification, values...), defaultCombination)
	return c
}

func (c *clause) NotExists(specification string, values ...interface{}) *clause {
	c.predicate.AddPredicate(notExists(specification, values...), defaultCombination)
	return c
}

func (c *clause) In(identifier string, values ...interface{}) *clause {
	c.predicate.AddPredicate(in(identifier, values...), defaultCombination)
	return c
}

func (c *clause) NotIn(identifier string, values ...interface{}) *clause {
	c.predicate.AddPredicate(notIn(identifier, values...), defaultCombination)
	return c
}

func (c *clause) IsNull(identifier string) *clause {
	c.predicate.AddPredicate(isNull(identifier), defaultCombination)
	return c
}

func (c *clause) IsNotNull(identifier string) *clause {
	c.predicate.AddPredicate(isNotNull(identifier), defaultCombination)
	return c
}
