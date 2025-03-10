package main

// Reward represents the reward object in the campaign.
type Reward struct {
	Kind     string  `json:"kind,omitempty"`
	Type     string  `json:"type"`
	Value    float64 `json:"value"`
	Currency string  `json:"currency"`
}
