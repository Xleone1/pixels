package bridge

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/niflaot/pixels/pkg/bus"
	sdkevent "github.com/niflaot/pixels/sdk/event"
)

// Dispatcher exposes the plugin event hub without coupling this bridge to its implementation.
type Dispatcher interface {
	// Dispatch sends one plugin-facing event.
	Dispatch(context.Context, sdkevent.Event) error
	// HasListeners reports whether an event has enabled observers.
	HasListeners(string) bool
}

// Register installs every generic post-commit realm notification bridge.
func Register(subscriber bus.Subscriber, dispatcher Dispatcher) error {
	for _, value := range eventNames {
		name := value
		_, err := subscriber.Subscribe(bus.Name(name), bus.PriorityNormal, func(ctx context.Context, event bus.Event) error {
			if !dispatcher.HasListeners(name) {
				return nil
			}
			return dispatcher.Dispatch(ctx, sdkevent.NewPublished(name, fields(event.Payload)))
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// Names returns a caller-owned copy of every bridged event identifier.
func Names() []string { return append([]string(nil), eventNames...) }

// fields converts one internal payload into normalized exported fields.
func fields(payload any) map[string]any {
	value := reflect.ValueOf(payload)
	for value.IsValid() && (value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer) {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}
	if !value.IsValid() {
		return nil
	}
	if value.Kind() != reflect.Struct {
		return map[string]any{"Value": normalize(value)}
	}
	result := make(map[string]any, value.NumField())
	kind := value.Type()
	for index := 0; index < value.NumField(); index++ {
		field := kind.Field(index)
		if field.PkgPath != "" {
			continue
		}
		result[field.Name] = normalize(value.Field(index))
	}
	return result
}

// normalize removes internal named types while retaining immutable values.
func normalize(value reflect.Value) any {
	for value.IsValid() && (value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer) {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}
	if !value.IsValid() {
		return nil
	}
	if value.Type() == reflect.TypeOf(time.Time{}) {
		return value.Interface().(time.Time)
	}
	switch value.Kind() {
	case reflect.Struct:
		return fields(value.Interface())
	case reflect.Slice, reflect.Array:
		result := make([]any, value.Len())
		for index := 0; index < value.Len(); index++ {
			result[index] = normalize(value.Index(index))
		}
		return result
	case reflect.Map:
		result := make(map[string]any, value.Len())
		iterator := value.MapRange()
		for iterator.Next() {
			result[fmt.Sprint(normalize(iterator.Key()))] = normalize(iterator.Value())
		}
		return result
	case reflect.String:
		return value.String()
	case reflect.Bool:
		return value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return value.Uint()
	case reflect.Float32, reflect.Float64:
		return value.Float()
	default:
		return fmt.Sprint(value.Interface())
	}
}
