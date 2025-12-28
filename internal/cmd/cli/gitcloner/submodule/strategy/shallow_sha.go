package strategy

import (
	"context"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/rancher/fleet/internal/cmd/cli/gitcloner/submodule/capability"
)

type ShallowSHAStrategy struct {
	auth transport.AuthMethod
}

func  NewShallowSHAStrategy(auth transport.AuthMethod) *ShallowSHAStrategy {
	return &ShallowSHAStrategy{auth: auth}
}

func (s *ShallowSHAStrategy) Type() capability.StrategyType {
	return capability.StrategyShallowSHA
}

func (s *ShallowSHAStrategy) Execute(ctx context.Context,r *git.Repository, req *FetchRequest) error {

	refSpec := config.RefSpec(fmt.Sprintf("%s:refs/heads/temp", req.CommitHash.String()))

	err := r.FetchContext(ctx, &git.FetchOptions{
	RefSpecs:   []config.RefSpec{refSpec},
	Depth:      1,
	Auth:       s.auth,
	Tags:       git.NoTags,
	})
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	checkoutOpts := &CheckoutOptions{
		Hash:        req.CommitHash,
	}
	if err := Checkout(r, checkoutOpts); err != nil {
		return fmt.Errorf("checkout failed: %w", err)
	}

	return nil
}
