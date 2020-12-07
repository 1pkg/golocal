package golocal

import (
	"sync/atomic"

	"github.com/modern-go/gls"
)

// DefaultCapacity defines default local store slots capacity.
const DefaultCapacity = 1024

var lstore *LocalStore

// LocalStore defines naive atomic lock; limited slots capacity;
// allocation free; goroutine local storage implementation.
type LocalStore struct {
	Store map[int64]uintptr
	cap   int64
	lock  int64
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

// Lock defines manual locking operation.
func (ls *LocalStore) Lock() {
	atomic.StoreInt64(&ls.lock, 1)
}

// Lock defines manual unlocking operation.
func (ls *LocalStore) Unlock() {
	atomic.StoreInt64(&ls.lock, 0)
}

// Lock defines manual lock chek operation.
func (ls *LocalStore) Locked() bool {
	return atomic.LoadInt64(&ls.lock) == 0
}

// Get returns local goroutine storage value.
// Note: it doesn't wait until atom lock is released,
// we can implement primitive spin lock here but it's out of scope
// for this library now.
func (ls *LocalStore) Get() uintptr {
	if i := atomic.LoadInt64(&ls.lock); i == 0 {
		atomic.StoreInt64(&ls.lock, 1)
		defer atomic.StoreInt64(&ls.lock, 0)
		if ptr, ok := ls.Store[gls.GoID()]; ok {
			return ptr
		}
	}
	return 0
}

// Set sets local goroutine storage value
// if there any free capacity slot avaliable.
func (ls *LocalStore) Set(v uintptr) {
	if i := atomic.LoadInt64(&ls.lock); i == 0 {
		atomic.StoreInt64(&ls.lock, 1)
		defer atomic.StoreInt64(&ls.lock, 0)
		if int64(len(ls.Store)) == ls.cap {
			return
		}
		ls.Store[gls.GoID()] = v
	}
}

// Del removes local goroutine storage value
// frees single capacity slot.
func (ls *LocalStore) Del() {
	if i := atomic.LoadInt64(&ls.lock); i == 0 {
		atomic.StoreInt64(&ls.lock, 1)
		defer atomic.StoreInt64(&ls.lock, 0)
		delete(ls.Store, gls.GoID())
	}
}

// ID return goroutine local store id.
func ID() int64 {
	return gls.GoID()
}
