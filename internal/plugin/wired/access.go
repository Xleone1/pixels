package wired

import (
	pluginruntime "github.com/niflaot/pixels/internal/plugin/runtime"
	sdkwired "github.com/niflaot/pixels/sdk/wired"
)

// Access exposes namespaced WIRED registration to one plugin.
type Access struct {
	// registry stores shared plugin WIRED behavior.
	registry *Registry
	// scope identifies the calling plugin.
	scope *pluginruntime.Scope
}

// NewAccess creates one plugin-scoped WIRED registrar.
func NewAccess(registry *Registry, scope *pluginruntime.Scope) *Access {
	return &Access{registry: registry, scope: scope}
}

// RegisterEffect adds one plugin-owned WIRED effect.
func (access *Access) RegisterEffect(key string, executor sdkwired.EffectExecutor, options ...sdkwired.Option) error {
	if access.registry == nil {
		return ErrInvalidRegistration
	}
	return access.registry.register(access.scope, key, sdkwired.ResolveDescriptor(options), executor, nil)
}

// RegisterCondition adds one plugin-owned WIRED condition.
func (access *Access) RegisterCondition(key string, evaluator sdkwired.ConditionEvaluator, options ...sdkwired.Option) error {
	if access.registry == nil {
		return ErrInvalidRegistration
	}
	return access.registry.register(access.scope, key, sdkwired.ResolveDescriptor(options), nil, evaluator)
}
