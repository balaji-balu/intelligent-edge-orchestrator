package reconciler

// --------------------
// In-memory Store implementation
// --------------------
import (
	"context"
	"sync"
	"time"
	"github.com/google/uuid"
)
type InMemoryStore struct {
	mu sync.RWMutex

	desired map[string]map[string]string                       // depID -> compName -> hash
	actual  map[string]map[string]map[string]string           // depID -> hostID -> compName -> hash
	nodes   map[string][]*Host                                 // siteID -> hosts
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		desired: make(map[string]map[string]string),
		actual:  make(map[string]map[string]map[string]string),
		nodes:   make(map[string][]*Host),
	}
}

func (s *InMemoryStore) SetDesired(ctx context.Context, depID string, comp ComponentSpec, specHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.desired[depID]; !ok {
		s.desired[depID] = make(map[string]string)
	}
	s.desired[depID][comp.Name] = specHash
	return nil
}

func (s *InMemoryStore) GetDesiredHashes(ctx context.Context, depID string) (map[string]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.desired[depID]
	if !ok {
		return map[string]string{}, nil
	}
	// return a copy
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out, nil
}

func (s *InMemoryStore) SetActualHash(ctx context.Context, depID string, hostID string, compName string, hash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.actual[depID]; !ok {
		s.actual[depID] = make(map[string]map[string]string)
	}
	if _, ok := s.actual[depID][hostID]; !ok {
		s.actual[depID][hostID] = make(map[string]string)
	}
	s.actual[depID][hostID][compName] = hash
	return nil
}

func (s *InMemoryStore) GetActualHashes(ctx context.Context, depID string) (map[string]map[string]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.actual[depID]
	if !ok {
		return map[string]map[string]string{}, nil
	}
	// deep copy
	out := make(map[string]map[string]string, len(m))
	for host, comps := range m {
		out[host] = make(map[string]string, len(comps))
		for c, h := range comps {
			out[host][c] = h
		}
	}
	return out, nil
}

func (s *InMemoryStore) GetActual(ctx context.Context, deploymentID string) ([]*DeploymentComponentStatus, error) {
	// build rows from actual map (for reporter compatibility)
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows := []*DeploymentComponentStatus{}
	if hosts, ok := s.actual[deploymentID]; ok {
		for hostID, comps := range hosts {
			for comp, hash := range comps {
				rows = append(rows, &DeploymentComponentStatus{
					ID:            uuid.New(),
					DeploymentID:  deploymentID,
					HostID:        hostID,
					ComponentName: comp,
					ActualHash:    hash,
					DesiredHash:   "",
					Status:        "unknown",
					LastUpdate:    time.Now(),
				})
			}
		}
	}
	return rows, nil
}

func (s *InMemoryStore) GetNodes(ctx context.Context, siteID string) ([]*Host, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	h := s.nodes[siteID]
	if h == nil {
		return []*Host{}, nil
	}
	out := make([]*Host, len(h))
	copy(out, h)
	return out, nil
}

// Add helper to populate hosts in-memory for tests
func (s *InMemoryStore) AddOrUpdateHost(siteID string, host *Host) {
	s.mu.Lock()
	defer s.mu.Unlock()
	hlist := s.nodes[siteID]
	for i, h := range hlist {
		if h.ID == host.ID {
			hlist[i] = host
			s.nodes[siteID] = hlist
			return
		}
	}
	s.nodes[siteID] = append(hlist, host)
}

func (s *InMemoryStore) SetNodesForSite(siteID string, hosts []*Host) { 
	s.nodes[siteID] = hosts 
}