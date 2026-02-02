package keylock

import (
	"sync"
)

type status[T any] struct {
	computing bool
	done      bool
	waiters   int
	cond      *sync.Cond
	result    T
	err       error
}

type KeyLock[T any] struct {
	locksMu sync.Mutex
	locks   map[string]*status[T]
}

func NewKeyLock[T any]() *KeyLock[T] {
	return &KeyLock[T]{locks: make(map[string]*status[T])}
}

func (k *KeyLock[T]) getLock(key string) *status[T] {
	k.locksMu.Lock()
	defer k.locksMu.Unlock()

	if s, ok := k.locks[key]; ok {
		return s
	}

	s := &status[T]{
		cond: sync.NewCond(&sync.Mutex{}),
	}
	k.locks[key] = s
	return s
}

func (k *KeyLock[T]) release(key string, s *status[T]) {
	s.cond.L.Lock()
	zeroWaiters := s.waiters == 0
	s.cond.L.Unlock()

	if zeroWaiters {
		k.locksMu.Lock()
		delete(k.locks, key)
		k.locksMu.Unlock()
	}
}

func (k *KeyLock[T]) Read(key string, fn func() (T, error)) (res T, err error) {
	s := k.getLock(key)

	s.cond.L.Lock()
	s.waiters++

	for s.computing {
		s.cond.Wait()
	}

	if s.done {
		res, err = s.result, s.err
		s.waiters--
		s.cond.L.Unlock()
		k.release(key, s)
		return
	}

	s.computing = true
	s.cond.L.Unlock()

	func() {
		defer func() {
			if r := recover(); r != nil {
				err = nil
			}
		}()
		res, err = fn()
	}()

	s.cond.L.Lock()
	s.result = res
	s.err = err
	s.done = true
	s.computing = false
	s.cond.Broadcast()
	s.waiters--
	s.cond.L.Unlock()

	k.release(key, s)
	return
}

func (k *KeyLock[T]) Write(key string, fn func()) {
	s := k.getLock(key)

	s.cond.L.Lock()
	s.waiters++

	for s.computing {
		s.cond.Wait()
	}

	if s.done {
		s.waiters--
		s.cond.L.Unlock()
		k.release(key, s)
		return
	}

	s.computing = true
	s.cond.L.Unlock()

	func() {
		defer func() { recover() }()
		fn()
	}()

	s.cond.L.Lock()
	s.done = true
	s.computing = false
	s.cond.Broadcast()
	s.waiters--
	s.cond.L.Unlock()

	k.release(key, s)
}
