package variable

import "testing"

// fixedSystemProvider exposes one immutable value for chain tests.
type fixedSystemProvider struct {
	// value stores the returned immutable variable.
	value Value
}

// ResolveSystem returns the fixed value for its exact name.
func (provider fixedSystemProvider) ResolveSystem(_ int64, _ Scope, _ int64, name string) (Value, bool) {
	return provider.value, provider.value.Name == name
}

// ListSystem returns the fixed immutable value.
func (provider fixedSystemProvider) ListSystem(_ int64, _ Scope, _ int64) []Value {
	return []Value{provider.value}
}

// TestSystemChainUsesStableProviderOrder verifies custom variables can precede room facts.
func TestSystemChainUsesStableProviderOrder(t *testing.T) {
	first := fixedSystemProvider{value: Value{Name: "@value", IntValue: 1}}
	second := fixedSystemProvider{value: Value{Name: "@value", IntValue: 2}}
	chain := NewSystemChain(nil, first, second)
	value, found := chain.ResolveSystem(1, ScopeFurni, 2, "@value")
	if !found || value.IntValue != 1 {
		t.Fatalf("resolved value=%#v found=%t", value, found)
	}
	values := chain.ListSystem(1, ScopeFurni, 2)
	if len(values) != 2 || values[0].IntValue != 1 || values[1].IntValue != 2 {
		t.Fatalf("listed values=%#v", values)
	}
}
