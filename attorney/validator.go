package attorney

import (
	"fmt"

	"github.com/freehandle/breeze/crypto"
)

type MutatingState struct {
	state     *State
	mutations *Mutations
}

func (m *MutatingState) Mutations() *Mutations {
	return m.mutations
}

func (s *MutatingState) SetNewGrantPower(token, attorney crypto.Token) bool {
	join := append(token[:], attorney[:]...)
	hash := crypto.Hasher(join)
	s.mutations.GrantPower[hash] = struct{}{}
	return true
}

func (s *MutatingState) SetNewRevokePower(token, attorney crypto.Token) bool {
	join := append(token[:], attorney[:]...)
	hash := crypto.Hasher(join)
	s.mutations.GrantPower[hash] = struct{}{}
	return true
}

func (s *MutatingState) SetNewMember(token crypto.Token, handle string) bool {
	if (!s.HasHandle(handle)) && (!s.state.HasMember(token)) {
		captionHash := crypto.Hasher([]byte(handle))
		tokenHash := crypto.HashToken(token)
		s.mutations.NewMembers[tokenHash] = struct{}{}
		s.mutations.NewCaption[captionHash] = struct{}{}
		return true
	}
	return false
}

func (s *MutatingState) PowerOfAttorney(token, attorney crypto.Token) bool {
	if token.Equal(attorney) {
		return true
	}
	join := append(token[:], attorney[:]...)
	hash := crypto.Hasher(join)
	_, ok := s.mutations.GrantPower[hash]
	return ok || s.state.Attorneys.ExistsHash(hash)
}

func (s *MutatingState) HasMember(token crypto.Token) bool {
	hash := crypto.HashToken(token)
	_, ok := s.mutations.NewMembers[hash]
	return ok || s.state.Members.ExistsHash(hash)
}

func (s *MutatingState) HasHandle(handle string) bool {
	hash := crypto.Hasher([]byte(handle))
	_, ok := s.mutations.NewCaption[hash]
	return ok || s.state.Captions.ExistsHash(hash)
}

func (v *MutatingState) Validate(data []byte) bool {
	kind := Kind(data)
	if kind == Invalid {
		fmt.Println("axe node: invalid kind", data)
		return false
	}
	var ok bool
	switch kind {
	case JoinNetworkType:
		join := ParseJoinNetwork(data)
		if join != nil {
			ok = v.SetNewMember(join.Author, join.Handle)
			//fmt.Printf("axe node %v:%+v\n", ok, *join)
		} else {
			//fmt.Printf("axe node: could not parse join\n %v\n", data)
		}
	case UpdateInfoType:
		update := ParseUpdateInfo(data)
		if update != nil {
			ok = v.PowerOfAttorney(update.Author, update.Signer)
			if ok {
				ok = v.HasMember(update.Author)
			}
			//fmt.Printf("axe node %v:%+v\n", ok, *update)
		} else {
			//fmt.Printf("axe node %v: could not parse update\n", ok)
		}
	case GrantPowerOfAttorneyType:
		grant := ParseGrantPowerOfAttorney(data)
		if grant != nil {
			ok = v.HasMember(grant.Author)
			if ok {
				ok = v.SetNewGrantPower(grant.Author, grant.Attorney)
			}
			//fmt.Printf("axe node %v:%+v\n", ok, *grant)
		} else {
			//fmt.Printf("axe node %v: could not parse grant\n", ok)
		}
	case RevokePowerOfAttorneyType:
		revoke := ParseRevokePowerOfAttorney(data)
		if revoke != nil {
			ok = v.HasMember(revoke.Author)
			if ok {
				ok = v.SetNewRevokePower(revoke.Author, revoke.Attorney)
			}
			//fmt.Printf("axe node revoke %v:%+v\n", ok, *revoke)
		} else {
			//fmt.Printf("axe node %v: could not parse revoke\n", ok)
		}
	case VoidType:
		fmt.Println("void", data)
		void := ParseVoid(data)
		if void != nil {
			fmt.Println(void)
			ok = v.HasMember(void.Author)
			if ok {
				ok = v.PowerOfAttorney(void.Author, void.Signer)
			}
			fmt.Printf("axe node void %v:%+v\n", ok, *void)
		} else {
			//fmt.Printf("axe node %v: could not parse void\n", ok)
		}
	}
	return ok
}
