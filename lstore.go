package golocal

import (
	"sync"

	"github.com/modern-go/gls"
)

// DefaultCapacity defines default local store slots capacity.
const DefaultCapacity = 1024

var lstore *LocalStore

// LocalStore defines rw mutex lock; limited slots capacity;
// allocation free; goroutine local storage implementation.
// It exposes store directly to avoid stack store copies.
type LocalStore struct {
	Store map[int64]uintptr
	cap   int64
	lock  sync.RWMutex
}

// LStore singleton local storage fetch.
func LStore(cap ...int64) *LocalStore {
	if lstore == nil {
		vcap := int64(DefaultCapacity)
		if len(cap) == 1 {
			vcap = cap[0]
		}
		lstore = &LocalStore{
			Store: make(map[int64]uintptr, vcap),
			cap:   vcap,
		}
	}
	return lstore
}

// RLock defines manual read locking operation.
func (ls *LocalStore) RLock() int64 {
	ls.lock.RLock()
	return gls.GoID()
}

// RUnlock defines manual read unlocking operation.
func (ls *LocalStore) RUnlock() {
	ls.lock.RUnlock()
}

// Set sets local goroutine storage value
// if there any free capacity slot avaliable.
func (ls *LocalStore) Set(v uintptr) {
	ls.lock.Lock()
	defer ls.lock.Unlock()
	if int64(len(ls.Store)) == ls.cap {
		return
	}
	ls.Store[gls.GoID()] = v
}

// Del removes local goroutine storage value
// frees single capacity slot.
func (ls *LocalStore) Del() {
	ls.lock.Lock()
	defer ls.lock.Unlock()
	delete(ls.Store, gls.GoID())
}
