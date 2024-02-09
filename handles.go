package handles

import (
	"context"

	"github.com/freehandle/breeze/crypto"
	"github.com/freehandle/breeze/middleware/social"
	"github.com/freehandle/breeze/socket"
	"github.com/freehandle/handles/attorney"
)

type HandlesBlock struct {
	Epoch    uint64
	Seal     crypto.Hash
	Grant    map[crypto.Hash]*attorney.GrantPowerOfAttorney
	Revoke   map[crypto.Hash]*attorney.RevokePowerOfAttorney
	Join     map[crypto.Hash]*attorney.JoinNetwork
	Update   map[crypto.Hash]*attorney.UpdateInfo
	Void     map[crypto.Hash]*attorney.Void
	Commited bool
}

func newHandlesBlock(block *social.SocialBlock) *HandlesBlock {
	handlesBlock := &HandlesBlock{
		Epoch:    block.Epoch,
		Seal:     block.SealHash,
		Grant:    make(map[crypto.Hash]*attorney.GrantPowerOfAttorney),
		Revoke:   make(map[crypto.Hash]*attorney.RevokePowerOfAttorney),
		Join:     make(map[crypto.Hash]*attorney.JoinNetwork),
		Update:   make(map[crypto.Hash]*attorney.UpdateInfo),
		Void:     make(map[crypto.Hash]*attorney.Void),
		Commited: block.CommitHash != crypto.ZeroValueHash,
	}
	invalidated := make(map[crypto.Hash]struct{})
	for _, hash := range block.Invalidated {
		invalidated[hash] = struct{}{}

	}
	for n := 0; n < block.Actions.Len(); n++ {
		action := block.Actions.Get(n)
		hash := crypto.Hasher(action)
		if _, ok := invalidated[hash]; ok {
			continue
		}
		switch attorney.Kind(action) {
		case attorney.GrantPowerOfAttorneyType:
			if grant := attorney.ParseGrantPowerOfAttorney(action); grant != nil {
				handlesBlock.Grant[hash] = grant
			}
		case attorney.RevokePowerOfAttorneyType:
			if revoke := attorney.ParseRevokePowerOfAttorney(action); revoke != nil {
				handlesBlock.Revoke[hash] = revoke
			}
		case attorney.JoinNetworkType:
			if join := attorney.ParseJoinNetwork(action); join != nil {
				handlesBlock.Join[hash] = join
			}
		case attorney.UpdateInfoType:
			if update := attorney.ParseUpdateInfo(action); update != nil {
				handlesBlock.Update[hash] = update
			}
		case attorney.VoidType:
			if void := attorney.ParseVoid(action); void != nil {
				handlesBlock.Void[hash] = void
			}
		}
	}
	return handlesBlock
}

func HandlesListener(ctx context.Context, sources *socket.TrustedAggregator) chan *HandlesBlock {
	blocks := make(chan *social.SocialBlock)
	commits := make(chan *social.SocialBlockCommit)
	social.SocialProtocolBlockListener(ctx, 1, sources, blocks, commits)
	newblock := make(chan *HandlesBlock)
	go func() {
		defer close(newblock)
		recent := make([]*HandlesBlock, 0)
		done := ctx.Done()
		for {
			select {
			case <-done:
				return
			case block, ok := <-blocks:
				if !ok {
					return
				}
				handlesBlock := newHandlesBlock(block)
				if handlesBlock.Commited {
					newblock <- handlesBlock
				} else {
					recent = append(recent, handlesBlock)
				}
			case commit, ok := <-commits:
				if !ok {
					return
				}
				for n, block := range recent {
					if block.Seal == commit.SealHash {
						for _, hash := range commit.Invalidated {
							delete(block.Grant, hash)
							delete(block.Revoke, hash)
							delete(block.Join, hash)
							delete(block.Update, hash)
							delete(block.Void, hash)
						}
					}
					newblock <- block
					recent = append(recent[:n], recent[n+1:]...)
				}
			}
		}
	}()
	return newblock
}
