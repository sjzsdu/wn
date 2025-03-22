package data

// CacheRecord 表示单个缓存记录
type CacheRecord struct {
	Path    string `json:"path"`
	Hash    string `json:"hash"`
	Content string `json:"Content"`
}

// CacheStorage 定义缓存存储接口
type CacheStorage interface {
	// Init 初始化存储
	Init(projectRoot string) error
	// Find 查找记录
	Find(path, hash string) (string, bool, error)
	// Save 保存记录
	Save(record *CacheRecord) error
	// Remove 删除记录
	Remove(path string) error
	// GetAll 获取所有记录
	GetAll() ([]*CacheRecord, error)
	// Close 关闭存储
	Close() error
}

// CacheManager 管理缓存数据
type CacheManager struct {
	storage CacheStorage
}

// NewCacheManager 创建新的缓存管理器
func NewCacheManager(projectRoot string, storage CacheStorage) (*CacheManager, error) {
	if err := storage.Init(projectRoot); err != nil {
		return nil, err
	}

	return &CacheManager{
		storage: storage,
	}, nil
}

// FindContent 根据文件路径和hash值查找AI响应
func (cm *CacheManager) FindContent(path, hash string) (string, bool, error) {
	return cm.storage.Find(path, hash)
}

// SetRecord 设置或更新缓存记录
func (cm *CacheManager) SetRecord(path, hash, content string) error {
	record := &CacheRecord{
		Path:    path,
		Hash:    hash,
		Content: content,
	}
	return cm.storage.Save(record)
}

// RemoveRecord 删除缓存记录
func (cm *CacheManager) RemoveRecord(path string) error {
	return cm.storage.Remove(path)
}

// GetAllRecords 获取所有缓存记录
func (cm *CacheManager) GetAllRecords() ([]*CacheRecord, error) {
	return cm.storage.GetAll()
}

// Close 关闭缓存管理器
func (cm *CacheManager) Close() error {
	return cm.storage.Close()
}
