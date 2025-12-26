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
	ViewTypeDashboard            // Dashboard with analytics panels
)

// GetView returns a view by index
func GetView(index int) View {
	views := []View{
		{Name: "Tree", Type: ViewTypeTree},
		{Name: "Box", Type: ViewTypeBox},
		{Name: "Dashboard", Type: ViewTypeDashboard},
	}
	return views[index]
}

// GetViewCount returns the total number of available views
func GetViewCount() int {
	return 3
}
