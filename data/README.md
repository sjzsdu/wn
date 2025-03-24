### **1. 缓存管理模块的类图**

```mermaid
classDiagram
    class CacheManager {
        +NewCacheManager(storage CacheStorage) *CacheManager
        +NewCacheManagerWithType(storageType StorageType) *CacheManager
        +GetDefaultCacheManager() *CacheManager
        +SetDefaultCacheManager(manager *CacheManager)
        +Init() error
        +Close() error
        +Find(key string) *CacheRecord
        +Save(record *CacheRecord) error
        +Remove(key string) error
        +GetAll() []*CacheRecord
    }

    class CacheRecord {
        +Key string
        +Content string
        +Hash string
        +NewCacheRecord(key string) *CacheRecord
        +SetContent(content string) *CacheRecord
        +SetHash(hash string) *CacheRecord
        +Close() error
    }

    class CacheStorage {
        <<interface>>
        +Load() error
        +Save() error
        +Delete(key string) error
    }

    class JSONStorage {
        +NewJSONStorage() *JSONStorage
        +Load() error
        +Save() error
        +Delete(key string) error
    }

    class SQLiteStorage {
        +NewSQLiteStorage() *SQLiteStorage
        +Load() error
        +Save() error
        +Delete(key string) error
    }

    class StorageType {
        <<enumeration>>
        JSON
        SQLite
    }

    CacheManager --> CacheRecord : 管理
    CacheManager --> CacheStorage : 依赖
    JSONStorage ..|> CacheStorage : 实现
    SQLiteStorage ..|> CacheStorage : 实现
    CacheManager --> StorageType : 使用
```

---

### **2. 存储模块的类图**
```mermaid
classDiagram
    class CacheStorage {
        <<interface>>
        +Load() error
        +Save() error
        +Delete(key string) error
    }

    class JSONStorage {
        +NewJSONStorage() *JSONStorage
        +Load() error
        +Save() error
        +Delete(key string) error
    }

    class SQLiteStorage {
        +NewSQLiteStorage() *SQLiteStorage
        +Load() error
        +Save() error
        +Delete(key string) error
    }

    JSONStorage ..|> CacheStorage : 实现
    SQLiteStorage ..|> CacheStorage : 实现
```

---

### **3. 默认缓存管理模块的类图**
```mermaid
classDiagram
    class CacheManager {
        +GetDefaultCacheManager() *CacheManager
        +SetDefaultCacheManager(manager *CacheManager)
    }

    class defaultCacheManager {
        -manager *CacheManager
    }

    CacheManager --> defaultCacheManager : 管理
```

---

### **4. 缓存管理模块的流程图**
以下是缓存管理模块的核心操作流程，包括初始化、查找、保存和删除缓存记录。

```mermaid
flowchart TD
    A[开始] --> B[初始化缓存管理器]
    B --> C{选择存储类型}
    C -->|JSON| D[创建JSONStorage实例]
    C -->|SQLite| E[创建SQLiteStorage实例]
    D --> F[加载缓存数据]
    E --> F
    F --> G[缓存管理器就绪]
    G --> H{执行操作}
    H -->|查找缓存| I[调用Find方法]
    H -->|保存缓存| J[调用Save方法]
    H -->|删除缓存| K[调用Remove方法]
    H -->|获取所有缓存| L[调用GetAll方法]
    I --> M[返回缓存记录]
    J --> N[保存缓存记录]
    K --> O[删除缓存记录]
    L --> P[返回所有缓存记录]
    M --> Q[结束]
    N --> Q
    O --> Q
    P --> Q
```

---

### **5. 存储模块的流程图**
以下是存储模块的核心操作流程，包括加载、保存和删除缓存数据。

```mermaid
flowchart TD
    A[开始] --> B{选择存储类型}
    B -->|JSON| C[JSONStorage实例]
    B -->|SQLite| D[SQLiteStorage实例]
    C --> E[加载缓存数据]
    D --> E
    E --> F{执行操作}
    F -->|保存缓存| G[调用Save方法]
    F -->|删除缓存| H[调用Delete方法]
    G --> I[保存数据到存储]
    H --> J[从存储中删除数据]
    I --> K[结束]
    J --> K
```

---