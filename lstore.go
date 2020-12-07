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
	mp   map[int64]uintptr
	cap  int64
	lock int64
}

// LStore singleton local storage fetch.
func LStore(cap ...int64) *LocalStore {
	if lstore == nil {
		vcap := int64(DefaultCapacity)
		if len(cap) == 1 {
			vcap = cap[0]
		}
		lstore = &LocalStore{
			mp:  make(map[int64]uintptr, vcap),
			cap: vcap,
		}
	}
	return lstore
}

// Get returns local goroutine storage value.
// Note: it doesn't wait until atom lock is released,
// we can implement primitive spin lock here but it's out of scope
// for this library now.
func (ls *LocalStore) Get() uintptr {
	if i := atomic.LoadInt64(&ls.lock); i == 0 {
		return ls.mp[gls.GoID()]
	}
	return 0
}

// Set sets local goroutine storage value
// if there any free capacity slot avaliable.
func (ls *LocalStore) Set(v uintptr) {
	atomic.StoreInt64(&ls.lock, 1)
	defer atomic.StoreInt64(&ls.lock, 0)
	if int64(len(ls.mp)) == ls.cap {
		return
	}
	ls.mp[gls.GoID()] = v
}

// Del removes local goroutine storage value
// frees single capacity slot.
func (ls *LocalStore) Del() {
	atomic.StoreInt64(&ls.lock, 1)
	defer atomic.StoreInt64(&ls.lock, 0)
	delete(ls.mp, gls.GoID())
}
