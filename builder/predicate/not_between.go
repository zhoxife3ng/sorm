package predicate

import "fmt"

type NotBetween struct {
	identifier         string
	minValue, maxValue interface{}
}

func NewNotBetween(identifier string, minValue, maxValue interface{}) *NotBetween {
	b := &NotBetween{}
	if identifier != "" {
		b.SetIdentifier(identifier)
	}
	if minValue != nil {
		b.SetMinValue(minValue)
	}
	if maxValue != nil {
		b.SetMaxValue(maxValue)
	}
	return b
}

func (b *NotBetween) SetIdentifier(identifier string) *NotBetween {
	b.identifier = identifier
	return b
}

func (b *NotBetween) SetMinValue(minValue interface{}) *NotBetween {
	b.minValue = minValue
	return b
}

func (b *NotBetween) SetMaxValue(maxValue interface{}) *NotBetween {
	b.maxValue = maxValue
	return b
}

func (b *NotBetween) GetIdentifier() string {
	return b.identifier
}

func (b *NotBetween) GetMinValue() interface{} {
	return b.minValue
}

func (b *NotBetween) GetMaxValue() interface{} {
	return b.maxValue
}

func (b *NotBetween) GetExpressionData() ([]interface{}, error) {
	return []interface{}{
		NewExpression(
			fmt.Sprintf("%s NOT BETWEEN ? AND ?", QuoteIdentifier(b.GetIdentifier())),
			b.GetMinValue(), b.GetMaxValue(),
		),
	}, nil
}
