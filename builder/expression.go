package builder

const PlaceHolder = "?"

type Expression struct {
	specification string
	values        []interface{}
	err           error
}

func NewExpression(specification string, values ...interface{}) *Expression {
	return &Expression{
		specification: specification,
		values:        values,
	}
}

func ErrExpression(err error) *Expression {
	return &Expression{
		err: err,
	}
}

func (e *Expression) GetSpecification() string {
	return e.specification
}

func (e *Expression) GetValues() []interface{} {
	return e.values
}

func (e *Expression) GetExpressionData() ([]interface{}, error) {
	return []interface{}{e}, e.err
}
