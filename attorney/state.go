package attorney

import (
	"github.com/freehandle/breeze/crypto"
	"github.com/freehandle/breeze/middleware/social"
	"github.com/freehandle/breeze/util"
)

type State struct {
	Members   *hashVault
	Captions  *hashVault
	Attorneys *hashVault
}

func OpenState(dataPath string, epoch uint64) *State {
	memebers := NewHashVault("members", epoch, 8, dataPath)
	captions := NewHashVault("captions", epoch, 8, dataPath)
	attorneys := NewHashVault("attorneys", epoch, 8, dataPath)
	if memebers == nil || captions == nil || attorneys == nil {
		return nil
	}
	return &State{
		Members:   memebers,
		Captions:  captions,
		Attorneys: attorneys,
	}
}

func NewGenesisState(dataPath string) *State {
	state := State{
		Members:   NewHashVault("members", 0, 8, dataPath),
		Captions:  NewHashVault("captions", 0, 8, dataPath),
		Attorneys: NewHashVault("poa", 0, 8, dataPath),
	}
	return &state
}

func StateFromBytes(data []byte) *State {
	return nil
}

func (s *State) Validator(mutations ...*Mutations) *MutatingState {
	if len(mutations) == 0 {
		return &MutatingState{
			state:     s,
			mutations: NewMutations(),
		}
	}
	if len(mutations) > 1 {
		mutations[0].Merge(mutations[1:]...)
	}
	return &MutatingState{
		state:     s,
		mutations: mutations[0],
	}
}

func (s *State) Incorporate(mutations *Mutations) {
	if mutations == nil {
		return
	}
	for hash := range mutations.GrantPower {
		s.Attorneys.InsertHash(hash)
	}
	for hash := range mutations.RevokePower {
		s.Attorneys.RemoveHash(hash)
	}
	for hash := range mutations.NewMembers {
		s.Members.InsertHash(hash)
	}
	for hash := range mutations.NewCaption {
		s.Captions.ExistsHash(hash)
	}
}

func (s *State) Checksum() crypto.Hash {
	return crypto.ZeroHash
}

func (s *State) Recover() error {
	return nil
}

func (s *State) PowerOfAttorney(token, attorney crypto.Token) bool {
	if token.Equal(attorney) {
		return true
	}
	join := append(token[:], attorney[:]...)
	hash := crypto.Hasher(join)
	return s.Attorneys.ExistsHash(hash)
}

func (s *State) HasMember(token crypto.Token) bool {
	hash := crypto.HashToken(token)
	return s.Members.ExistsHash(hash)
}

func (s *State) HasHandle(handle string) bool {
	hash := crypto.Hasher([]byte(handle))
	return s.Captions.ExistsHash(hash)
}

func (s *State) Shutdown() {
	s.Members.Close()
	s.Attorneys.Close()
	s.Captions.Close()
}

// Clone creates a copy of the state by cloning the underlying papirus hashtable
// stores.
func (s *State) Clone() chan social.Stateful[*Mutations, *MutatingState] {
	cloned := make(chan social.Stateful[*Mutations, *MutatingState], 2)
	cloned <- &State{
		Members:   s.Members.Clone(),
		Captions:  s.Captions.Clone(),
		Attorneys: s.Attorneys.Clone(),
	}
	return cloned
}

// CloneAsync starts a jobe to cloning the underlying hashtable stores. Returns
// a channel to a state object.
func (s *State) CloneAsync() chan *State {
	output := make(chan *State)
	members := s.Members.hs.CloneAsync()
	captions := s.Captions.hs.CloneAsync()
	attorneys := s.Attorneys.hs.CloneAsync()
	clone := &State{
		Members:   &hashVault{},
		Captions:  &hashVault{},
		Attorneys: &hashVault{},
	}
	go func() {
		count := 0
		for {
			select {
			case clone.Members.hs = <-members:
				count += 1
			case clone.Attorneys.hs = <-attorneys:
				count += 1
			case clone.Captions.hs = <-captions:
			}
			if count == 3 {
				output <- clone
				return
			}
		}
	}()
	return output
}

// ChecksumHash returns the hash of the checksum of the state.
func (s *State) ChecksumPoint() crypto.Hash {
	membersHash := s.Members.hs.Hash(crypto.Hasher)
	captionsHash := s.Captions.hs.Hash(crypto.Hasher)
	attorneysHash := s.Attorneys.hs.Hash(crypto.Hasher)
	return crypto.Hasher(append(membersHash[:], append(captionsHash[:], attorneysHash[:]...)...))
}

func (s *State) Serialize() []byte {
	members := s.Members.Bytes()
	captions := s.Captions.Bytes()
	attorneys := s.Attorneys.Bytes()
	bytes := []byte{}
	util.PutUint64(uint64(len(members)), &bytes)
	bytes = append(bytes, members...)
	util.PutUint64(uint64(len(captions)), &bytes)
	bytes = append(bytes, captions...)
	util.PutUint64(uint64(len(attorneys)), &bytes)
	bytes = append(bytes, attorneys...)
	return bytes
}

func NewStateFromBytes(datapath string) social.StateFromBytes[*Mutations, *MutatingState] {
	return func(data []byte) (social.Stateful[*Mutations, *MutatingState], bool) {
		membersSize, _ := util.ParseUint64(data, 0)
		members := data[8 : 8+membersSize]
		captionsSize, _ := util.ParseUint64(data, int(8+membersSize))
		captions := data[16+membersSize : 16+membersSize+captionsSize]
		attorneys := data[24+membersSize+captionsSize:]
		if datapath == "" {
			return &State{
				Members:   NewMemoryHashVaultFromBytes("members", members),
				Captions:  NewMemoryHashVaultFromBytes("captions", captions),
				Attorneys: NewMemoryHashVaultFromBytes("attorneys", attorneys),
			}, true
		} else {
			return &State{
				Members:   NewFileHashVaultFromBytes(datapath, "members", members),
				Captions:  NewFileHashVaultFromBytes(datapath, "captions", captions),
				Attorneys: NewFileHashVaultFromBytes(datapath, "attorneys", attorneys),
			}, true
		}
	}
}
