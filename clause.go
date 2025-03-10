package main

import "fmt"

// Clause represents a single comparison clause inside a simple condition.
type Clause struct {
	Source     string        `json:"source"`
	Operator   string        `json:"operator"`
	Parameters []interface{} `json:"parameters"`
}

func (c Clause) String() string {
	return fmt.Sprintf("Clause{Source: %s, Operator: %s, Parameters: %v}", c.Source, c.Operator, c.Parameters)
}

// EvaluateClause applies the clause's operator to its evaluated parameters.
func (clause Clause) Evaluate(dataContext map[string]interface{}) (bool, error) {
	// Retrieve the source data from the context.
	sourceData, exists := dataContext[clause.Source]
	if !exists {
		return false, fmt.Errorf("source %s not found in context", clause.Source)
	}
	if len(clause.Parameters) < 2 {
		return false, fmt.Errorf("expected at least 2 parameters for operator %s", clause.Operator)
	}
	// Evaluate the first two parameters.
	left, err := evaluateParameter(clause.Parameters[0], sourceData)
	if err != nil {
		return false, err
	}
	right, err := evaluateParameter(clause.Parameters[1], sourceData)
	if err != nil {
		return false, err
	}
	return clause.execute(left, right)
}

// execute applies the specified operator to the left and right values.
func (c Clause) execute(left, right interface{}) (bool, error) {
	switch c.Operator {
	case "equals":
		return isEqual(left, right), nil
	case "not_equals":
		return !isEqual(left, right), nil
	case "greater_than", "greater_than_eq", "less_than", "less_than_eq":
		leftFloat, ok1 := toFloat(left)
		rightFloat, ok2 := toFloat(right)
		if !ok1 || !ok2 {
			return false, fmt.Errorf("cannot convert values to float for operator %s", c.Operator)
		}
		switch c.Operator {
		case "greater_than":
			return leftFloat > rightFloat, nil
		case "greater_than_eq":
			return leftFloat >= rightFloat, nil
		case "less_than":
			return leftFloat < rightFloat, nil
		case "less_than_eq":
			return leftFloat <= rightFloat, nil
		}
	case "in":
		// Expect the right parameter to be an array.
		if slice, ok := right.([]interface{}); ok {
			for _, item := range slice {
				if isEqual(left, item) {
					return true, nil
				}
			}
			return false, nil
		}
		return false, fmt.Errorf("c.Operator 'in' expects an array for the right-hand parameter")
	case "not_in":
		if slice, ok := right.([]interface{}); ok {
			for _, item := range slice {
				if isEqual(left, item) {
					return false, nil
				}
			}
			return true, nil
		}
		return false, fmt.Errorf("c.Operator 'not_in' expects an array for the right-hand parameter")
	case "same":
		return isEqual(left, right), nil
	default:
		return false, fmt.Errorf("unsupported c.Operator: %s", c.Operator)
	}
	return false, nil
}
