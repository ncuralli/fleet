package submodule

import (
	"context"
	"github.com/rancher/fleet/internal/cmd/cli/gitcloner/submodule/capability"
	"github.com/rancher/fleet/internal/cmd/cli/gitcloner/submodule/strategy"

	"github.com/rancher/fleet/internal/mocks"
	"go.uber.org/mock/gomock"
	"testing"
	"errors"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/storage/memory"
)

// =============================================================================
// Repository tests
// =============================================================================
func newTestRepository(t *testing.T, remoteUrl string)  *git.Repository {
	t.Helper()
	r, err := git.Init(memory.NewStorage(),nil)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}
	_, err = r.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{remoteUrl},
	})
	if err != nil {
		t.Fatalf("failed to create remote: %v", err)
	}

	return r
}

func newTestRepositoryWithRemote(t *testing.T, remoteName, remoteURL string) *git.Repository {
	t.Helper()

	r, err := git.Init(memory.NewStorage(), nil)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	_, err = r.CreateRemote(&config.RemoteConfig{
		Name: remoteName,
		URLs: []string{remoteURL},
	})
	if err != nil {
		t.Fatalf("failed to create remote: %v", err)
	}

	return r
}


// =============================================================================
// NewFetcher tests
// =============================================================================

func TestNewFetcher_Success(t *testing.T) {
	r := newTestRepository(t, "https://github.com/test/repo.git")
	f, err := NewFetcher(nil,r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f == nil {
		t.Fatal("expected fetcher, got nil")
	}
}

func TestNewFetcher_noRemote(t *testing.T) {
	r, _ := git.Init(memory.NewStorage(), nil)
	_, err := NewFetcher(nil, r)
	if err == nil {
		t.Fatal("expected error")
	}

}

func TestNewFetcher_RemoteNoURLs(t *testing.T) {
	repo, _ := git.Init(memory.NewStorage(), nil)
	repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{},
	})

	_, err := NewFetcher(nil, repo)

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewFetcher_WithCustomRemoteName(t *testing.T) {
	repo := newTestRepositoryWithRemote(t, "upstream", "https://github.com/test/repo.git")

	f, err := NewFetcher(nil, repo, WithRemoteName("upstream"))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f == nil {
		t.Fatal("expected fetcher, got nil")
	}
}

func TestNewFetcher_WithCustomRemoteName_NotFound(t *testing.T) {
	repo := newTestRepository(t, "https://github.com/test/repo.git")

	_, err := NewFetcher(nil, repo, WithRemoteName("upstream"))
	if err == nil {
		t.Fatal("expected error")
	}
}


// =============================================================================
// Fetch tests
// =============================================================================


func TestFetch_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDetector := mocks.NewMockCapabilityDetector(ctrl)
	mockStrategy := mocks.NewMockStrategy(ctrl)

	testURL := "https://github.com/test/repo.git"
	r := newTestRepository(t, testURL)
	caps := &capability.Capabilities{Shallow: true}

	//Setup expectations
	mockDetector.EXPECT().
		Detect(testURL, nil).
		Return(caps, nil)

	mockDetector.EXPECT().
		ChooseStrategy(caps).
		Return(capability.StrategyShallowSHA)

	mockStrategy.EXPECT().
		Execute(gomock.Any(), gomock.Any(),gomock.Any()).
		Return(nil)

	f, err := NewFetcher(nil, r,
		WithDetector(mockDetector),
		WithStrategies(map[capability.StrategyType]Strategy{
			capability.StrategyShallowSHA: mockStrategy,
		}),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = f.Fetch(context.Background(), &strategy.FetchRequest{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestFetch_DetectError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDetector := mocks.NewMockCapabilityDetector(ctrl)
	r := newTestRepository(t, "https://github.com/test/repo.git")


	mockDetector.EXPECT().
		Detect(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("connection refused"))

	f, err := NewFetcher(nil, r,
		WithDetector(mockDetector),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = f.Fetch(context.Background(), &strategy.FetchRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
}


func TestFetch_StrategyNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDetector := mocks.NewMockCapabilityDetector(ctrl)

	r := newTestRepository(t, "https://github.com/test/repo.git")
	caps := &capability.Capabilities{}

	mockDetector.EXPECT().
		Detect(gomock.Any(), gomock.Any()).
		Return(caps, nil)

	mockDetector.EXPECT().
		ChooseStrategy(caps).
		Return(capability.StrategyShallowSHA)

	f, err := NewFetcher(nil, r,
		WithDetector(mockDetector),
		WithStrategies(map[capability.StrategyType]Strategy{}),
	)
	if err != nil {
		t.Fatal(err)
	}
	err = f.Fetch(context.Background(), &strategy.FetchRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFetch_StrategyError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDetector := mocks.NewMockCapabilityDetector(ctrl)
	mockStrategy := mocks.NewMockStrategy(ctrl)

	r := newTestRepository(t, "https://github.com/test/repo.git")
	caps := &capability.Capabilities{}

	mockDetector.EXPECT().
		Detect(gomock.Any(), gomock.Any()).
		Return(caps, nil)

	mockDetector.EXPECT().
		ChooseStrategy(caps).
		Return(capability.StrategyShallowSHA)

	mockStrategy.EXPECT().
		Execute(gomock.Any(), gomock.Any(),gomock.Any()).
		Return(errors.New("fetch failed"))

	f, err := NewFetcher(nil, r,
		WithDetector(mockDetector),
		WithStrategies(map[capability.StrategyType]Strategy{
			capability.StrategyShallowSHA: mockStrategy,
		}),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = f.Fetch(context.Background(), &strategy.FetchRequest{})

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFetch_NilCaps(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDetector := mocks.NewMockCapabilityDetector(ctrl)
	mockStrategy := mocks.NewMockStrategy(ctrl)


	r := newTestRepository(t, "https://github.com/test/repo.git")

	mockDetector.EXPECT().
		Detect(gomock.Any(), gomock.Any()).
		Return(nil, nil)

	mockDetector.EXPECT().
		ChooseStrategy(gomock.Any()).
		Return(capability.StategyFullClone)

	mockStrategy.EXPECT().
		Execute(gomock.Any(), gomock.Any(),gomock.Any()).
		Return(nil)

	f, err := NewFetcher(nil, r,
		WithDetector(mockDetector),
		WithStrategies(map[capability.StrategyType]Strategy{
			capability.StategyFullClone: mockStrategy,
		}),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = f.Fetch(context.Background(), &strategy.FetchRequest{})


	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
