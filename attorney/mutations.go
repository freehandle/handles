package attorney

import (
	"log/slog"

	"github.com/freehandle/breeze/crypto"
)

type Mutations struct {
	GrantPower  map[crypto.Hash]struct{}
	RevokePower map[crypto.Hash]struct{}
	NewMembers  map[crypto.Hash]struct{}
	NewCaption  map[crypto.Hash]struct{}
}

func NewMutations() *Mutations {
	return &Mutations{
		GrantPower:  make(map[crypto.Hash]struct{}),
		RevokePower: make(map[crypto.Hash]struct{}),
		NewMembers:  make(map[crypto.Hash]struct{}),
		NewCaption:  make(map[crypto.Hash]struct{}),
	}
}

func (m *Mutations) HasGrantPower(hash crypto.Hash) bool {
	if m.GrantPower == nil {
		slog.Error("mutations.GrantPower is nil")
		return false
	}
	_, ok := m.GrantPower[hash]
	return ok
}

func (m *Mutations) HasRevokePower(hash crypto.Hash) bool {
	if m.RevokePower == nil {
		slog.Error("mutations.RevokePower is nil")
		return false
	}
	_, ok := m.RevokePower[hash]
	return ok
}

func (m *Mutations) HasMember(hash crypto.Hash) bool {
	if m.NewMembers == nil {
		slog.Error("mutations.NewMembers is nil")
		return false
	}
	_, ok := m.NewMembers[hash]
	return ok
}

func (m *Mutations) HasCaption(hash crypto.Hash) bool {
	if m.NewCaption == nil {
		slog.Error("mutations.NewCaption is nil")
		return false
	}
	_, ok := m.NewCaption[hash]
	return ok
}

func (m *Mutations) Merge(others ...*Mutations) *Mutations {
	grouped := &Mutations{
		GrantPower:  make(map[crypto.Hash]struct{}),
		RevokePower: make(map[crypto.Hash]struct{}),
		NewMembers:  make(map[crypto.Hash]struct{}),
		NewCaption:  make(map[crypto.Hash]struct{}),
	}
	for _, mutations := range others {
		for hash := range mutations.GrantPower {
			grouped.GrantPower[hash] = struct{}{}
		}
		for hash := range mutations.RevokePower {
			grouped.RevokePower[hash] = struct{}{}
			delete(grouped.GrantPower, hash)
		}

		for hash := range mutations.NewMembers {
			grouped.NewMembers[hash] = struct{}{}
		}

		for hash := range mutations.NewCaption {
			grouped.NewCaption[hash] = struct{}{}
		}
	}
	return grouped
}
