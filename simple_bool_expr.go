package main

import "fmt"

// SimpleBoolExpr holds a set of clauses evaluated with AND logic.
type SimpleBoolExpr struct {
	Type    string   `json:"type"`
	Clauses []Clause `json:"clauses"`
}

func (c SimpleBoolExpr) String() string {
	return fmt.Sprintf("SimpleBoolExpr{Type: %s, Clauses: %v}", c.Type, c.Clauses)
}

// EvaluateSimpleCondition evaluates all clauses in a simple condition (AND logic).
func EvaluateSimpleCondition(condition SimpleBoolExpr, dataContext map[string]interface{}) (bool, error) {
	for _, clause := range condition.Clauses {
		ok, err := clause.Evaluate(dataContext)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}
