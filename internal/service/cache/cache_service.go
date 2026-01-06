package cache

import (
	"fmt"
	"sync"

	"01agent_server/internal/repository"
	"01agent_server/internal/tools"
)

// CacheService 缓存服务
type CacheService struct {
	redis *tools.Redis
}

// NewCacheService 创建缓存服务
func NewCacheService() *CacheService {
	return &CacheService{
		redis: tools.GetRedisInstance(),
	}
}

// ListCacheRequest 列出缓存的请求参数
type ListCacheRequest struct {
	DBIndex  int    `form:"db_index"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Pattern  string `form:"pattern"`
	Keyword  string `form:"keyword"` // 关键词，用于模糊查询key
}

// ListCacheResponse 列出缓存的响应
type ListCacheResponse struct {
	DBIndex  int             `json:"db_index"`
	DBSize   int64           `json:"db_size"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
	List     []tools.KeyInfo `json:"list"`
}

// ListCache 列出Redis缓存
func (s *CacheService) ListCache(req *ListCacheRequest) (*ListCacheResponse, error) {
	// 参数验证和默认值设置
	if req.DBIndex < 0 {
		req.DBIndex = 0
	}
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// 如果提供了 keyword，则使用它来构建模糊查询的 pattern
	// 如果同时提供了 pattern 和 keyword，keyword 优先
	if req.Keyword != "" {
		req.Pattern = "*" + req.Keyword + "*"
	} else if req.Pattern == "" {
		req.Pattern = "*"
	}

	// 获取所有匹配的键
	allKeys, err := s.redis.GetAllKeys(req.Pattern, req.DBIndex)
	if err != nil {
		repository.Errorf("获取Redis键列表失败: %v", err)
		return nil, fmt.Errorf("获取缓存列表失败: %w", err)
	}

	total := int64(len(allKeys))
	offset := (req.Page - 1) * req.PageSize
	end := offset + req.PageSize
	if end > len(allKeys) {
		end = len(allKeys)
	}

	// 分页获取键
	var keys []string
	if offset < len(allKeys) {
		keys = allKeys[offset:end]
	}

	// 使用并发获取每个键的详细信息，提升性能
	items := s.getKeysInfoConcurrently(keys, req.DBIndex)

	// 获取数据库大小
	dbSize, _ := s.redis.DBSize(req.DBIndex)

	return &ListCacheResponse{
		DBIndex:  req.DBIndex,
		DBSize:   dbSize,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		List:     items,
	}, nil
}

// getKeysInfoConcurrently 并发获取多个键的详细信息
// 使用worker pool模式控制并发数量，提升性能
func (s *CacheService) getKeysInfoConcurrently(keys []string, dbIndex int) []tools.KeyInfo {
	if len(keys) == 0 {
		return []tools.KeyInfo{}
	}

	// 控制并发数量，避免创建过多goroutines导致Redis连接池耗尽
	// 根据键的数量动态调整worker数量，最多20个并发
	maxWorkers := 20
	if len(keys) < maxWorkers {
		maxWorkers = len(keys)
	}

	// 预分配结果切片，保持顺序
	items := make([]tools.KeyInfo, len(keys))
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 创建任务channel
	keyChan := make(chan int, maxWorkers*2) // 带缓冲，提高吞吐量

	// 启动worker goroutines
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range keyChan {
				key := keys[idx]
				info, err := s.redis.GetKeyInfo(key, dbIndex)

				// 准备结果，减少锁持有时间
				var result tools.KeyInfo
				if err != nil {
					repository.Warnf("获取键信息失败 %s: %v", key, err)
					result = tools.KeyInfo{
						Key:  key,
						Type: "unknown",
						TTL:  -1,
					}
				} else {
					result = *info
				}

				// 使用mutex保护共享的items切片写入
				mu.Lock()
				items[idx] = result
				mu.Unlock()
			}
		}()
	}

	// 发送所有索引到channel
	go func() {
		defer close(keyChan)
		for i := range keys {
			keyChan <- i
		}
	}()

	// 等待所有worker完成
	wg.Wait()

	return items
}

// GetCacheDetailRequest 获取缓存详情的请求参数
type GetCacheDetailRequest struct {
	DBIndex int    `form:"db_index"`
	Key     string `form:"key"`
}

// GetCacheDetailResponse 获取缓存详情的响应
type GetCacheDetailResponse struct {
	Key      string `json:"key"`
	Type     string `json:"type"`
	TTL      int64  `json:"ttl"`
	Size     int64  `json:"size,omitempty"`
	Value    string `json:"value,omitempty"`
	ValueLen int64  `json:"value_len,omitempty"`
}

// GetCacheDetail 获取缓存详情
func (s *CacheService) GetCacheDetail(req *GetCacheDetailRequest) (*GetCacheDetailResponse, error) {
	if req.Key == "" {
		return nil, fmt.Errorf("键名不能为空")
	}

	if req.DBIndex < 0 {
		req.DBIndex = 0
	}

	// 获取键信息
	info, err := s.redis.GetKeyInfo(req.Key, req.DBIndex)
	if err != nil {
		repository.Errorf("获取缓存详情失败: %v", err)
		return nil, fmt.Errorf("获取缓存详情失败: %w", err)
	}

	// 如果键不存在
	if info.Type == "none" {
		return nil, fmt.Errorf("键不存在")
	}

	response := &GetCacheDetailResponse{
		Key:  info.Key,
		Type: info.Type,
		TTL:  info.TTL,
	}

	// 根据类型填充不同的字段
	if info.Type == "string" {
		response.Value = info.Value
		response.Size = info.Size
	} else {
		response.ValueLen = info.ValueLen
	}

	return response, nil
}

// ClearCacheRequest 清除缓存的请求参数
type ClearCacheRequest struct {
	DBIndex int    `json:"db_index"`
	Pattern string `json:"pattern"`
}

// ClearCacheResponse 清除缓存的响应
type ClearCacheResponse struct {
	DBIndex      int    `json:"db_index"`
	Pattern      string `json:"pattern"`
	DeletedCount int64  `json:"deleted_count"`
}

// ClearCache 清除Redis缓存
func (s *CacheService) ClearCache(req *ClearCacheRequest) (*ClearCacheResponse, error) {
	if req.DBIndex < 0 {
		return nil, fmt.Errorf("数据库索引不能为负数")
	}

	var deletedCount int64
	var err error

	if req.Pattern == "" {
		// 清空整个数据库
		err = s.redis.ClearAll(req.DBIndex)
		if err == nil {
			dbSize, _ := s.redis.DBSize(req.DBIndex)
			deletedCount = dbSize
		}
	} else {
		// 按模式清除
		keys, err := s.redis.GetAllKeys(req.Pattern, req.DBIndex)
		if err != nil {
			repository.Errorf("获取匹配的键失败: %v", err)
			return nil, fmt.Errorf("清除缓存失败: %w", err)
		}
		if len(keys) > 0 {
			err = s.redis.DeleteKeys(keys, req.DBIndex)
			if err == nil {
				deletedCount = int64(len(keys))
			}
		}
	}

	if err != nil {
		repository.Errorf("清除Redis缓存失败: %v", err)
		return nil, fmt.Errorf("清除缓存失败: %w", err)
	}

	return &ClearCacheResponse{
		DBIndex:      req.DBIndex,
		Pattern:      req.Pattern,
		DeletedCount: deletedCount,
	}, nil
}

// UpdateCacheRequest 更新缓存的请求参数
type UpdateCacheRequest struct {
	DBIndex    int    `json:"db_index"`
	Key        string `json:"key"`
	Value      string `json:"value"`      // 值，如果为空则只更新TTL
	Expiration *int   `json:"expiration"` // 过期时间（秒），nil表示未设置，0表示永不过期
	TTL        *int   `json:"ttl"`        // TTL（过期时间，秒），如果提供则优先使用，nil表示未设置，0表示永不过期
}

// UpdateCacheResponse 更新缓存的响应
type UpdateCacheResponse struct {
	Key        string `json:"key"`
	Value      string `json:"value"`
	Expiration int    `json:"expiration"`
	Updated    bool   `json:"updated"`
}

// UpdateCache 更新Redis缓存
func (s *CacheService) UpdateCache(req *UpdateCacheRequest) (*UpdateCacheResponse, error) {
	if req.Key == "" {
		return nil, fmt.Errorf("键名不能为空")
	}

	if req.DBIndex < 0 {
		req.DBIndex = 0
	}

	// 检查键是否存在
	exists, err := s.redis.Exists(req.Key, req.DBIndex)
	if err != nil {
		repository.Errorf("检查键是否存在失败: %v", err)
		return nil, fmt.Errorf("检查键失败: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("键不存在")
	}

	// 确定使用的过期时间：优先使用TTL字段（如果提供了），否则使用Expiration字段
	// 使用指针类型来区分"未设置"和"设置为0"
	var expiration int
	if req.TTL != nil {
		// TTL字段已提供，优先使用TTL
		expiration = *req.TTL
	} else if req.Expiration != nil {
		// TTL未提供，使用Expiration
		expiration = *req.Expiration
	} else {
		// 两者都未提供，默认为0（永不过期）
		expiration = 0
	}

	// 获取当前值（如果只更新TTL时需要保留原值）
	var currentValue string
	onlyUpdateTTL := req.Value == ""

	if onlyUpdateTTL {
		// 只更新TTL，需要先获取当前值
		info, err := s.redis.GetKeyInfo(req.Key, req.DBIndex)
		if err != nil {
			repository.Errorf("获取键信息失败: %v", err)
			return nil, fmt.Errorf("获取键信息失败: %w", err)
		}

		// 检查键类型
		if info.Type != "string" {
			return nil, fmt.Errorf("只支持更新 string 类型的键，当前类型: %s", info.Type)
		}

		currentValue = info.Value

		// 只更新TTL
		err = s.redis.Expire(req.Key, expiration, req.DBIndex)
		if err != nil {
			repository.Errorf("更新TTL失败: %v", err)
			return nil, fmt.Errorf("更新TTL失败: %w", err)
		}
	} else {
		// 同时更新值和TTL
		// 检查键类型，只允许更新 string 类型
		info, err := s.redis.GetKeyInfo(req.Key, req.DBIndex)
		if err != nil {
			repository.Errorf("获取键信息失败: %v", err)
			return nil, fmt.Errorf("获取键信息失败: %w", err)
		}

		if info.Type != "string" {
			return nil, fmt.Errorf("只支持更新 string 类型的键，当前类型: %s", info.Type)
		}

		// 更新值和TTL
		err = s.redis.Set(req.Key, req.Value, expiration, req.DBIndex)
		if err != nil {
			repository.Errorf("更新缓存失败: %v", err)
			return nil, fmt.Errorf("更新缓存失败: %w", err)
		}
		currentValue = req.Value
	}

	return &UpdateCacheResponse{
		Key:        req.Key,
		Value:      currentValue,
		Expiration: expiration,
		Updated:    true,
	}, nil
}

// DeleteCacheRequest 删除缓存的请求参数
type DeleteCacheRequest struct {
	DBIndex int      `json:"db_index"`
	Keys    []string `json:"keys"`
}

// DeleteCacheResponse 删除缓存的响应
type DeleteCacheResponse struct {
	DBIndex      int      `json:"db_index"`
	Keys         []string `json:"keys"`
	DeletedCount int64    `json:"deleted_count"`
}

// DeleteCache 删除指定的Redis缓存键
func (s *CacheService) DeleteCache(req *DeleteCacheRequest) (*DeleteCacheResponse, error) {
	if req.DBIndex < 0 {
		return nil, fmt.Errorf("数据库索引不能为负数")
	}

	if len(req.Keys) == 0 {
		return nil, fmt.Errorf("键列表不能为空")
	}

	// 批量删除键
	err := s.redis.DeleteKeys(req.Keys, req.DBIndex)
	if err != nil {
		repository.Errorf("删除Redis缓存失败: %v", err)
		return nil, fmt.Errorf("删除缓存失败: %w", err)
	}

	return &DeleteCacheResponse{
		DBIndex:      req.DBIndex,
		Keys:         req.Keys,
		DeletedCount: int64(len(req.Keys)),
	}, nil
}
