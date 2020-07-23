package builder

import (
	"github.com/x554462/sorm/builder/predicate"
	"strings"
)

type clause struct {
	predicate *Predicate
}

func buildPredicate(where interface{}, values ...interface{}) *Predicate {
	var (
		wherePredicate = NewPredicate()
		combination    = CombinedByAnd
	)
	switch w := where.(type) {
	case string:
		var predicateIfe Predicator
		if strings.Contains(w, predicate.PlaceHolder) {
			predicateIfe = predicate.NewExpression(w, values...)
		} else if len(values) > 0 {
			if len(values) == 1 {
				if values[0] == nil {
					if strings.Index(w, "!") == 0 {
						predicateIfe = predicate.NewIsNotNull(w[1:])
					} else {
						predicateIfe = predicate.NewIsNull(w)
					}
				} else {
					predicateIfe = predicate.NewOperator(w, predicate.OpEq, values[0])
				}
			} else {
				if strings.Index(w, "!") == 0 {
					predicateIfe = predicate.NewNotIn(w[1:], values...)
				} else {
					predicateIfe = predicate.NewIn(w, values...)
				}
			}
		}
		wherePredicate.AddPredicate(predicateIfe, combination)
	case map[string]interface{}:
		for key, value := range w {
			var predicateIfe Predicator
			if strings.Contains(key, predicate.PlaceHolder) {
				if v, ok := value.([]interface{}); ok {
					predicateIfe = predicate.NewExpression(key, v...)
				} else {
					predicateIfe = predicate.NewExpression(key, value)
				}
			} else if value == nil {
				if strings.Index(key, "!") == 0 {
					predicateIfe = predicate.NewIsNotNull(key[1:])
				} else {
					predicateIfe = predicate.NewIsNull(key)
				}
			} else if v, ok := value.([]interface{}); ok {
				if strings.Index(key, "!") == 0 {
					predicateIfe = predicate.NewNotIn(key[1:], v...)
				} else {
					predicateIfe = predicate.NewIn(key, v...)
				}
			} else {
				if strings.Index(key, "!") == 0 {
					predicateIfe = predicate.NewOperator(key[1:], predicate.OpNe, value)
				} else {
					predicateIfe = predicate.NewOperator(key, predicate.OpEq, value)
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

func (w *clause) Or(where interface{}, values ...interface{}) *clause {
	w.predicate.AddPredicate(buildPredicate(where, values...), CombinedByOr)
	return w
}

func (w *clause) And(where interface{}, values ...interface{}) *clause {
	w.predicate.AddPredicate(buildPredicate(where, values...), CombinedByAnd)
	return w
}
