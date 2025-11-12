// Package builtin provides registration of default adapters.
package builtin

import (
	"github.com/mlOS-foundation/axon/internal/registry/core"
)

// RegisterDefaultAdapters registers all builtin adapters with the registry.
// This is called automatically when the CLI initializes.
func RegisterDefaultAdapters(registry *core.AdapterRegistry, localRegistryURL string, mirrors []string, hfToken string, enableHF bool) {
	// 1. Local registry (if configured) - highest priority
	if localRegistryURL != "" {
		localAdapter := NewLocalRegistryAdapter(localRegistryURL, mirrors)
		registry.Register(localAdapter)
	}

	// 2. PyTorch Hub - handles pytorch/ and torch/ namespaces
	pytorchAdapter := NewPyTorchHubAdapter()
	registry.Register(pytorchAdapter)

	// 3. TensorFlow Hub - handles tfhub/ and tf/ namespaces
	tfhubAdapter := NewTensorFlowHubAdapter()
	registry.Register(tfhubAdapter)

	// 4. ModelScope - handles modelscope/ and ms/ namespaces
	modelscopeAdapter := NewModelScopeAdapter()
	registry.Register(modelscopeAdapter)

	// 5. Hugging Face (fallback - can handle any model)
	if enableHF {
		if hfToken != "" {
			hfAdapter := NewHuggingFaceAdapterWithToken(hfToken)
			registry.Register(hfAdapter)
		} else {
			hfAdapter := NewHuggingFaceAdapter()
			registry.Register(hfAdapter)
		}
	}
}
