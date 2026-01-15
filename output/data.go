package output

import "qa-script/rules"

// OutputData contains all data needed for output writers
type OutputData struct {
	TitleOrder  []string
	Grouped     rules.TitleGroupedLocations
	PrioritySet map[string]struct{}
	Gap         int // Number of empty columns between groups
	MaxRows     int // Max rows per column before spillover (0 = no limit)
}

// IsPriority checks if a location is in the priority set
func (d *OutputData) IsPriority(location string) bool {
	_, exists := d.PrioritySet[location]
	return exists
}

// NewOutputData creates an OutputData from processor results
func NewOutputData(titleOrder []string, grouped rules.TitleGroupedLocations, priorities []string, gap, maxRows int) *OutputData {
	prioritySet := make(map[string]struct{}, len(priorities))
	for _, p := range priorities {
		prioritySet[p] = struct{}{}
	}

	return &OutputData{
		TitleOrder:  titleOrder,
		Grouped:     grouped,
		PrioritySet: prioritySet,
		Gap:         gap,
		MaxRows:     maxRows,
	}
}

// ColumnsNeeded returns how many columns a group needs based on item count and max_rows
func (d *OutputData) ColumnsNeeded(itemCount int) int {
	if d.MaxRows <= 0 || itemCount == 0 {
		return 1
	}
	return (itemCount + d.MaxRows - 1) / d.MaxRows // ceil(itemCount / max_rows)
}
