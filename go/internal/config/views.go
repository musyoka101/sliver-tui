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
		{Name: "Tree", Type: ViewTypeTree},
		{Name: "Box", Type: ViewTypeBox},
		{Name: "Table", Type: ViewTypeTable},
		{Name: "Dashboard", Type: ViewTypeDashboard},
		{Name: "Network Map", Type: ViewTypeNetworkMap},
	}
	return views[index]
}

// GetViewCount returns the total number of available views
func GetViewCount() int {
	return 5
}
