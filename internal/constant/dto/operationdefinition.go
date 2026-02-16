package dto

type Data struct {
	Groups []OperationalGroup     `json:"groups"`
	Types  []OperationalGroupType `json:"types"`
}
type OperationsDefinition struct {
	Status string `json:"status"`
	Data   Data   `json:"data"`
}
