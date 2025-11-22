package db

import (
    "context"
    "time"
    "fmt"
    "github.com/google/uuid"
)
func (s *DbStore) SetHostAlive(hostid string, lastSeen time.Time) error {
    uid, err := uuid.Parse(hostid)
    if err != nil {
        return fmt.Errorf("invalid hostID %q: %w", hostid, err)
    } 
    return s.client.Host.
        UpdateOneID(uid).
        SetStatus("alive").
        SetLastSeen(lastSeen).
        SetMisses(0).
        Exec(context.Background())
}

func (s *DbStore) SetHostDead(id string, lastSeen time.Time) error {
    uid, err := uuid.Parse(id)
    if err != nil {
        return fmt.Errorf("invalid hostID %q: %w", id, err)
    }
    return s.client.Host.
        UpdateOneID(uid).
        SetStatus("dead").
        SetLastSeen(lastSeen).
        Exec(context.Background())
}

func (s *DbStore) IncrementMisses(id string, misses int) error {
    uid, err := uuid.Parse(id)
    if err != nil {
        return fmt.Errorf("invalid hostID %q: %w", id, err)
    }    
    return s.client.Host.
        UpdateOneID(uid).
        SetMisses(misses).
        Exec(context.Background())
}
