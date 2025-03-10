package main

import "fmt"

// ListBoolExpr holds parameters for aggregating events.
type ListBoolExpr struct {
	Type       string          `json:"type"`
	Parameters AggregateParams `json:"parameters"`
}

func (c ListBoolExpr) String() string {
	return fmt.Sprintf("ListBoolExpr{Type: %s, Parameters: %v}", c.Type, c.Parameters)
}
