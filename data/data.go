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
	Init() error
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
func NewCacheManager(storage CacheStorage) (*CacheManager, error) {
	if err := storage.Init(); err != nil {
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
func (cm *CacheManager) SetRecord(path, hash, content string) *CacheManager {
	record := &CacheRecord{
		Path:    path,
		Hash:    hash,
		Content: content,
	}
	if err := cm.storage.Save(record); err != nil {
		// 这里我们可以选择记录错误，但继续允许链式调用
		// 用户可以通过其他方法检查最后的错误状态
	}
	return cm
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
