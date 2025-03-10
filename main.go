package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"errors"
	"strconv"

	"github.com/xeipuuv/gojsonschema"

	"github.com/oliveagle/jsonpath"
)

// ----------------- Condition Types -----------------

// ----------------- Condition Evaluation -----------------

// EvaluateCondition determines whether the given condition (simple or aggregate)
// is met based on the provided data context.
func EvaluateCondition(conditionRaw json.RawMessage, dataContext map[string]interface{}) (bool, error) {
	var temp map[string]interface{}
	if err := json.Unmarshal(conditionRaw, &temp); err != nil {
		return false, err
	}
	condType, ok := temp["type"].(string)
	if !ok {
		return false, errors.New("missing type field in condition")
	}
	switch condType {
	case "simple_condition":
		var simple SimpleBoolExpr
		if err := json.Unmarshal(conditionRaw, &simple); err != nil {
			return false, err
		}
		return EvaluateSimpleCondition(simple, dataContext)
	case "aggregate_condition":
		var aggregate ListBoolExpr
		if err := json.Unmarshal(conditionRaw, &aggregate); err != nil {
			return false, err
		}
		return EvaluateAggregateCondition(aggregate, dataContext)
	default:
		return false, fmt.Errorf("unknown condition type: %s", condType)
	}
}

// evaluateParameter checks if the parameter is an object with a JSONPath.
// If so, it extracts the value from the source data; otherwise, it returns the literal.
func evaluateParameter(param interface{}, sourceData interface{}) (interface{}, error) {
	if m, ok := param.(map[string]interface{}); ok {
		if jp, exists := m["jsonPath"]; exists {
			if jpStr, ok := jp.(string); ok {
				res, err := jsonpath.JsonPathLookup(sourceData, jpStr)
				if err != nil {
					return nil, err
				}
				return res, nil
			}
		}
	}
	return param, nil
}

// isEqual compares two values by converting them to strings.
func isEqual(a, b interface{}) bool {
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// toFloat attempts to convert a value to a float64.
func toFloat(val interface{}) (int64, bool) {
	switch v := val.(type) {
	case int64:
		return v, true
	case int:
		return int64(v), true
	case int32:
		return int64(v), true
	case string:
		f, err := strconv.ParseInt(v, 0, 64)
		if err == nil {
			return f, true
		}
		return 0, false
	default:
		return 0, false
	}
}

// EvaluateAggregateCondition implements a basic aggregate condition evaluator.
// It assumes that dataContext["eventHistory"] contains a slice of event objects.
// The implementation filters events, groups them using the specified keys,
// and then applies inner conditions (e.g. checking the count of events in a group).
func EvaluateAggregateCondition(aggregate ListBoolExpr, dataContext map[string]interface{}) (bool, error) {
	// Get events from the context.
	eventsRaw, exists := dataContext["eventHistory"]
	if !exists {
		return false, fmt.Errorf("no events found in context for aggregate condition")
	}
	eventsSlice, ok := eventsRaw.([]interface{})
	if !ok {
		return false, fmt.Errorf("events in context is not a slice")
	}

	// Filter events based on each filter condition.
	filteredEvents := []interface{}{}
	for _, event := range eventsSlice {
		passes := true
		for _, filter := range aggregate.Parameters.Filter {
			// For filtering, we assume the event is used for every source.
			tempContext := map[string]interface{}{
				"triggerEvent": dataContext["triggerEvent"],
				"customer":     dataContext["customer"],
				"eventHistory": dataContext["eventHistory"],
				"events":       eventsSlice,
			}

			res, err := filter.Evaluate(tempContext)
			if err != nil {
				return false, err
			}
			if !res {
				passes = false
				break
			}
		}
		if passes {
			filteredEvents = append(filteredEvents, event)
		}
	}

	if len(filteredEvents) == 0 {
		return false, nil
	}

	// Group events by the group_by keys.
	groups := make(map[string][]interface{})
	for _, event := range filteredEvents {
		groupKey := ""
		for _, groupBy := range aggregate.Parameters.GroupBy {
			value, err := jsonpath.JsonPathLookup(event, groupBy.JsonPath)
			if err != nil {
				return false, err
			}
			// Apply a simple transform if specified.
			if groupBy.Transform == "calendar_year" {
				if dateStr, ok := value.(string); ok {
					t, err := time.Parse(time.RFC3339, dateStr)
					if err == nil {
						value = t.Year()
					}
				}
			}
			groupKey += fmt.Sprintf("%v-", value)
		}
		groups[groupKey] = append(groups[groupKey], event)
	}

	// Evaluate inner conditions on each group.
	for _, groupEvents := range groups {
		// For demonstration, we aggregate by counting events.
		aggregateData := map[string]interface{}{
			"count": len(groupEvents),
		}
		tempContext := map[string]interface{}{
			"eventHistory": aggregateData,
		}
		groupPassed := true
		for _, innerCond := range aggregate.Parameters.Conditions {
			res, err := innerCond.Evaluate(tempContext)
			if err != nil {
				return false, err
			}
			if !res {
				groupPassed = false
				break
			}
		}
		if groupPassed {
			return true, nil // At least one group meets the inner conditions.
		}
	}
	return false, nil
}

// InterpretCampaign validates the provided JSON payload against the campaign schema,
// and if valid, unmarshals it into a Campaign struct.
func InterpretCampaign(campaignJSON []byte) (*Campaign, error) {
	// Load the campaign schema from the local file system.
	schemaLoader := gojsonschema.NewReferenceLoader("file://./schemas/campaign.json")
	documentLoader := gojsonschema.NewBytesLoader(campaignJSON)

	// Validate the JSON payload against the schema.
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return nil, fmt.Errorf("error validating campaign: %v", err)
	}

	if !result.Valid() {
		var errorMessages string
		for _, desc := range result.Errors() {
			errorMessages += fmt.Sprintf("- %s\n", desc)
		}
		return nil, fmt.Errorf("campaign JSON is not valid:\n%s", errorMessages)
	}

	// If valid, unmarshal the JSON payload into a Campaign struct.
	var campaign Campaign
	if err := json.Unmarshal(campaignJSON, &campaign); err != nil {
		return nil, fmt.Errorf("error unmarshalling campaign JSON: %v", err)
	}

	return &campaign, nil
}

func ValidateInputJSON(inputJSON []byte) error {
	// Load the input schema from the local file system.
	schemaLoader := gojsonschema.NewReferenceLoader("file://./schemas/input.json")
	documentLoader := gojsonschema.NewBytesLoader(inputJSON)

	// Validate the JSON payload against the schema.
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("error validating input JSON: %v", err)
	}

	if !result.Valid() {
		var errorMessages string
		for _, desc := range result.Errors() {
			errorMessages += fmt.Sprintf("- %s\n", desc)
		}
		return fmt.Errorf("input JSON is not valid:\n%s", errorMessages)
	}

	return nil
}

func main() {
	options := parseOpts()

	// Load a sample campaign JSON (this file should be a valid campaign JSON payload)
	data, err := os.ReadFile(options.campaignFile)
	if err != nil {
		log.Fatalf("Error reading sample campaign JSON: %v", err)
	}

	campaign, err := InterpretCampaign(data)
	if err != nil {
		log.Fatalf("Error interpreting campaign: %v", err)
	}

	fmt.Printf("Campaign '%s' loaded successfully:\n", campaign.Name)

	// Load the input JSON file and check it conforms to the 'input.json' schema
	inputData, err := os.ReadFile(options.dataFile)
	if err != nil {
		log.Fatalf("Error reading input JSON: %v", err)
	}
	if err = ValidateInputJSON(inputData); err != nil {
		log.Fatalf("Error validating input JSON: %s %v", options.dataFile, err)
	}
	log.Printf("Input JSON '%s' is valid", options.dataFile)
	// Further processing can be done here—for example, iterating over triggers or conditions,
	// and interpreting each condition type (e.g. decoding simple vs. aggregate conditions based on the "type" field).

	// Evalute the trigger condiution
	pass, err := campaign.EvaluateTriggers(inputData)
	if err != nil {
		log.Fatalf("Error evaluating triggers: %v", err)
	}
	if pass {
		fmt.Println("Triggers passed")
	} else {
		fmt.Println("Triggers failed")
		return
	}

	// Evalute the conditions condiution
	pass, err = campaign.EvaluateConditions(inputData)
	if err != nil {
		log.Fatalf("Error evaluating conditions: %v", err)
	}
	if pass {
		fmt.Println("Conditions passed, issue reward")
	} else {
		fmt.Println("Conditions failed")
		return
	}

}

type opts struct {
	campaignFile string
	dataFile     string
}

func parseOpts() opts {
	// Parse command line arguments to fill opts
	var options opts
	flag.StringVar(&options.campaignFile, "campaignFile", "./campaigns/sample_1.json", "Path to the campaign JSON file")
	flag.StringVar(&options.dataFile, "data", "./tests/sample_1.json", "Path to the input JSON file")
	flag.Parse()
	return options

}
