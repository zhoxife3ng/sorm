package predicate

import "fmt"

type IsNotNull struct {
	identifier string
}

func NewIsNotNull(identifier string) *IsNotNull {
	i := &IsNotNull{}
	if identifier != "" {
		i.SetIdentifier(identifier)
	}
	return i
}

func (i *IsNotNull) SetIdentifier(identifier string) *IsNotNull {
	i.identifier = identifier
	return i
}

func (i *IsNotNull) GetIdentifier() string {
	return i.identifier
}

func (i *IsNotNull) GetExpressionData() ([]interface{}, error) {
	return []interface{}{
		fmt.Sprintf("%s IS NOT NULL", QuoteIdentifier(i.GetIdentifier())),
	}, nil
}
