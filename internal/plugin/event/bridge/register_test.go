package bridge

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/niflaot/pixels/pkg/bus"
	sdkevent "github.com/niflaot/pixels/sdk/event"
)

// bridgeDispatcher records immutable events emitted by the bridge.
type bridgeDispatcher struct {
	// events stores dispatched events in order.
	events []sdkevent.Event
}

// Dispatch records one plugin-facing event.
func (dispatcher *bridgeDispatcher) Dispatch(_ context.Context, event sdkevent.Event) error {
	dispatcher.events = append(dispatcher.events, event)
	return nil
}

// HasListeners enables every bridge in focused tests.
func (*bridgeDispatcher) HasListeners(string) bool { return true }

// TestNamesCoverEveryInternalRealmFact verifies the bridge catalog cannot silently drift.
func TestNamesCoverEveryInternalRealmFact(t *testing.T) {
	root := repositoryRoot(t)
	expected := internalEventNames(t, filepath.Join(root, "internal", "realm"))
	delete(expected, sdkevent.PlayerConnectedName)
	delete(expected, sdkevent.CurrencyChangedName)

	actual := make(map[string]struct{}, len(eventNames))
	for _, name := range eventNames {
		if _, duplicate := actual[name]; duplicate {
			t.Fatalf("duplicate bridge event %q", name)
		}
		actual[name] = struct{}{}
	}
	if !reflect.DeepEqual(sortedNames(actual), sortedNames(expected)) {
		t.Fatalf("bridge mismatch\nactual:   %v\nexpected: %v", sortedNames(actual), sortedNames(expected))
	}
}

// TestRegisterNormalizesPayloadWithoutRealmTypes verifies the SDK receives detached values.
func TestRegisterNormalizesPayloadWithoutRealmTypes(t *testing.T) {
	local := bus.New()
	dispatcher := &bridgeDispatcher{}
	if err := Register(local, dispatcher); err != nil {
		t.Fatalf("register: %v", err)
	}
	payload := struct {
		PlayerID int64
		Tags     []string
	}{PlayerID: 7, Tags: []string{"one"}}
	if err := local.Publish(context.Background(), bus.Event{Name: "room.created", Payload: payload}); err != nil {
		t.Fatalf("publish: %v", err)
	}
	if len(dispatcher.events) != 1 {
		t.Fatalf("events=%d", len(dispatcher.events))
	}
	published := dispatcher.events[0].(*sdkevent.Published)
	fields := published.Fields()
	if fields["PlayerID"] != int64(7) || fields["Tags"].([]any)[0] != "one" {
		t.Fatalf("fields=%#v", fields)
	}
}

// TestRegisterForwardsEveryCatalogName verifies every declared bridge is operational.
func TestRegisterForwardsEveryCatalogName(t *testing.T) {
	local := bus.New()
	dispatcher := &bridgeDispatcher{}
	if err := Register(local, dispatcher); err != nil {
		t.Fatalf("register: %v", err)
	}
	for index, name := range eventNames {
		payload := struct {
			Index int
		}{Index: index}
		if err := local.Publish(context.Background(), bus.Event{Name: bus.Name(name), Payload: payload}); err != nil {
			t.Fatalf("publish %s: %v", name, err)
		}
	}
	if len(dispatcher.events) != len(eventNames) {
		t.Fatalf("events=%d names=%d", len(dispatcher.events), len(eventNames))
	}
	for index, bridged := range dispatcher.events {
		published, valid := bridged.(*sdkevent.Published)
		if !valid || published.Name() != eventNames[index] {
			t.Fatalf("event %d=%T %q", index, bridged, bridged.Name())
		}
		value, found := published.Field("Index")
		if !found || value != int64(index) {
			t.Fatalf("event %s fields=%#v", published.Name(), published.Fields())
		}
	}
}

// TestFieldsNormalizesEverySupportedShape verifies reflection never leaks realm named types.
func TestFieldsNormalizesEverySupportedShape(t *testing.T) {
	type privateNumber int32
	type payload struct {
		Public   privateNumber
		Unsigned uint16
		Enabled  bool
		Ratio    float32
		When     time.Time
		Map      map[int]privateNumber
		Pointer  *privateNumber
		hidden   string
	}
	number := privateNumber(7)
	now := time.Unix(100, 0).UTC()
	values := fields(&payload{
		Public: number, Unsigned: 8, Enabled: true, Ratio: 1.5, When: now,
		Map: map[int]privateNumber{9: 10}, Pointer: &number, hidden: "private",
	})
	if values["Public"] != int64(7) || values["Unsigned"] != uint64(8) || values["Enabled"] != true {
		t.Fatalf("scalar fields=%#v", values)
	}
	if values["Ratio"] != float64(1.5) || values["When"] != now || values["Pointer"] != int64(7) {
		t.Fatalf("composite fields=%#v", values)
	}
	if values["Map"].(map[string]any)["9"] != int64(10) {
		t.Fatalf("map=%#v", values["Map"])
	}
	if _, exposed := values["hidden"]; exposed {
		t.Fatal("unexported field crossed the SDK boundary")
	}
	if value := fields(privateNumber(11))["Value"]; value != int64(11) {
		t.Fatalf("root scalar=%#v", value)
	}
	var missing *payload
	if values := fields(missing); values != nil {
		t.Fatalf("nil fields=%#v", values)
	}
	names := Names()
	names[0] = "changed"
	if Names()[0] == "changed" {
		t.Fatal("event names returned shared storage")
	}
}

// repositoryRoot returns the module root derived from this test file.
func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", "..", ".."))
}

// internalEventNames parses every realm event package declaration.
func internalEventNames(t *testing.T, root string) map[string]struct{} {
	t.Helper()
	names := make(map[string]struct{})
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() || filepath.Ext(entry.Name()) != ".go" || strings.HasSuffix(entry.Name(), "_test.go") {
			return walkErr
		}
		file, parseErr := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if parseErr != nil {
			return parseErr
		}
		collectEventNames(file, names)
		return nil
	})
	if err != nil {
		t.Fatalf("scan internal events: %v", err)
	}
	return names
}

// collectEventNames extracts const Name bus.Name string declarations.
func collectEventNames(file *ast.File, names map[string]struct{}) {
	ast.Inspect(file, func(node ast.Node) bool {
		spec, ok := node.(*ast.ValueSpec)
		if !ok || len(spec.Names) != 1 || spec.Names[0].Name != "Name" || len(spec.Values) != 1 {
			return true
		}
		selector, ok := spec.Type.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		identifier, identifierOK := selector.X.(*ast.Ident)
		literal, literalOK := spec.Values[0].(*ast.BasicLit)
		if !identifierOK || identifier.Name != "bus" || selector.Sel.Name != "Name" || !literalOK || literal.Kind != token.STRING {
			return true
		}
		if value, err := strconv.Unquote(literal.Value); err == nil {
			names[value] = struct{}{}
		}
		return true
	})
}

// sortedNames returns deterministic keys for diagnostics.
func sortedNames(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
