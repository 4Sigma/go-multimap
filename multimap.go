package multimap

import (
	"encoding/json"
	"sync"
)

// MultiMap is a thread-safe map that allows multiple values per key.
// K is the key type (must be comparable), V is the value type.
// The equalsFunc is used to determine value equality.
type MultiMap[K comparable, V any] struct {
	mu         sync.RWMutex
	data       map[K][]V
	equalsFunc func(a, b V) bool
}

// New creates a new MultiMap with the provided value equality function.
func New[K comparable, V any](equalsFunc func(a, b V) bool) *MultiMap[K, V] {
	return &MultiMap[K, V]{
		data:       make(map[K][]V),
		equalsFunc: equalsFunc,
	}
}

// NewFromJSON creates a MultiMap from JSON data and an equality function.
// The JSON should represent a map[K][]V.
func NewFromJSON[K comparable, V any](data []byte, equalsFunc func(a, b V) bool) (*MultiMap[K, V], error) {
	var raw map[K][]V
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	return &MultiMap[K, V]{
		data:       raw,
		equalsFunc: equalsFunc,
	}, nil
}

// Add inserts a value for the given key if it does not already exist (by equality).
func (m *MultiMap[K, V]) Add(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, v := range m.data[key] {
		if m.equalsFunc(v, value) {
			return
		}
	}
	m.data[key] = append(m.data[key], value)
}

// Get returns a copy of the values for the given key.
// If the key does not exist, returns an empty slice.
func (m *MultiMap[K, V]) Get(key K) []V {
	m.mu.RLock()
	defer m.mu.RUnlock()
	vals := m.data[key]
	copied := make([]V, len(vals))
	copy(copied, vals)
	return copied
}

// Remove deletes all values for the given key.
func (m *MultiMap[K, V]) Remove(key K) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
}

// RemoveValue removes a specific value for the given key (by equality).
// If no values remain for the key, the key is removed.
func (m *MultiMap[K, V]) RemoveValue(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	values := m.data[key]
	newValues := make([]V, 0, len(values))
	for _, v := range values {
		if !m.equalsFunc(v, value) {
			newValues = append(newValues, v)
		}
	}
	if len(newValues) == 0 {
		delete(m.data, key)
	} else {
		m.data[key] = newValues
	}
}

// Has returns true if the given value exists for the key (by equality).
func (m *MultiMap[K, V]) Has(key K, value V) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, v := range m.data[key] {
		if m.equalsFunc(v, value) {
			return true
		}
	}
	return false
}

// HasKey returns true if the key exists in the map.
func (m *MultiMap[K, V]) HasKey(key K) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.data[key]
	return ok
}

// Keys returns a slice of all keys in the map.
func (m *MultiMap[K, V]) Keys() []K {
	m.mu.RLock()
	defer m.mu.RUnlock()
	keys := make([]K, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys
}

// Len returns the number of keys in the map.
func (m *MultiMap[K, V]) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.data)
}

// Count returns the total number of values across all keys.
func (m *MultiMap[K, V]) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, values := range m.data {
		count += len(values)
	}
	return count
}

// Clear removes all keys and values from the map.
func (m *MultiMap[K, V]) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = make(map[K][]V)
}

// ForEach calls the provided function for each key-value pair.
func (m *MultiMap[K, V]) ForEach(f func(K, V)) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for k, values := range m.data {
		for _, v := range values {
			f(k, v)
		}
	}
}

// Clone returns a deep copy of the MultiMap.
func (m *MultiMap[K, V]) Clone() *MultiMap[K, V] {
	m.mu.RLock()
	defer m.mu.RUnlock()

	clone := New[K](m.equalsFunc)

	for k, values := range m.data {
		copied := make([]V, len(values))
		copy(copied, values)
		clone.data[k] = copied
	}
	return clone
}

// Equal returns true if the two MultiMaps contain the same keys and values (by equality).
func Equal[K comparable, V any](a, b *MultiMap[K, V]) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	b.mu.RLock()
	defer b.mu.RUnlock()

	if len(a.data) != len(b.data) {
		return false
	}

	for k, vals1 := range a.data {
		vals2, ok := b.data[k]
		if !ok || len(vals1) != len(vals2) {
			return false
		}
		for _, v1 := range vals1 {
			found := false
			for _, v2 := range vals2 {
				if a.equalsFunc(v1, v2) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}
	return true
}

// Equal returns true if the other MultiMap contains the same keys and values (by equality).
func (m *MultiMap[K, V]) Equal(other *MultiMap[K, V]) bool {
	return Equal(m, other)
}

// MarshalJSON implements json.Marshaler for MultiMap.
func (m *MultiMap[K, V]) MarshalJSON() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return json.Marshal(m.data)
}

// UnmarshalJSON implements json.Unmarshaler for MultiMap.
func (m *MultiMap[K, V]) UnmarshalJSON(data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return json.Unmarshal(data, &m.data)
}
