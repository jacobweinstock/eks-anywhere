// Package component is a task that will
package component

// TODO(jacobweinstock): rename everything to something like Installmanagementcomponents
import "context"

type ProviderComponents struct { // moved to different package
	// path to a file
	kubeconfig         string
	IsBootstrapCluster bool
}

type Provider interface {
	InstallProvider(context.Context, ProviderComponents) error
}

type BoostrapComponents struct {
	ProviderComponents ProviderComponents // get this from context instead.
	Provider           Provider
}

// At the end of this task there should be a bootstrap cluster that is capable of creating a permanent management cluster via CAPI.
func (b BoostrapComponents) RunTask(ctx context.Context) (context.Context, error) {
	// 1. CAPI components (generic, clusterctl init) (needs version info)
	// 2. EKS-A components (generic?)
	// 3. EKS-A Provider components (provider specific, takes a concrete struct?, clusterctl init --infrastructure)
	if err := b.Provider.InstallProvider(ctx, b.ProviderComponents); err != nil {
		return ctx, err
	}

	return ctx, nil
}
