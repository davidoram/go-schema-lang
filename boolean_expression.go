package main

import (
	"encoding/json"
	"errors"
)

// BooleanExpression represents either a SimpleCondition or an AggregateCondition.
type BooleanExpression struct {
	Simple      *SimpleBoolExpr
	Conjunction *ListBoolExpr
}

func (t BooleanExpression) Evaluate(data map[string]interface{}) (bool, error) {
	if t.Simple != nil {
		return EvaluateSimpleCondition(*t.Simple, data)
	} else if t.Conjunction != nil {
		return EvaluateAggregateCondition(*t.Conjunction, data)
	}
	return false, errors.New("unknown condition type")
}

// UnmarshalJSON custom unmarshals a Condition.
func (t *BooleanExpression) UnmarshalJSON(data []byte) error {
	var temp map[string]interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	condType, ok := temp["type"].(string)
	if !ok {
		return errors.New("missing type field in boolean expression")
	}

	switch condType {
	case "simple_condition":
		var simple SimpleBoolExpr
		if err := json.Unmarshal(data, &simple); err != nil {
			return err
		}
		t.Simple = &simple
	case "aggregate_condition":
		var aggregate ListBoolExpr
		if err := json.Unmarshal(data, &aggregate); err != nil {
			return err
		}
		t.Conjunction = &aggregate
	default:
		return errors.New("unknown condition type " + condType)
	}

	return nil
}
