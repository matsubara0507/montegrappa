package persistence

import "strings"

type MemoryDB struct {
	storage map[string]map[string][]byte
}

func NewMemoryDB() *MemoryDB {
	return &MemoryDB{
		storage: make(map[string]map[string][]byte),
	}
}

func (m *MemoryDB) Get(tableName string, key string) ([]byte, error) {
	t, ok := m.storage[tableName]
	if ok == false {
		return nil, ErrTableNotFound
	}

	v, ok := t[key]
	if ok == false {
		return nil, ErrKeyNotFound
	}
	return v, nil
}

func (m *MemoryDB) Set(tableName string, key string, value []byte) error {
	t, ok := m.storage[tableName]
	if ok == false {
		m.storage[tableName] = make(map[string][]byte)
		t = m.storage[tableName]
	}

	t[key] = value
	return nil
}

func (m *MemoryDB) List(tableName string) ([]string, error) {
	t, ok := m.storage[tableName]
	if ok == false {
		return nil, ErrTableNotFound
	}

	keys := make([]string, 0)
	for k, _ := range t {
		keys = append(keys, k)
	}

	return keys, nil
}

func (m *MemoryDB) Delete(tableName, key string) error {
	t, ok := m.storage[tableName]
	if ok == false {
		return ErrTableNotFound
	}

	delete(t, key)
	return nil
}

func (m *MemoryDB) ListPrefix(tableName string, prefix string) ([]string, error) {
	t, ok := m.storage[tableName]
	if ok == false {
		return nil, ErrTableNotFound
	}

	keys := make([]string, 0)
	for k, _ := range t {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}

	return keys, nil
}

func (m *MemoryDB) Close() error {
	return nil
}
