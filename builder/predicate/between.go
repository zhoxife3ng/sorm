package predicate

import "fmt"

type Between struct {
	identifier         string
	minValue, maxValue interface{}
}

func NewBetween(identifier string, minValue, maxValue interface{}) *Between {
	b := &Between{}
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

func (b *Between) SetIdentifier(identifier string) *Between {
	b.identifier = identifier
	return b
}

func (b *Between) SetMinValue(minValue interface{}) *Between {
	b.minValue = minValue
	return b
}

func (b *Between) SetMaxValue(maxValue interface{}) *Between {
	b.maxValue = maxValue
	return b
}

func (b *Between) GetIdentifier() string {
	return b.identifier
}

func (b *Between) GetMinValue() interface{} {
	return b.minValue
}

func (b *Between) GetMaxValue() interface{} {
	return b.maxValue
}

func (b *Between) GetExpressionData() ([]interface{}, error) {
	return []interface{}{
		NewExpression(
			fmt.Sprintf("%s BETWEEN ? AND ?", QuoteIdentifier(b.GetIdentifier())),
			b.GetMinValue(), b.GetMaxValue(),
		),
	}, nil
}
