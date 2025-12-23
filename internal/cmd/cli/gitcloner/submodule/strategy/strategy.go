package strategy

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

)

type CheckoutOptions struct {
	Hash           plumbing.Hash
	//SparsePatterns []string
}


func Checkout(r *git.Repository, opts *CheckoutOptions) error {
	wt, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("worktree: %w", err)
	}

	err = wt.Checkout(&git.CheckoutOptions{
		Hash:  opts.Hash,
		Force: true,
	})
	if err != nil {
		return fmt.Errorf("checkout: %w", err)
	}

	return nil
}
