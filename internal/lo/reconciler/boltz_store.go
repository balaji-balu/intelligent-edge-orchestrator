package reconciler

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	bolt "go.etcd.io/bbolt"
	"github.com/google/uuid"
)

// ---------------------------
// BUCKET NAMES
// ---------------------------
const (
	bucketDesired    = "desired"
	bucketActual     = "actual"
	bucketNodes      = "nodes"
	bucketHostState  = "host_state"
)

// HostState kept here (you already had it)
type HostState struct {
	LastSeen int64  `json:"last_seen"`
	Misses   int    `json:"misses"`
	Status   string `json:"status"`
}

// Ensure Host and DeploymentComponentStatus are defined in this package.
// If they are defined elsewhere, import or move them.
// type Host struct {
// 	ID     string            `json:"id"`
// 	Labels map[string]string `json:"labels"`
// 	Status string            `json:"status"`
// }

// type DeploymentComponentStatus struct {
// 	ID            uuid.UUID
// 	DeploymentID  string
// 	HostID        string
// 	ComponentName string
// 	ActualHash    string
// 	LastUpdate    time.Time
// }

// ---------------------------
// BoltStore
// ---------------------------
type BoltStore struct {
	db    *bolt.DB
	cache *InMemoryStore // your in-memory cache; ensure type exists in same package
}

// NewBoltStore opens a bolt store wrapper and ensures top-level buckets exist.
// If cache is nil it will create a new in-memory cache instance.
func NewBoltStore(db *bolt.DB, cache *InMemoryStore) (*BoltStore, error) {
	if db == nil {
		return nil, errors.New("nil bolt DB")
	}
	if cache == nil {
		cache = NewInMemoryStore() // ensure this constructor exists
	}

	// ensure top-level buckets exist to avoid nil bucket panics later
	if err := db.Update(func(tx *bolt.Tx) error {
		buckets := [][]byte{
			[]byte(bucketDesired),
			[]byte(bucketActual),
			[]byte(bucketNodes),
			[]byte(bucketHostState),
		}
		for _, b := range buckets {
			if _, err := tx.CreateBucketIfNotExists(b); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &BoltStore{db: db, cache: cache}, nil
}

//
// ============================================================
// helpers: encode/decode JSON
// ============================================================
func encode(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func decode(b []byte, out interface{}) error {
	return json.Unmarshal(b, out)
}

//
// ============================================================
// DESIRED HASHES
// ============================================================
//

// SetDesired stores: desired/depID/component = hash
func (b *BoltStore) SetDesired(ctx context.Context, depID string, comp ComponentSpec, specHash string) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bd, err := tx.CreateBucketIfNotExists([]byte(bucketDesired))
		if err != nil {
			return err
		}
		depB, err := bd.CreateBucketIfNotExists([]byte(depID))
		if err != nil {
			return err
		}
		return depB.Put([]byte(comp.Name), []byte(specHash))
	})
}

func (b *BoltStore) GetDesiredHashes(ctx context.Context, depID string) (map[string]string, error) {
	out := map[string]string{}
	err := b.db.View(func(tx *bolt.Tx) error {
		bd := tx.Bucket([]byte(bucketDesired))
		if bd == nil {
			return nil
		}
		depB := bd.Bucket([]byte(depID))
		if depB == nil {
			return nil
		}
		return depB.ForEach(func(k, v []byte) error {
			out[string(k)] = string(v)
			return nil
		})
	})
	return out, err
}

//
// ============================================================
// ACTUAL HASHES (per host)
// ============================================================
func (b *BoltStore) SetActualHash(ctx context.Context, depID, hostID, compName, hash string) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		ba, err := tx.CreateBucketIfNotExists([]byte(bucketActual))
		if err != nil {
			return err
		}
		depB, err := ba.CreateBucketIfNotExists([]byte(depID))
		if err != nil {
			return err
		}
		hostB, err := depB.CreateBucketIfNotExists([]byte(hostID))
		if err != nil {
			return err
		}
		return hostB.Put([]byte(compName), []byte(hash))
	})
}

func (b *BoltStore) GetActualHashes(ctx context.Context, depID string) (map[string]map[string]string, error) {
	out := map[string]map[string]string{}
	err := b.db.View(func(tx *bolt.Tx) error {
		ba := tx.Bucket([]byte(bucketActual))
		if ba == nil {
			return nil
		}
		depB := ba.Bucket([]byte(depID))
		if depB == nil {
			return nil
		}
		return depB.ForEach(func(k, v []byte) error {
			hostID := string(k)
			hostB := depB.Bucket(k)
			if hostB == nil {
				return nil
			}
			m := map[string]string{}
			hostB.ForEach(func(ck, cv []byte) error {
				m[string(ck)] = string(cv)
				return nil
			})
			out[hostID] = m
			return nil
		})
	})
	return out, err
}

func (b *BoltStore) GetActual(ctx context.Context, deploymentID string) ([]*DeploymentComponentStatus, error) {
	rows := []*DeploymentComponentStatus{}
	err := b.db.View(func(tx *bolt.Tx) error {
		ba := tx.Bucket([]byte(bucketActual))
		if ba == nil {
			return nil
		}
		depB := ba.Bucket([]byte(deploymentID))
		if depB == nil {
			return nil
		}
		return depB.ForEach(func(hk, _ []byte) error {
			hostB := depB.Bucket(hk)
			if hostB == nil {
				return nil
			}
			hostID := string(hk)
			return hostB.ForEach(func(ck, cv []byte) error {
				rows = append(rows, &DeploymentComponentStatus{
					ID:            uuid.New(),
					DeploymentID:  deploymentID,
					HostID:        hostID,
					ComponentName: string(ck),
					ActualHash:    string(cv),
					LastUpdate:    time.Now(),
				})
				return nil
			})
		})
	})
	return rows, err
}

//
// ============================================================
// HOST / NODE MANAGEMENT
// ============================================================
//

func (s *BoltStore) GetNodes(ctx context.Context, siteID string) ([]*Host, error) {
	// try cache first
	if s.cache != nil {
		if cached, _ := s.cache.GetNodes(ctx, siteID); len(cached) > 0 {
			return cached, nil
		}
	}

	out := []*Host{}
	err := s.db.View(func(tx *bolt.Tx) error {
		nodesB := tx.Bucket([]byte(bucketNodes))
		if nodesB == nil {
			return nil
		}
		siteB := nodesB.Bucket([]byte(siteID))
		if siteB == nil {
			return nil
		}
		return siteB.ForEach(func(k, v []byte) error {
			var h Host
			if err := decode(v, &h); err != nil {
				return err
			}
			out = append(out, &h)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	// populate cache if possible
	if s.cache != nil {
		s.cache.SetNodesForSite(siteID, out)
	}
	return out, nil
}

// Add or update host: write-through to cache and persist into DB safely.
func (s *BoltStore) AddOrUpdateHost(ctx context.Context, siteID string, host *Host) error {
	// write-through to cache if present
	if s.cache != nil {
		s.cache.AddOrUpdateHost(siteID, host)
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		// create top-level nodes bucket if missing
		nodesB, err := tx.CreateBucketIfNotExists([]byte(bucketNodes))
		if err != nil {
			return err
		}
		siteB, err := nodesB.CreateBucketIfNotExists([]byte(siteID))
		if err != nil {
			return err
		}
		buf, err := encode(host)
		if err != nil {
			return err
		}
		return siteB.Put([]byte(host.ID), buf)
	})
}

//
// ---------------------------
// HostState Management
// ---------------------------
func (s *BoltStore) SetHostAlive(id string, t time.Time) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucketHostState))
		if err != nil {
			return err
		}
		st := HostState{
			LastSeen: t.Unix(),
			Misses:   0,
			Status:   "alive",
		}
		data, _ := encode(st)
		return b.Put([]byte(id), data)
	})
}

func (s *BoltStore) SetHostDead(id string, lastSeen time.Time) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucketHostState))
		if err != nil {
			return err
		}
		st := HostState{
			LastSeen: lastSeen.Unix(),
			Misses:   0,
			Status:   "dead",
		}
		data, _ := encode(st)
		return b.Put([]byte(id), data)
	})
}

func (s *BoltStore) IncrementMisses(id string, misses int) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucketHostState))
		if err != nil {
			return err
		}
		raw := b.Get([]byte(id))
		if raw == nil {
			return errors.New("host not found")
		}
		var st HostState
		if err := decode(raw, &st); err != nil {
			return err
		}
		st.Misses = misses
		data, _ := encode(st)
		return b.Put([]byte(id), data)
	})
}
