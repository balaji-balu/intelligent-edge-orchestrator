package db

import (
	"github.com/balaji-balu/margo-hello-world/ent"
)

type DbStore struct {
	client *ent.Client
}

func NewDbStore(client *ent.Client) *DbStore {
    return &DbStore{client: client}
}