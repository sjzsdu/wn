package data

import (
	"testing"
)

type MockStorage struct {
	records map[string]*CacheRecord
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		records: make(map[string]*CacheRecord),
	}
}

func (m *MockStorage) Init() error                     { return nil }
func (m *MockStorage) Close() error                    { return nil }
func (m *MockStorage) Save(record *CacheRecord) error  { m.records[record.Path] = record; return nil }
func (m *MockStorage) Remove(path string) error        { delete(m.records, path); return nil }
func (m *MockStorage) GetAll() ([]*CacheRecord, error) { return nil, nil }
func (m *MockStorage) Find(path, hash string) (string, bool, error) {
	if record, exists := m.records[path]; exists && record.Hash == hash {
		return record.Content, true, nil
	}
	return "", false, nil
}

func TestCacheManager(t *testing.T) {
	mockStorage := NewMockStorage()
	manager, err := NewCacheManager(mockStorage) // 移除了 projectRoot 参数
	if err != nil {
		t.Fatalf("Failed to create CacheManager: %v", err)
	}

	t.Run("SetRecord and FindContent", func(t *testing.T) {
		// 使用链式调用
		manager.SetRecord("/test/file.go", "hash123", "content123")
		content, found, err := manager.FindContent("/test/file.go", "hash123")
		if err != nil {
			t.Errorf("FindContent failed: %v", err)
		}
		if !found {
			t.Error("Content not found")
		}
		if content != "content123" {
			t.Errorf("Content mismatch, got %s, want %s", content, "content123")
		}
	})

	t.Run("RemoveRecord", func(t *testing.T) {
		err := manager.RemoveRecord("/test/file.go")
		if err != nil {
			t.Errorf("RemoveRecord failed: %v", err)
		}

		_, found, err := manager.FindContent("/test/file.go", "hash123")
		if err != nil {
			t.Errorf("FindContent after remove failed: %v", err)
		}
		if found {
			t.Error("Record should have been removed")
		}
	})
}
