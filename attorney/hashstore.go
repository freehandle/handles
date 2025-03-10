package attorney

import (
	"log/slog"
	"path/filepath"

	"github.com/freehandle/breeze/crypto"
	"github.com/freehandle/papirus"
)

const (
	remove byte = iota
	exists
	insert
)

func deleteOrInsert(found bool, hash crypto.Hash, b *papirus.Bucket, item int64, param []byte) papirus.OperationResult {
	if len(param) < 1 {
		slog.Error("deleteOrInsert called with zero param length")
		return papirus.OperationResult{
			Result: papirus.QueryResult{Ok: false},
		}
	}
	if found {
		if param[0] == remove { //Delete
			return papirus.OperationResult{
				Deleted: &papirus.Item{Bucket: b, Item: item},
				Result:  papirus.QueryResult{Ok: true},
			}
		} else if param[0] == exists { // exists?
			return papirus.OperationResult{
				Result: papirus.QueryResult{Ok: true},
			}
		} else { // insert
			return papirus.OperationResult{
				Result: papirus.QueryResult{Ok: false},
			}
		}
	} else {
		if param[0] == insert {
			b.WriteItem(item, hash[:])
			return papirus.OperationResult{
				Added:  &papirus.Item{Bucket: b, Item: item},
				Result: papirus.QueryResult{Ok: true},
			}
		} else {
			return papirus.OperationResult{
				Result: papirus.QueryResult{Ok: false},
			}
		}
	}
}

type hashVault struct {
	hs *papirus.HashStore[crypto.Hash]
}

func (h *hashVault) Clone() *hashVault {
	return &hashVault{
		hs: h.hs.Clone(),
	}
}

func (w *hashVault) ExistsHash(hash crypto.Hash) bool {
	response := make(chan papirus.QueryResult)
	ok, _ := w.hs.Query(papirus.Query[crypto.Hash]{Hash: hash, Param: []byte{exists}, Response: response})
	return ok
}

func (w *hashVault) ExistsToken(token crypto.Token) bool {
	hash := crypto.HashToken(token)
	return w.ExistsHash(hash)
}

func (w *hashVault) InsertHash(hash crypto.Hash) bool {
	response := make(chan papirus.QueryResult)
	ok, _ := w.hs.Query(papirus.Query[crypto.Hash]{Hash: hash, Param: []byte{insert}, Response: response})
	return ok
}

func (w *hashVault) InsertToken(token crypto.Token) bool {
	hash := crypto.HashToken(token)
	return w.InsertHash(hash)
}

func (w *hashVault) RemoveHash(hash crypto.Hash) bool {
	response := make(chan papirus.QueryResult)
	ok, _ := w.hs.Query(papirus.Query[crypto.Hash]{Hash: hash, Param: []byte{remove}, Response: response})
	return ok
}

func (w *hashVault) RemoveToken(token crypto.Token) bool {
	hash := crypto.HashToken(token)
	return w.RemoveHash(hash)
}

func (w *hashVault) Close() bool {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("hashVault.Close", "msg", err)
		}
	}()
	ok := make(chan bool)
	w.hs.Stop <- ok
	return <-ok
}

func OpenHashVaultFromFile(name string, epoch uint64, bitsForBucket int64, dataPath string) *hashVault {
	if store := papirus.OpenFileStore(filepath.Join(dataPath, name)); store == nil {
		slog.Error("OpenHashVaultFromFile: OpenFileStore returned nil")
		return nil
	} else {
		bucketstore := papirus.NewBucketStore(32, 6, store)
		if bucketstore == nil {
			slog.Error("OpenHashVaultFromFile: NewBucketStore returned nil")
			return nil
		}
		hs := papirus.NewHashStore(name, bucketstore, int(bitsForBucket), deleteOrInsert)
		vault := &hashVault{
			hs: hs,
		}
		vault.hs.Start()
		return vault
	}
}

func NewHashVault(name string, epoch uint64, bitsForBucket int64, dataPath string) *hashVault {
	nbytes := 56 + (32*6+8)*int64(1<<bitsForBucket)
	var bytestore papirus.ByteStore
	if dataPath == "" {
		if store := papirus.NewMemoryStore(nbytes); store == nil {
			slog.Error("NewHashVault: NewMemoryStore returned nil")
			return nil
		} else {
			bytestore = store
		}
	} else {
		if store := papirus.NewFileStore(filepath.Join(dataPath, name), nbytes); store == nil {
			slog.Error("NewHashVault: NewFileStore returned nil")
			return nil
		} else {
			bytestore = store
		}
	}
	bucketstore := papirus.NewBucketStore(32, 6, bytestore)
	if bucketstore == nil {
		slog.Error("NewHashVault: NewBucketStore returned nil")
		return nil
	}
	vault := &hashVault{
		hs: papirus.NewHashStore(name, bucketstore, int(bitsForBucket), deleteOrInsert),
	}
	vault.hs.Start()
	return vault
}

func NewFileHashVaultFromBytes(filePath, name string, data []byte) *hashVault {
	bytestore := papirus.NewFileStore(filePath, 0)
	return newHashVaultFromBytes(name, bytestore, data)
}

func NewMemoryHashVaultFromBytes(name string, data []byte) *hashVault {
	bytestore := papirus.NewMemoryStore(0)
	return newHashVaultFromBytes(name, bytestore, data)
}

func newHashVaultFromBytes(name string, store papirus.ByteStore, data []byte) *hashVault {
	hs := papirus.NewHashStoreFromClonedBytes(name, store, deleteOrInsert, data)
	vault := &hashVault{
		hs: hs,
	}
	vault.hs.Start()
	return vault
}

func (h *hashVault) Bytes() []byte {
	return h.hs.Bytes()
}
