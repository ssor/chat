package server

import "sync"

type SafeUserList struct {
	lock *sync.RWMutex
	bm   map[string]*User
}

// NewBeeMap return new safemap
func NewSafeUserList() *SafeUserList {
	return &SafeUserList{
		lock: new(sync.RWMutex),
		bm:   make(map[string]*User),
	}
}

func (m *SafeUserList) Length() int {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return len(m.bm)
}

// Get from maps return the k's value
func (m *SafeUserList) Get(k string) *User {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if val, ok := m.bm[k]; ok {
		return val
	}
	return nil
}

// Maps the given key and value. Returns false
// if the key is already in the map and changes nothing.
func (m *SafeUserList) Set(k string, v *User) bool {
	m.lock.Lock()
	defer m.lock.Unlock()
	if val, ok := m.bm[k]; !ok {
		m.bm[k] = v
	} else if val != v {
		m.bm[k] = v
	} else {
		return false
	}
	return true
}

// Returns true if k is exist in the map.
func (m *SafeUserList) Check(k string) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if _, ok := m.bm[k]; !ok {
		return false
	}
	return true
}

// Delete the given key and value.
func (m *SafeUserList) Delete(k string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.bm, k)
}

// Items returns all items in safemap.
func (m *SafeUserList) Items() map[string]*User {
	m.lock.RLock()
	defer m.lock.RUnlock()
	r := make(map[string]*User)
	for k, v := range m.bm {
		r[k] = v
	}
	return r
}
