package scopedata

// NestedScope represents one value for a single scope type. Instances may have
// children of the next scope type in the hierarchy.
// This is the type that we expect to read and write to "scope data" files.
type NestedScope struct {
	Type     string `hcl:"type,label"`
	Name     string `hcl:"name,label"`
	Address  string
	Children []*NestedScope `hcl:"scope,block"`

	scopeTypeIndex int
}

// Count returns the number of NestedScopes known to the receiver, including
// itself.
func (ns *NestedScope) Count() int {
	if ns.Children == nil || len(ns.Children) == 0 {
		return 1
	}

	childCount := 0
	for _, child := range ns.Children {
		childCount += child.Count()
	}
	return 1 + childCount
}

// CompiledScope returns a CompiledScope object equivalent to the receiver by
// itself, without accounting for its children.
// Optionally provide a parent scope to
func (ns *NestedScope) CompiledScope(parent *CompiledScope) *CompiledScope {
	attrs := make(map[string]interface{})
	scopeTypes := make([]string, 0)
	scopeValues := make([]string, 0)
	if parent != nil {
		attrs = parent.Attributes
		scopeTypes = append(scopeTypes, parent.ScopeTypes...)
		scopeValues = append(scopeValues, parent.ScopeValues...)
	}

	scopeTypes = append(scopeTypes, ns.Type)
	scopeValues = append(scopeValues, ns.Name)
	// TODO: scope value attributes

	return &CompiledScope{
		Attributes:  attrs,
		ScopeTypes:  scopeTypes,
		ScopeValues: scopeValues,
	}
}

// CompiledScopes returns the complete set of CompiledScope objects equivalent
// to every permutation of the reciever and its children.
func (ns *NestedScope) CompiledScopes(parent *CompiledScope) []*CompiledScope {
	scopes := make([]*CompiledScope, 0, ns.Count())

	this := ns.CompiledScope(parent)
	scopes = append(scopes, this)
	for _, child := range ns.Children {
		scopes = append(scopes, child.CompiledScopes(this)...)
	}

	return scopes
}
