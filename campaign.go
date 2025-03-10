package main

import (
	"encoding/json"
	"log"
	"time"
)

// Campaign represents the campaign object.
// Note: For dynamic fields like conditions (which can be one of two types),
// we use json.RawMessage to delay interpretation.
type Campaign struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Active      bool                `json:"active"`
	StartDate   time.Time           `json:"start_date"`
	EndDate     *time.Time          `json:"end_date,omitempty"`
	Triggers    []BooleanExpression `json:"triggers,omitempty"`
	Conditions  []BooleanExpression `json:"conditions,omitempty"`
	Rewards     []Reward            `json:"rewards"`
}

func (c *Campaign) EvaluateTriggers(input json.RawMessage) (bool, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(input, &data); err != nil {
		return false, err
	}
	return c.Evaluate(data, c.Triggers)
}

func (c *Campaign) EvaluateConditions(input json.RawMessage) (bool, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(input, &data); err != nil {
		return false, err
	}
	return c.Evaluate(data, c.Conditions)
}

func (c *Campaign) Evaluate(data map[string]interface{}, conditions []BooleanExpression) (bool, error) {

	for _, cond := range conditions {
		if cond.Simple != nil {
			ok, err := EvaluateSimpleCondition(*cond.Simple, data)
			log.Printf("Evaluate: %v -> ok: %t, err: %v", *cond.Simple, ok, err)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		} else if cond.Conjunction != nil {
			ok, err := EvaluateAggregateCondition(*cond.Conjunction, data)
			log.Printf("Evaluate: %v -> ok: %t, err: %v", *cond.Conjunction, ok, err)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}
	}
	return false, nil
}
