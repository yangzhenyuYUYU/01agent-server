package service

import (
	"01agent_server/internal/models"
	"01agent_server/internal/repository"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// ArticleEditService 文章编辑服务
type ArticleEditService struct {
	db *gorm.DB
}

// NewArticleEditService 创建文章编辑服务
func NewArticleEditService() *ArticleEditService {
	return &ArticleEditService{
		db: repository.DB,
	}
}

// PublishEditTaskRequest 发布请求参数
type PublishEditTaskRequest struct {
	ThumbURL           *string `json:"thumb_url"`
	Title              *string `json:"title"`
	Content            *string `json:"content"`
	Author             *string `json:"author"`
	Digest             *string `json:"digest"`
	NeedOpenComment    *int    `json:"need_open_comment"`
	OnlyFansCanComment *int    `json:"only_fans_can_comment"`
	SyncOnline         *bool   `json:"sync_online"`
	SectionHTML        *string `json:"section_html"`
}

// ProcessPublishTask 处理发布任务（后台执行）
func (s *ArticleEditService) ProcessPublishTask(editTaskID string, params *PublishEditTaskRequest, userID string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[ProcessPublishTask] 任务执行panic: %v", r)
		}
	}()

	log.Printf("[ProcessPublishTask] 开始处理发布任务: edit_task_id=%s, user_id=%s", editTaskID, userID)

	// 首先将状态设置为待发布
	if err := s.db.Model(&models.ArticleEditTask{}).
		Where("id = ?", editTaskID).
		Update("status", models.ArticleEditStatusPending).Error; err != nil {
		log.Printf("[ProcessPublishTask] 更新状态失败: %v", err)
		return
	}

	// 查询编辑任务
	var editTask models.ArticleEditTask
	if err := s.db.Where("id = ?", editTaskID).First(&editTask).Error; err != nil {
		log.Printf("[ProcessPublishTask] 找不到编辑任务: %v", err)
		return
	}

	// 查询用户信息
	var user models.User
	if err := s.db.Where("user_id = ?", userID).First(&user).Error; err != nil {
		log.Printf("[ProcessPublishTask] 找不到用户: %v", err)
		s.updateTaskStatus(editTaskID, models.ArticleEditStatusEditing)
		return
	}

	// 检查用户是否绑定公众号
	if user.AppID == nil || *user.AppID == "" {
		log.Printf("[ProcessPublishTask] 用户未配置appid")
		s.updateTaskStatus(editTaskID, models.ArticleEditStatusEditing)
		return
	}

	// 查询用户参数
	var userParams models.UserParameters
	if err := s.db.Where("user_id = ?", userID).First(&userParams).Error; err != nil {
		log.Printf("[ProcessPublishTask] 找不到用户参数: %v", err)
		s.updateTaskStatus(editTaskID, models.ArticleEditStatusEditing)
		return
	}

	if !userParams.IsGzhBind {
		log.Printf("[ProcessPublishTask] 用户未绑定公众号")
		s.updateTaskStatus(editTaskID, models.ArticleEditStatusEditing)
		return
	}

	// TODO: 实现实际的发布逻辑
	// 1. 上传封面图
	// 2. 创建草稿
	// 3. 如果需要，发布草稿

	// 模拟发布成功，更新为草稿状态
	log.Printf("[ProcessPublishTask] 发布任务模拟完成，更新状态为draft")
	s.updateTaskStatus(editTaskID, models.ArticleEditStatusDraft)
}

// updateTaskStatus 更新任务状态
func (s *ArticleEditService) updateTaskStatus(editTaskID string, status string) {
	updates := map[string]interface{}{
		"status": status,
	}

	if status == models.ArticleEditStatusPublished {
		updates["published_at"] = time.Now()
	}

	if err := s.db.Model(&models.ArticleEditTask{}).
		Where("id = ?", editTaskID).
		Updates(updates).Error; err != nil {
		log.Printf("[updateTaskStatus] 更新状态失败: %v", err)
	}
}

// ConvertMarkdownToHTML 将Markdown转换为HTML
func (s *ArticleEditService) ConvertMarkdownToHTML(content string, theme string) (string, error) {
	// TODO: 实现Markdown到HTML的转换
	// 这里需要集成markdown处理器
	return content, nil
}

// UpdateSectionHTML 更新section_html字段
func (s *ArticleEditService) UpdateSectionHTML(editTask *models.ArticleEditTask) error {
	if editTask.Content == "" {
		editTask.SectionHTML = nil
		return nil
	}

	theme := editTask.Theme
	if theme == "" || theme == "none" {
		theme = "default"
	}

	// TODO: 调用Markdown转换
	htmlContent, err := s.ConvertMarkdownToHTML(editTask.Content, theme)
	if err != nil {
		log.Printf("[UpdateSectionHTML] Markdown转换失败: %v", err)
		editTask.SectionHTML = nil
		return err
	}

	editTask.SectionHTML = &htmlContent
	return nil
}

// GetFirstImageFromArticleTask 从文章任务中获取第一张图片
func (s *ArticleEditService) GetFirstImageFromArticleTask(articleTask *models.ArticleTask) (string, error) {
	if articleTask.Images == nil || *articleTask.Images == "" {
		return "", fmt.Errorf("没有图片")
	}

	var imagesMap map[string][]map[string]interface{}
	if err := json.Unmarshal([]byte(*articleTask.Images), &imagesMap); err != nil {
		return "", fmt.Errorf("解析images失败: %v", err)
	}

	// 遍历map获取第一张图片
	for _, images := range imagesMap {
		if len(images) > 0 {
			if imageURL, ok := images[0]["imageUrl"].(string); ok {
				return imageURL, nil
			}
		}
	}

	return "", fmt.Errorf("没有找到有效的图片URL")
}
