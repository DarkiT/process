package utils

// StrAnyMap implements map[string]interface{} with RWMutex that has switch.
type StrAnyMap struct {
	mu   RWMutex
	data map[string]interface{}
}

// NewStrAnyMap returns an empty StrAnyMap object.
// The parameter `safe` is used to specify whether using map in concurrent-safety,
// which is false in default.
func NewStrAnyMap(safe ...bool) *StrAnyMap {
	return &StrAnyMap{
		mu:   Create(safe...),
		data: make(map[string]interface{}),
	}
}

// Search searches the map with given `key`.
// Second return parameter `found` is true if key was found, otherwise false.
func (m *StrAnyMap) Search(key string) (value interface{}, found bool) {
	m.mu.RLock()
	if m.data != nil {
		value, found = m.data[key]
	}
	m.mu.RUnlock()
	return
}

// Set sets key-value to the hash map.
func (m *StrAnyMap) Set(key string, val interface{}) {
	m.mu.Lock()
	if m.data == nil {
		m.data = make(map[string]interface{})
	}
	m.data[key] = val
	m.mu.Unlock()
}

// Remove deletes value from map by given `key`, and return this deleted value.
func (m *StrAnyMap) Remove(key string) (value interface{}) {
	m.mu.Lock()
	if m.data != nil {
		var ok bool
		if value, ok = m.data[key]; ok {
			delete(m.data, key)
		}
	}
	m.mu.Unlock()
	return
}

// Clear deletes all data of the map, it will remake a new underlying data map.
func (m *StrAnyMap) Clear() {
	m.mu.Lock()
	m.data = make(map[string]interface{})
	m.mu.Unlock()
}

// Iterator iterates the hash map readonly with custom callback function `f`.
// If `f` returns true, then it continues iterating; or false to stop.
func (m *StrAnyMap) Iterator(f func(k string, v interface{}) bool) {
	for k, v := range m.Map() {
		if !f(k, v) {
			break
		}
	}
}

// Map returns the underlying data map.
// Note that, if it's in concurrent-safe usage, it returns a copy of underlying data,
// or else a pointer to the underlying data.
func (m *StrAnyMap) Map() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if !m.mu.IsSafe() {
		return m.data
	}
	data := make(map[string]interface{}, len(m.data))
	for k, v := range m.data {
		data[k] = v
	}
	return data
}

// AnyAnyMap wraps map type `map[interface{}]interface{}` and provides more map features.
type AnyAnyMap struct {
	mu   RWMutex
	data map[interface{}]interface{}
}

// NewAnyAnyMap creates and returns an empty hash map.
// The parameter `safe` is used to specify whether using map in concurrent-safety,
// which is false in default.
func NewAnyAnyMap(safe ...bool) *AnyAnyMap {
	return &AnyAnyMap{
		mu:   Create(safe...),
		data: make(map[interface{}]interface{}),
	}
}

// Set sets key-value to the hash map.
func (m *AnyAnyMap) Set(key interface{}, value interface{}) {
	m.mu.Lock()
	if m.data == nil {
		m.data = make(map[interface{}]interface{})
	}
	m.data[key] = value
	m.mu.Unlock()
}

// StrStrMap implements map[string]string with RWMutex that has switch.
type StrStrMap struct {
	mu   RWMutex
	data map[string]string
}

// NewStrStrMap returns an empty StrStrMap object.
// The parameter `safe` is used to specify whether using map in concurrent-safety,
// which is false in default.
func NewStrStrMap(safe ...bool) *StrStrMap {
	return &StrStrMap{
		data: make(map[string]string),
		mu:   Create(safe...),
	}
}

// Set sets key-value to the hash map.
func (m *StrStrMap) Set(key string, val string) {
	m.mu.Lock()
	if m.data == nil {
		m.data = make(map[string]string)
	}
	m.data[key] = val
	m.mu.Unlock()
}

// Sets batch sets key-values to the hash map.
func (m *StrStrMap) Sets(data map[string]string) {
	m.mu.Lock()
	if m.data == nil {
		m.data = data
	} else {
		for k, v := range data {
			m.data[k] = v
		}
	}
	m.mu.Unlock()
}

// Size returns the size of the map.
func (m *StrStrMap) Size() int {
	m.mu.RLock()
	length := len(m.data)
	m.mu.RUnlock()
	return length
}

// Map returns the underlying data map.
// Note that, if it's in concurrent-safe usage, it returns a copy of underlying data,
// or else a pointer to the underlying data.
func (m *StrStrMap) Map() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if !m.mu.IsSafe() {
		return m.data
	}
	data := make(map[string]string, len(m.data))
	for k, v := range m.data {
		data[k] = v
	}
	return data
}
