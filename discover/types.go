// Package discover provides generic resource discovery infrastructure for wetwire domain packages.
package discover

// DiscoveredResource represents a resource found during discovery.
type DiscoveredResource struct {
	// Name is the variable name of the resource.
	Name string
	// Type is the qualified type name (e.g., "schema.NodeType", "aws.S3Bucket").
	Type string
	// File is the path to the file containing the resource.
	File string
	// Line is the line number (1-based) where the resource is defined.
	Line int
	// Dependencies are the names of other resources this resource depends on.
	Dependencies []string
}

// TypeMatcher is a function that determines if a type represents a discoverable resource.
// It takes the package name, type name, and imports map, and returns the resource type
// string and whether a match was found.
type TypeMatcher func(pkgName, typeName string, imports map[string]string) (resourceType string, ok bool)

// DiscoverOptions configures the discovery process.
type DiscoverOptions struct {
	// Packages specifies the packages or directories to scan.
	Packages []string
	// Verbose enables verbose output during discovery.
	Verbose bool
	// TypeMatcher is the function used to identify resource types.
	// If nil, no resources will be discovered (but variables will still be tracked).
	TypeMatcher TypeMatcher
}

// DiscoverResult contains the results of a discovery operation.
type DiscoverResult struct {
	// Resources is the list of discovered resources.
	Resources []DiscoveredResource
	// AllVars is a map of all variable names found (for dependency tracking).
	AllVars map[string]bool
	// Errors contains any non-fatal errors encountered during discovery.
	Errors []error
}

// NewDiscoverResult creates an initialized DiscoverResult.
func NewDiscoverResult() *DiscoverResult {
	return &DiscoverResult{
		Resources: make([]DiscoveredResource, 0),
		AllVars:   make(map[string]bool),
		Errors:    make([]error, 0),
	}
}

// Merge combines another DiscoverResult into this one.
func (r *DiscoverResult) Merge(other *DiscoverResult) {
	r.Resources = append(r.Resources, other.Resources...)
	for k, v := range other.AllVars {
		r.AllVars[k] = v
	}
	r.Errors = append(r.Errors, other.Errors...)
}

// AddResource adds a resource to the result.
func (r *DiscoverResult) AddResource(resource DiscoveredResource) {
	r.Resources = append(r.Resources, resource)
}

// AddVar adds a variable name to the AllVars set.
func (r *DiscoverResult) AddVar(name string) {
	r.AllVars[name] = true
}

// AddError adds an error to the result.
func (r *DiscoverResult) AddError(err error) {
	r.Errors = append(r.Errors, err)
}
