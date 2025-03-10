package main

import "fmt"

// AggregateParams contains the filter, grouping, and inner conditions.
type AggregateParams struct {
	Filter     []BooleanExpression `json:"filter"`
	GroupBy    []GroupBy           `json:"group_by"`
	Conditions []BooleanExpression `json:"conditions"`
}

// GroupBy represents a single grouping instruction.
type GroupBy struct {
	JsonPath  string `json:"jsonPath"`
	Transform string `json:"transform,omitempty"`
}

func (a AggregateParams) String() string {
	return fmt.Sprintf("AggregateParams{Filter: %v, GroupBy: %v, Conditions: %v}", a.Filter, a.GroupBy, a.Conditions)
}
