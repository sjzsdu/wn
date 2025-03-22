package data

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type JSONStorage struct {
	cacheDir string
	records  map[string]*CacheRecord
	mu       sync.RWMutex
}

func NewJSONStorage() *JSONStorage {
	return &JSONStorage{
		records: make(map[string]*CacheRecord),
	}
}

func (s *JSONStorage) Init(projectRoot string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	projectName := filepath.Base(projectRoot)
	s.cacheDir = filepath.Join(homeDir, ".wn", projectName)
	if err := os.MkdirAll(s.cacheDir, 0755); err != nil {
		return err
	}

	return s.load()
}

func (s *JSONStorage) Find(path, hash string) (string, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if record, exists := s.records[path]; exists && record.Hash == hash {
		return record.Content, true, nil
	}
	return "", false, nil
}

func (s *JSONStorage) Save(record *CacheRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.records[record.Path] = record
	return s.persist()
}

func (s *JSONStorage) Remove(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.records, path)
	return s.persist()
}

func (s *JSONStorage) GetAll() ([]*CacheRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	records := make([]*CacheRecord, 0, len(s.records))
	for _, record := range s.records {
		records = append(records, record)
	}
	return records, nil
}

func (s *JSONStorage) Close() error {
	return s.persist()
}

func (s *JSONStorage) load() error {
	cacheFile := filepath.Join(s.cacheDir, "cache.json")
	data, err := os.ReadFile(cacheFile)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	var records []*CacheRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return err
	}

	for _, record := range records {
		s.records[record.Path] = record
	}
	return nil
}

func (s *JSONStorage) persist() error {
	records := make([]*CacheRecord, 0, len(s.records))
	for _, record := range s.records {
		records = append(records, record)
	}

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}

	cacheFile := filepath.Join(s.cacheDir, "cache.json")
	return os.WriteFile(cacheFile, data, 0644)
}
