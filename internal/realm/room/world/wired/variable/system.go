package variable

// SystemChain resolves immutable system variables in stable provider order.
type SystemChain struct {
	// providers stores focused immutable sources.
	providers []SystemProvider
}

// NewSystemChain creates an ordered immutable variable provider chain.
func NewSystemChain(providers ...SystemProvider) *SystemChain {
	filtered := make([]SystemProvider, 0, len(providers))
	for _, provider := range providers {
		if provider != nil {
			filtered = append(filtered, provider)
		}
	}
	return &SystemChain{providers: filtered}
}

// ResolveSystem returns the first exact immutable variable match.
func (chain *SystemChain) ResolveSystem(roomID int64, scope Scope, scopeID int64, name string) (Value, bool) {
	if chain == nil {
		return Value{}, false
	}
	for _, provider := range chain.providers {
		if value, found := provider.ResolveSystem(roomID, scope, scopeID, name); found {
			return value, true
		}
	}
	return Value{}, false
}

// ListSystem returns immutable variables from every focused provider.
func (chain *SystemChain) ListSystem(roomID int64, scope Scope, scopeID int64) []Value {
	if chain == nil {
		return nil
	}
	var result []Value
	for _, provider := range chain.providers {
		result = append(result, provider.ListSystem(roomID, scope, scopeID)...)
	}
	return result
}
