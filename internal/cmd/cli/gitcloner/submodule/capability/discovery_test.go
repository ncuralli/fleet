package capability

import (
	"github.com/go-git/go-git/v5/plumbing/protocol/packp/capability"
	"testing"
)

func TestCapabilities_CanFetchBySHA(t *testing.T) {
	tests := []struct {
		name     string
		caps     Capabilities
		expected bool
	}{
		{
			name:     "both false",
			caps:     Capabilities{},
			expected: false,
		},
		{
			name:     "only reachable",
			caps:     Capabilities{AllowReachableSHA1InWant: true},
			expected: true,
		},
		{
			name:     "both true",
			caps:     Capabilities{AllowReachableSHA1InWant: true, AllowTipSHA1InWant: true},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.caps.CanFetchBySHA(); got != tt.expected {
				t.Errorf("CanFetchBySHA() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCapabilities_CanShallow(t *testing.T) {
	c := Capabilities{Shallow: true}
	if !c.CanFetchSwallow() {
		t.Error("expected true")
	}

	c = Capabilities{Shallow: false}
	if c.CanFetchSwallow() {
		t.Error("expected false")
	}
}

func TestCapabilitiesFromList(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*capability.List)
		expected Capabilities
	}{
		{
			name:     "empty",
			setup:    func(c *capability.List) {},
			expected: Capabilities{},
		},
		{
			name: "shallow only",
			setup: func(c *capability.List) {
				c.Set(capability.Shallow)
			},
			expected: Capabilities{Shallow: true},
		},
		{
			name: "all capabilities",
			setup: func(c *capability.List) {
				c.Set(capability.AllowReachableSHA1InWant)
				c.Set(capability.AllowTipSHA1InWant)
				c.Set(capability.Shallow)
			},
			expected: Capabilities{
				AllowReachableSHA1InWant: true,
				AllowTipSHA1InWant:       true,
				Shallow:                  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			caps := capability.NewList()
			tt.setup(caps)

			result := capabilitiesFromList(caps)

			if *result != tt.expected {
				t.Errorf("got %+v, want %+v", result, tt.expected)
			}
		})
	}

}

func TestChooseStrategy(t *testing.T) {

	tests := []struct {
		name     string
		caps     *Capabilities
		expected StrategyType
	}{
		{
			name: "shallow SHA - can fetch by SHA and swallow",
			caps: &Capabilities{
				AllowReachableSHA1InWant: true,
				Shallow:                  true,
			},
			expected: StrategyShallowSHA,
		},
		{
			name: "full SHA - can fetch by SHA but no shallow",
			caps: &Capabilities{
				AllowReachableSHA1InWant: true,
				Shallow:                  false,
			},
			expected: StrategyFullSHA,
		},
		{
			name: "incremental deepen - shallow only",
			caps: &Capabilities{
				Shallow: true,
			},
			expected: StrategyIncrementalDeepen,
		},
		{
			name:     "full clone - no capabilities",
			caps:     &Capabilities{},
			expected: StategyFullClone,
		},
	}

	d := &CapabilityDetector{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {})
		got := d.ChooseStrategy(tt.caps)
		if got != tt.expected {
			t.Errorf("got %s, want %s", got, tt.expected)
		}
	}
}

func TestStrategyType_String(t *testing.T) {
	tests := []struct {
		st       StrategyType
		expected string
	}{
		{StrategyShallowSHA, "ShallowSHA"},
		{StrategyFullSHA, "FullSHA"},
		{StrategyIncrementalDeepen, "StrategyIncrementalDeepen"},
		{StategyFullClone, "StategyFullClone"},
		{StrategyType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.st.String(); got != tt.expected {
				t.Errorf("got %s, want %s", got, tt.expected)
			}
		})
	}
}
