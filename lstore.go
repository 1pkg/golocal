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
	mpok map[int64]bool
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
			mp:   make(map[int64]uintptr, vcap),
			mpok: make(map[int64]bool, vcap),
			cap:  vcap,
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
		id := gls.GoID()
		if ls.mpok[id] {
			return ls.mp[gls.GoID()]
		}
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
	id := gls.GoID()
	ls.mp[id] = v
	ls.mpok[id] = true
}

// Del removes local goroutine storage value
// frees single capacity slot.
func (ls *LocalStore) Del() {
	atomic.StoreInt64(&ls.lock, 1)
	defer atomic.StoreInt64(&ls.lock, 0)
	id := gls.GoID()
	delete(ls.mp, id)
	delete(ls.mpok, id)
}
