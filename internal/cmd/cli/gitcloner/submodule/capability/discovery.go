package capability

import (
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp/capability"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/client"
)

// Fetch related capabilities from the Git Server
type Capabilities struct {
	AllowReachableSHA1InWant bool
	AllowTipSHA1InWant       bool
	Shallow                  bool
}

type StrategyType int


const (
	StrategyShallowSHA StrategyType = iota
	StrategyFullSHA
	StrategyIncrementalDeepen
	StategyFullClone
)


func (c *Capabilities) CanFetchBySHA() bool {
	return c.AllowReachableSHA1InWant
}

func (c *Capabilities) CanFetchSwallow() bool {
	return c.Shallow
}

// CapabilityDetector detect if the server provides capabilities from Capabilities struct
type CapabilityDetector struct{}

func NewCapabilityDetector() *CapabilityDetector {
	return &CapabilityDetector{}
}

// Detect ask the server for capabilities and return the supported capabilities
func (d *CapabilityDetector) Detect(url string, auth transport.AuthMethod) (*Capabilities, error) {
	endpoint, err := transport.NewEndpoint(url)
	if err != nil {
		return nil, fmt.Errorf("endpoint: %w", err)
	}

	cli, err := client.NewClient(endpoint)
	if err != nil {
		return nil, fmt.Errorf("client: %w", err)
	}

	session, err := cli.NewUploadPackSession(endpoint, auth)
	if err != nil {
		return nil, fmt.Errorf("session: %w", err)
	}

	defer session.Close()

	advRefs, err := session.AdvertisedReferences()
	if err != nil {
		return nil, fmt.Errorf("advertised refs: %w", err)
	}

	return   capabilitiesFromList(advRefs.Capabilities), nil

}

func capabilitiesFromList(caps *capability.List) *Capabilities {
	return &Capabilities{
		AllowReachableSHA1InWant: caps.Supports(capability.AllowReachableSHA1InWant),
		AllowTipSHA1InWant:       caps.Supports(capability.AllowTipSHA1InWant),
		Shallow:                  caps.Supports(capability.Shallow),
	}
}

// The `ChooseStrategy` method in the `CapabilityDetector` struct is
// determining the appropriate strategy type based on the capabilities
// provided by the Git server. It takes a `Capabilities` struct as input
// and evaluates the capabilities to decide which strategy should be used
// for fetching data from the server.
func (d *CapabilityDetector) ChooseStrategy(caps *Capabilities) StrategyType {
	if caps.CanFetchBySHA() && caps.CanFetchSwallow() {
		return StrategyShallowSHA
	}

	if caps.CanFetchBySHA() && !caps.CanFetchSwallow() {
		return StrategyFullSHA
	}

	if caps.Shallow {
		return StrategyIncrementalDeepen
	}

	return StategyFullClone
}

func (st StrategyType) String() string {
	switch st {
	case StrategyShallowSHA:
		return "ShallowSHA"
	case StrategyFullSHA:
		return "FullSHA"
	case StrategyIncrementalDeepen:
		return "StrategyIncrementalDeepen"
	case StategyFullClone:
		return "StategyFullClone"
	default:
		return "Unknown"
	}
}
