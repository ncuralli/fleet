package strategy

import (
	"context"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/rancher/fleet/internal/cmd/cli/gitcloner/submodule/capability"
)

// FullSHAStrategy fetches a specific commit with full history (no depth limit).
// This is used when the server supports allow-reachable-sha1-in-want but not shallow.
type FullSHAStrategy struct {
	auth transport.AuthMethod
}

func NewFullSHAStrategy(auth transport.AuthMethod) *FullSHAStrategy {
	return &FullSHAStrategy{auth: auth}
}

func (s *FullSHAStrategy) Type() capability.StrategyType {
	return capability.StrategyFullSHA
}

func (s *FullSHAStrategy) Execute(ctx context.Context, r *git.Repository, req *FetchRequest) error {
	refSpec := config.RefSpec(fmt.Sprintf("%s:refs/heads/temp", req.CommitHash.String()))

	err := r.FetchContext(ctx, &git.FetchOptions{
		RefSpecs: []config.RefSpec{refSpec},
		// No Depth - fetch full history up to this commit
		Auth: s.auth,
		Tags: git.NoTags,
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
