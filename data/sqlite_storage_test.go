package data

import (
	"testing"
)

func TestSQLiteStorage(t *testing.T) {
    storage := NewSQLiteStorage()

    t.Run("Init", func(t *testing.T) {
        err := storage.Init()  // 移除了 tmpDir 参数
        if err != nil {
            t.Errorf("Init failed: %v", err)
        }
    })

	t.Run("Save and Find", func(t *testing.T) {
		record := &CacheRecord{
			Path:    "/test/file.go",
			Hash:    "testhash123",
			Content: "test content",
		}

		// 保存记录
		err := storage.Save(record)
		if err != nil {
			t.Errorf("Save failed: %v", err)
		}

		// 查找记录
		content, found, err := storage.Find(record.Path, record.Hash)
		if err != nil {
			t.Errorf("Find failed: %v", err)
		}
		if !found {
			t.Error("Record not found")
		}
		if content != record.Content {
			t.Errorf("Content mismatch, got %s, want %s", content, record.Content)
		}
	})

	t.Run("GetAll", func(t *testing.T) {
		records, err := storage.GetAll()
		if err != nil {
			t.Errorf("GetAll failed: %v", err)
		}
		if len(records) != 1 {
			t.Errorf("Expected 1 record, got %d", len(records))
		}
	})

	t.Run("Remove", func(t *testing.T) {
		err := storage.Remove("/test/file.go")
		if err != nil {
			t.Errorf("Remove failed: %v", err)
		}

		_, found, err := storage.Find("/test/file.go", "testhash123")
		if err != nil {
			t.Errorf("Find after remove failed: %v", err)
		}
		if found {
			t.Error("Record should have been removed")
		}
	})

	t.Run("Close", func(t *testing.T) {
		err := storage.Close()
		if err != nil {
			t.Errorf("Close failed: %v", err)
		}
	})
}