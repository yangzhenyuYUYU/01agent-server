package cache

import (
	"fmt"

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
	if req.Pattern == "" {
		req.Pattern = "*"
	}
	if req.PageSize > 100 {
		req.PageSize = 100
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

	// 获取每个键的详细信息
	items := make([]tools.KeyInfo, 0, len(keys))
	for _, key := range keys {
		info, err := s.redis.GetKeyInfo(key, req.DBIndex)
		if err != nil {
			repository.Warnf("获取键信息失败 %s: %v", key, err)
			// 创建一个基本的 KeyInfo，表示获取失败
			items = append(items, tools.KeyInfo{
				Key:  key,
				Type: "unknown",
				TTL:  -1,
			})
			continue
		}
		items = append(items, *info)
	}

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
	Value      string `json:"value"`
	Expiration int    `json:"expiration"` // 过期时间（秒），0表示永不过期
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

	// 检查键类型，只允许更新 string 类型
	info, err := s.redis.GetKeyInfo(req.Key, req.DBIndex)
	if err != nil {
		repository.Errorf("获取键信息失败: %v", err)
		return nil, fmt.Errorf("获取键信息失败: %w", err)
	}

	if info.Type != "string" {
		return nil, fmt.Errorf("只支持更新 string 类型的键，当前类型: %s", info.Type)
	}

	// 更新值
	err = s.redis.Set(req.Key, req.Value, req.Expiration, req.DBIndex)
	if err != nil {
		repository.Errorf("更新缓存失败: %v", err)
		return nil, fmt.Errorf("更新缓存失败: %w", err)
	}

	return &UpdateCacheResponse{
		Key:        req.Key,
		Value:      req.Value,
		Expiration: req.Expiration,
		Updated:    true,
	}, nil
}
