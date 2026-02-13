package stacks

const (
	stateUnvisited = 0
	stateVisiting  = 1
	stateVisited   = 2
)

// CycleDetector detects circular dependencies in stack definitions
type CycleDetector struct {
	graph map[string][]string // stack -> dependencies
	state map[string]int       // stack -> visit state
	path  []string             // current DFS path
}

// NewCycleDetector creates a cycle detector from enabled stacks
func NewCycleDetector(stackNames []string) (*CycleDetector, error) {
	detector := &CycleDetector{
		graph: make(map[string][]string),
		state: make(map[string]int),
		path:  make([]string, 0),
	}

	// Build adjacency list
	for _, name := range stackNames {
		stack, err := LoadStack(name)
		if err != nil {
			return nil, err
		}

		detector.graph[name] = stack.Requires
		detector.state[name] = stateUnvisited
	}

	return detector, nil
}

// DetectCycles finds all cycles in the dependency graph
func (d *CycleDetector) DetectCycles() [][]string {
	var cycles [][]string

	// Try starting DFS from each unvisited node
	for node := range d.graph {
		if d.state[node] == stateUnvisited {
			if cycle := d.dfs(node); cycle != nil {
				cycles = append(cycles, cycle)
			}
		}
	}

	return cycles
}

// dfs performs depth-first search to detect cycles
func (d *CycleDetector) dfs(node string) []string {
	d.state[node] = stateVisiting
	d.path = append(d.path, node)

	// Visit all dependencies
	for _, dep := range d.graph[node] {
		switch d.state[dep] {
		case stateVisiting:
			// Back edge found - cycle detected!
			return d.extractCycle(dep)

		case stateUnvisited:
			// Continue DFS
			if cycle := d.dfs(dep); cycle != nil {
				return cycle
			}
		}
		// stateVisited: already fully explored, skip
	}

	// Mark as fully visited
	d.state[node] = stateVisited
	d.path = d.path[:len(d.path)-1] // Pop from path

	return nil
}

// extractCycle extracts the cycle from the current path
func (d *CycleDetector) extractCycle(backNode string) []string {
	// Find where the cycle starts in the path
	cycleStart := -1
	for i, node := range d.path {
		if node == backNode {
			cycleStart = i
			break
		}
	}

	if cycleStart == -1 {
		// Should never happen
		return d.path
	}

	// Extract cycle: from cycleStart to end of path
	cycle := make([]string, len(d.path)-cycleStart)
	copy(cycle, d.path[cycleStart:])

	return cycle
}
