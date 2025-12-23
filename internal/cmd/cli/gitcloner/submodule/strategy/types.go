package strategy

import "github.com/go-git/go-git/v5/plumbing"

type FetchRequest struct {
    CommitHash plumbing.Hash
}
