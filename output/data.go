package output

import "qa-script/rules"

// OutputData contains all data needed for output writers
type OutputData struct {
	TitleOrder  []string
	Grouped     rules.TitleGroupedLocations
	PrioritySet map[string]struct{}
}

// IsPriority checks if a location is in the priority set
func (d *OutputData) IsPriority(location string) bool {
	_, exists := d.PrioritySet[location]
	return exists
}

// NewOutputData creates an OutputData from processor results
func NewOutputData(titleOrder []string, grouped rules.TitleGroupedLocations, priorities []string) *OutputData {
	prioritySet := make(map[string]struct{}, len(priorities))
	for _, p := range priorities {
		prioritySet[p] = struct{}{}
	}

	return &OutputData{
		TitleOrder:  titleOrder,
		Grouped:     grouped,
		PrioritySet: prioritySet,
	}
}
