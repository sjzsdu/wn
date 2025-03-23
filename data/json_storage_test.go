package data

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestJSONStorage(t *testing.T) {
	// 创建临时测试目录
	tempDir, err := os.MkdirTemp("", "json_storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("GetAll", func(t *testing.T) {
		// 为每个测试使用唯一的文件名
		cacheFile := filepath.Join(tempDir, fmt.Sprintf("cache_%d.json", time.Now().UnixNano()))
		
		// 创建存储实例
		storage := NewJSONStorage()
		storage.cacheDir = filepath.Dir(cacheFile)
		
		// 添加测试记录
		record := &CacheRecord{
			Path:    "/test/file.go",
			Hash:    "testhash123",
			Content: "test content",
		}
		storage.records = make(map[string]*CacheRecord)
		storage.records[record.Path] = record
		
		err := storage.persist()
		if err != nil {
			t.Fatalf("Failed to persist: %v", err)
		}

		records, err := storage.GetAll()
		if err != nil {
			t.Errorf("Failed to get records: %v", err)
		}

		if len(records) != 1 {
			t.Errorf("Expected 1 record, got %d", len(records))
		}
	})

	t.Run("Find", func(t *testing.T) {
		storage := NewJSONStorage()
		storage.cacheDir = tempDir
		storage.records = make(map[string]*CacheRecord)
		
		// 添加测试记录
		record := &CacheRecord{
			Path:    "/test/file.go",
			Hash:    "testhash123",
			Content: "test content",
		}
		storage.records[record.Path] = record
		
		// 测试查找存在的记录
		content, found, err := storage.Find("/test/file.go", "testhash123")
		if err != nil {
			t.Errorf("Find failed: %v", err)
		}
		if !found {
			t.Error("Record should have been found")
		}
		if content != "test content" {
			t.Errorf("Expected 'test content', got '%s'", content)
		}
		
		// 测试查找不存在的记录
		_, found, err = storage.Find("/nonexistent", "hash")
		if err != nil {
			t.Errorf("Find failed: %v", err)
		}
		if found {
			t.Error("Record should not have been found")
		}
	})

	t.Run("Save", func(t *testing.T) {
		storage := NewJSONStorage()
		storage.cacheDir = tempDir
		storage.records = make(map[string]*CacheRecord)
		
		record := &CacheRecord{
			Path:    "/test/newfile.go",
			Hash:    "newhash",
			Content: "new content",
		}
		
		err := storage.Save(record)
		if err != nil {
			t.Errorf("Save failed: %v", err)
		}
		
		if len(storage.records) != 1 {
			t.Errorf("Expected 1 record, got %d", len(storage.records))
		}
		
		saved := storage.records["/test/newfile.go"]
		if saved.Hash != "newhash" || saved.Content != "new content" {
			t.Error("Record not saved correctly")
		}
	})

	t.Run("Remove", func(t *testing.T) {
		storage := NewJSONStorage()
		storage.cacheDir = tempDir
		storage.records = make(map[string]*CacheRecord)
		
		// 添加测试记录
		record := &CacheRecord{
			Path:    "/test/file.go",
			Hash:    "testhash123",
			Content: "test content",
		}
		storage.records[record.Path] = record
		
		err := storage.Remove("/test/file.go")
		if err != nil {
			t.Errorf("Remove failed: %v", err)
		}

		if len(storage.records) != 0 {
			t.Errorf("Expected 0 records after removal, got %d", len(storage.records))
		}
		
		_, found, err := storage.Find("/test/file.go", "testhash123")
		if err != nil {
			t.Errorf("Find after remove failed: %v", err)
		}
		if found {
			t.Error("Record should have been removed")
		}
	})
	
	t.Run("Load", func(t *testing.T) {
		// 创建一个新的存储实例并保存数据
		storage1 := NewJSONStorage()
		storage1.cacheDir = tempDir
		storage1.records = make(map[string]*CacheRecord)
		
		record := &CacheRecord{
			Path:    "/test/loadtest.go",
			Hash:    "loadhash",
			Content: "load content",
		}
		storage1.records[record.Path] = record
		
		err := storage1.persist()
		if err != nil {
			t.Fatalf("Failed to persist: %v", err)
		}
		
		// 创建另一个实例并加载数据
		storage2 := NewJSONStorage()
		storage2.cacheDir = tempDir
		storage2.records = make(map[string]*CacheRecord)
		
		err = storage2.load()
		if err != nil {
			t.Errorf("Load failed: %v", err)
		}
		
		loaded, found, _ := storage2.Find("/test/loadtest.go", "loadhash")
		if !found {
			t.Error("Record should have been loaded")
		}
		if loaded != "load content" {
			t.Errorf("Expected 'load content', got '%s'", loaded)
		}
	})
}
