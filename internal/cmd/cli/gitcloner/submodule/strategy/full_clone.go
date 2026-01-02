package strategy

import (
	"context"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/rancher/fleet/internal/cmd/cli/gitcloner/submodule/capability"
)

// FullCloneStrategy fetches the entire repository (all branches and tags).
// This is the fallback strategy when no optimizations are available.
type FullCloneStrategy struct {
	auth transport.AuthMethod
}

func NewFullCloneStrategy(auth transport.AuthMethod) *FullCloneStrategy {
	return &FullCloneStrategy{auth: auth}
}

func (s *FullCloneStrategy) Type() capability.StrategyType {
	return capability.StrategyFullClone
}

func (s *FullCloneStrategy) Execute(ctx context.Context, r *git.Repository, req *FetchRequest) error {
	// Fetch all branches
	err := r.FetchContext(ctx, &git.FetchOptions{
		RefSpecs: []config.RefSpec{"refs/heads/*:refs/remotes/origin/*"},
		Auth:     s.auth,
		Tags:     git.AllTags,
	})
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	checkoutOpts := &CheckoutOptions{
		Hash: req.CommitHash,
	}
	if err := Checkout(r, checkoutOpts); err != nil {
		return fmt.Errorf("checkout failed: %w", err)
	}

	return nil
}
