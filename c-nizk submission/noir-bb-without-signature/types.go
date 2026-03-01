package main

type Constraint struct {
	Operation     int `json:"operation"`
	Value         int `json:"value"`
	AttributeType int `json:"attribute_type"`
}

type ProveRequest struct {
	Data        []int        `json:"data"`
	Constraints []Constraint `json:"constraints"`
}

type ProveResult struct {
	Proof        []byte
	VK           []byte
	PublicInputs []byte
	Err          error
}
