package data

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sjzsdu/wn/helper"
)

type SQLiteStorage struct {
	db *sql.DB
}

func NewSQLiteStorage() *SQLiteStorage {
	return &SQLiteStorage{}
}

func (s *SQLiteStorage) Init() error {
	cacheDir := helper.GetPath("cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}

	dbPath := filepath.Join(cacheDir, "cache.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	s.db = db

	// 创建缓存表
	_, err = s.db.Exec(`
        CREATE TABLE IF NOT EXISTS cache_records (
            path TEXT PRIMARY KEY,
            hash TEXT NOT NULL,
            content TEXT NOT NULL
        )
    `)
	return err
}

func (s *SQLiteStorage) Find(path, hash string) (string, bool, error) {
	var content string
	err := s.db.QueryRow("SELECT content FROM cache_records WHERE path = ? AND hash = ?",
		path, hash).Scan(&content)

	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return content, true, nil
}

func (s *SQLiteStorage) Save(record *CacheRecord) error {
	_, err := s.db.Exec(`
        INSERT OR REPLACE INTO cache_records (path, hash, content)
        VALUES (?, ?, ?)
    `, record.Path, record.Hash, record.Content)
	return err
}

func (s *SQLiteStorage) Remove(path string) error {
	_, err := s.db.Exec("DELETE FROM cache_records WHERE path = ?", path)
	return err
}

func (s *SQLiteStorage) GetAll() ([]*CacheRecord, error) {
	rows, err := s.db.Query("SELECT path, hash, content FROM cache_records")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*CacheRecord
	for rows.Next() {
		record := &CacheRecord{}
		if err := rows.Scan(&record.Path, &record.Hash, &record.Content); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

func (s *SQLiteStorage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
