package predicate

import "fmt"

type IsNull struct {
	identifier string
}

func NewIsNull(identifier string) *IsNull {
	i := &IsNull{}
	if identifier != "" {
		i.SetIdentifier(identifier)
	}
	return i
}

func (i *IsNull) SetIdentifier(identifier string) *IsNull {
	i.identifier = identifier
	return i
}

func (i *IsNull) GetIdentifier() string {
	return i.identifier
}

func (i *IsNull) GetExpressionData() ([]interface{}, error) {
	return []interface{}{
		fmt.Sprintf("%s IS NULL", QuoteIdentifier(i.GetIdentifier())),
	}, nil
}
