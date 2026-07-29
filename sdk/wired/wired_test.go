package wired

import "testing"

// TestResolveDescriptorAppliesSafeDefaultsAndOptions verifies registration metadata.
func TestResolveDescriptorAppliesSafeDefaultsAndOptions(t *testing.T) {
	defaults := ResolveDescriptor(nil)
	if defaults.Selection != SelectionOptional || defaults.Actor != ActorOptional || !defaults.Editor {
		t.Fatalf("defaults=%#v", defaults)
	}
	expected := Descriptor{ClientCode: 7, Selection: SelectionRequired, Actor: ActorPlayer}
	resolved := ResolveDescriptor([]Option{nil, WithDescriptor(expected)})
	if resolved != expected {
		t.Fatalf("resolved=%#v expected=%#v", resolved, expected)
	}
}
