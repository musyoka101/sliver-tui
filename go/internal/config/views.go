package config

// View defines how agents are rendered
type View struct {
	Name string
	Type ViewType
}

type ViewType int

const (
	ViewTypeTree ViewType = iota // Current tree layout with arrows
	ViewTypeBox                  // Compact box with side connectors
	ViewTypeTable                // Professional table view with columns
	ViewTypeDashboard            // Dashboard with analytics panels
	ViewTypeNetworkMap           // Network topology map with subnet grouping
)

// GetView returns a view by index
func GetView(index int) View {
	views := []View{
		{Name: "Box", Type: ViewTypeBox},        // Index 0 - Default
		{Name: "Table", Type: ViewTypeTable},    // Index 1
		{Name: "Dashboard", Type: ViewTypeDashboard}, // Index 2
		{Name: "Network Map", Type: ViewTypeNetworkMap}, // Index 3
	}
	return views[index]
}

// GetViewCount returns the total number of available views
func GetViewCount() int {
	return 4
}
