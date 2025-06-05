package vmkv

import (
	"fmt"
	"sync"

	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"

	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
)

const (
	// ModuleName defines the module name
	ModuleName = "vmkv"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)

const HOST_WASMX_ENV_KVDB_i32_VER1 = "wasmx_kvdb_i32_1"
const HOST_WASMX_ENV_KVDB_i64_VER1 = "wasmx_kvdb_i64_1"

const HOST_WASMX_ENV_KVDB_EXPORT = "wasmx_kvdb_"

const HOST_WASMX_ENV_KVDB = "kvdb"

type ContextKey string

const KvDbContextKey ContextKey = "kvdb-context"

type Context struct {
	*vmtypes.Context
}

type KvOpenConnection struct {
	Connection    string
	Db            dbm.DB
	Store         storetypes.KVStore
	TempStoresMap map[string]storetypes.CacheKVStore
	StoreKeys     []string
	Closed        chan struct{}
}

func (v *KvOpenConnection) reset() {
	v.Store = nil
	v.TempStoresMap = make(map[string]storetypes.CacheKVStore, 0)
	v.StoreKeys = []string{}
}

func (v *KvOpenConnection) getTempStore(id string) (storetypes.CacheKVStore, bool) {
	store, ok := v.TempStoresMap[id]
	return store, ok
}

func (v *KvOpenConnection) newTempStore(id string) error {
	// nested caches
	var store storetypes.KVStore
	if len(v.StoreKeys) == 0 {
		store = v.Store
	} else {
		store = v.getCurrentStore()
	}
	tmpstore, ok := store.CacheWrap().(storetypes.CacheKVStore)
	if !ok {
		return fmt.Errorf("CacheWrap interface not CacheKVStore")
	}
	v.TempStoresMap[id] = tmpstore
	return nil
}

func (v *KvOpenConnection) newCurrentTempStore(id string) error {
	err := v.newTempStore(id)
	if err != nil {
		return err
	}
	v.StoreKeys = append(v.StoreKeys, id)
	return nil
}

func (v *KvOpenConnection) getCurrentStore() storetypes.KVStore {
	store, _ := v.getTempStore(v.StoreKeys[len(v.StoreKeys)-1])
	return store
}

type KvDbContext struct {
	mtx           sync.Mutex
	DbConnections map[string]*KvOpenConnection
}

func (p *KvDbContext) GetConnection(id string) (*KvOpenConnection, bool) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	db, found := p.DbConnections[id]
	return db, found
}

func (p *KvDbContext) SetConnection(id string, connection string, db dbm.DB, closed chan struct{}) error {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	_, found := p.DbConnections[id]
	if found {
		return fmt.Errorf("cannot overwrite kv db connection: %s", id)
	}
	p.DbConnections[id] = &KvOpenConnection{Db: db, Connection: connection, Closed: closed, StoreKeys: []string{}, TempStoresMap: map[string]storetypes.CacheKVStore{}}
	return nil
}

func (p *KvDbContext) DeleteConnection(id string) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	delete(p.DbConnections, id)
}

type KvConnectionRequest struct {
	Driver string `json:"driver"`
	Dir    string `json:"dir"`
	Name   string `json:"name"`
	Id     string `json:"id"`
	// Options map[string]string TODO
}

type KvConnectionResponse struct {
	Error string `json:"error"`
}

type KvCloseRequest struct {
	Id string `json:"id"`
}

type KvCloseResponse struct {
	Error string `json:"error"`
}

type KvGetRequest struct {
	Id  string `json:"id"`
	Key []byte `json:"key"`
}

type KvGetResponse struct {
	Error string `json:"error"`
	Value []byte `json:"value"`
}

type KvHasRequest struct {
	Id  string `json:"id"`
	Key []byte `json:"key"`
}

type KvHasResponse struct {
	Error string `json:"error"`
	Found bool   `json:"found"`
}

type KvSetRequest struct {
	Id    string `json:"id"`
	Key   []byte `json:"key"`
	Value []byte `json:"value"`
}

type KvSetResponse struct {
	Error string `json:"error"`
}

type KvDeleteRequest struct {
	Id  string `json:"id"`
	Key []byte `json:"key"`
}

type KvDeleteResponse struct {
	Error string `json:"error"`
}

type KvIteratorRequest struct {
	Id    string `json:"id"`
	Start []byte `json:"start"`
	End   []byte `json:"end"`
}

type KvIteratorResponse struct {
	Error  string   `json:"error"`
	Values [][]byte `json:"values"`
}

type KvNewBatchRequest struct {
	Id   string `json:"id"`
	Size int    `json:"size"`
}

type KvNewBatchResponse struct {
	Error string `json:"error"`
}

type KvStatsRequest struct {
	Id string `json:"id"`
}
