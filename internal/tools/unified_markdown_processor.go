package tools

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"01agent_server/internal/models"
	"01agent_server/internal/repository"

	"gorm.io/gorm"
)

// UnifiedMarkdownProcessor 统一的 Markdown 处理器
type UnifiedMarkdownProcessor struct {
	headingCounter map[string]int
	mu             sync.RWMutex
}

// configCache 配置缓存
type configCacheEntry struct {
	config       map[string]interface{}
	updatedAt    *float64
	templateType string
}

var (
	configCache          = make(map[string]*configCacheEntry)
	markdownProcessCache = make(map[string]string)
	cacheMaxSize         = 500
	processCacheMaxSize  = 1000
	cacheMutex           sync.RWMutex
)

// NewUnifiedMarkdownProcessor 创建新的处理器
func NewUnifiedMarkdownProcessor() *UnifiedMarkdownProcessor {
	return &UnifiedMarkdownProcessor{
		headingCounter: make(map[string]int),
	}
}

// LoadUnifiedConfig 加载统一配置（从数据库查找用户模板或官方模板）
func (p *UnifiedMarkdownProcessor) LoadUnifiedConfig(configName string) (map[string]interface{}, error) {
	// 检查缓存
	cacheMutex.RLock()
	if cached, ok := configCache[configName]; ok {
		// 如果不是默认配置，检查模板是否已更新
		if cached.templateType != "default" && cached.updatedAt != nil {
			currentUpdatedAt := p.getTemplateUpdatedAt(configName, cached.templateType)
			if currentUpdatedAt != nil && *currentUpdatedAt == *cached.updatedAt {
				cacheMutex.RUnlock()
				return cached.config, nil
			}
			// 缓存失效，删除
			delete(configCache, configName)
		} else {
			cacheMutex.RUnlock()
			return cached.config, nil
		}
	}
	cacheMutex.RUnlock()

	// 从数据库加载
	config, updatedAt, templateType, err := p.loadFromDatabase(configName)
	if err == nil && config != nil {
		cacheMutex.Lock()
		if len(configCache) >= cacheMaxSize {
			// 删除第一个
			for k := range configCache {
				delete(configCache, k)
				break
			}
		}
		configCache[configName] = &configCacheEntry{
			config:       config,
			updatedAt:    updatedAt,
			templateType: templateType,
		}
		cacheMutex.Unlock()
		return config, nil
	}

	// 使用默认配置
	defaultConfig := getDefaultConfig()
	cacheMutex.Lock()
	if len(configCache) >= cacheMaxSize {
		for k := range configCache {
			delete(configCache, k)
			break
		}
	}
	configCache[configName] = &configCacheEntry{
		config:       defaultConfig,
		updatedAt:    nil,
		templateType: "default",
	}
	cacheMutex.Unlock()
	return defaultConfig, nil
}

// ProcessMarkdown 使用统一配置处理 Markdown 文本（带缓存）
func (p *UnifiedMarkdownProcessor) ProcessMarkdown(markdownText string, configName string) (string, error) {
	// 生成缓存键
	hash := md5.Sum([]byte(markdownText))
	cacheKey := fmt.Sprintf("%s:%x", configName, hash)

	// 检查缓存
	cacheMutex.RLock()
	if cached, ok := markdownProcessCache[cacheKey]; ok {
		cacheMutex.RUnlock()
		return cached, nil
	}
	cacheMutex.RUnlock()

	// 加载统一配置
	unifiedConfig, err := p.LoadUnifiedConfig(configName)
	if err != nil {
		unifiedConfig = getDefaultConfig()
	}

	// 执行处理
	result := p.processWithUnifiedConfig(markdownText, unifiedConfig)

	// 存入缓存
	cacheMutex.Lock()
	if len(markdownProcessCache) >= processCacheMaxSize {
		for k := range markdownProcessCache {
			delete(markdownProcessCache, k)
			break
		}
	}
	markdownProcessCache[cacheKey] = result
	cacheMutex.Unlock()

	return result, nil
}

// processWithUnifiedConfig 使用统一配置处理 Markdown 文本
func (p *UnifiedMarkdownProcessor) processWithUnifiedConfig(markdownText string, unifiedConfig map[string]interface{}) string {
	// 重置计数器
	p.mu.Lock()
	p.headingCounter = make(map[string]int)
	p.mu.Unlock()

	// 第一步：基础 Markdown 转 HTML
	themeConfig := p.extractThemeConfig(unifiedConfig)
	baseHTML := convertMarkdownToHTML(markdownText, themeConfig)

	// 第二步：应用装饰
	decoratedHTML := p.applyUnifiedDecorations(baseHTML, unifiedConfig)

	return decoratedHTML
}

// extractThemeConfig 从统一配置中提取主题配置
func (p *UnifiedMarkdownProcessor) extractThemeConfig(unifiedConfig map[string]interface{}) map[string]interface{} {
	themeConfig := make(map[string]interface{})
	if base, ok := unifiedConfig["base"].(map[string]interface{}); ok {
		themeConfig["base"] = base
	}
	if block, ok := unifiedConfig["block"].(map[string]interface{}); ok {
		themeConfig["block"] = block
	}
	if inline, ok := unifiedConfig["inline"].(map[string]interface{}); ok {
		themeConfig["inline"] = inline
	}
	return themeConfig
}

// applyUnifiedDecorations 应用统一配置中的装饰元素
func (p *UnifiedMarkdownProcessor) applyUnifiedDecorations(htmlText string, unifiedConfig map[string]interface{}) string {
	result := htmlText
	components, _ := unifiedConfig["components"].(map[string]interface{})
	rules, _ := unifiedConfig["rules"].(map[string]interface{})

	// 处理标题装饰
	for _, level := range []string{"h1", "h2", "h3", "h4", "h5", "h6"} {
		if rule, ok := rules[level].(map[string]interface{}); ok {
			if decorationName, ok := rule["decoration"].(string); ok {
				if component, ok := components[decorationName].(map[string]interface{}); ok {
					if enabled, _ := component["enabled"].(bool); enabled {
						result = p.processHeadingDecoration(result, level, component, rule)
					}
				}
			}
		}
	}

	// 处理引用装饰
	if rule, ok := rules["blockquote"].(map[string]interface{}); ok {
		if decorationName, ok := rule["decoration"].(string); ok {
			if component, ok := components[decorationName].(map[string]interface{}); ok {
				if enabled, _ := component["enabled"].(bool); enabled {
					result = p.processBlockquoteDecoration(result, component, rule)
				}
			}
		}
	}

	// 应用布局
	if layout, ok := unifiedConfig["layout"].(map[string]interface{}); ok {
		result = p.applyLayoutUnified(result, layout)
	}

	return result
}

// processHeadingDecoration 处理标题装饰
func (p *UnifiedMarkdownProcessor) processHeadingDecoration(htmlText string, headingLevel string, component map[string]interface{}, rule map[string]interface{}) string {
	if replaceOriginal, _ := rule["replace_original"].(bool); !replaceOriginal {
		return htmlText
	}

	pattern := fmt.Sprintf(`<%s[^>]*>(.*?)</%s>`, headingLevel, headingLevel)
	re := regexp.MustCompile(`(?i)` + pattern)

	template, _ := component["template"].(string)
	style, _ := component["style"].(map[string]interface{})

	return re.ReplaceAllStringFunc(htmlText, func(match string) string {
		// 提取内容
		contentRe := regexp.MustCompile(fmt.Sprintf(`<%s[^>]*>(.*?)</%s>`, headingLevel, headingLevel))
		matches := contentRe.FindStringSubmatch(match)
		if len(matches) < 2 {
			return match
		}
		content := matches[1]
		// 移除HTML标签
		content = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(content, "")

		// 自动编号
		var number string
		if autoNumber, _ := rule["auto_number"].(bool); autoNumber {
			p.mu.Lock()
			if _, ok := p.headingCounter[headingLevel]; !ok {
				p.headingCounter[headingLevel] = 0
			}
			p.headingCounter[headingLevel]++
			number = fmt.Sprintf("%02d", p.headingCounter[headingLevel])
			p.mu.Unlock()
		} else {
			number = "01"
		}

		// 渲染模板
		return p.renderTemplate(template, map[string]interface{}{
			"content": content,
			"number":  number,
			"style":   style,
		})
	})
}

// processBlockquoteDecoration 处理引用装饰
func (p *UnifiedMarkdownProcessor) processBlockquoteDecoration(htmlText string, component map[string]interface{}, rule map[string]interface{}) string {
	if replaceOriginal, _ := rule["replace_original"].(bool); !replaceOriginal {
		return htmlText
	}

	pattern := `<blockquote[^>]*>(.*?)</blockquote>`
	re := regexp.MustCompile(`(?i)` + pattern)

	template, _ := component["template"].(string)
	style, _ := component["style"].(map[string]interface{})

	return re.ReplaceAllStringFunc(htmlText, func(match string) string {
		contentRe := regexp.MustCompile(`<blockquote[^>]*>(.*?)</blockquote>`)
		matches := contentRe.FindStringSubmatch(match)
		if len(matches) < 2 {
			return match
		}
		content := matches[1]
		// 清理内容，移除内层的p标签
		content = regexp.MustCompile(`<p[^>]*>(.*?)</p>`).ReplaceAllString(content, "$1")
		content = strings.TrimSpace(content)

		return p.renderTemplate(template, map[string]interface{}{
			"content": content,
			"style":   style,
		})
	})
}

// applyLayoutUnified 应用布局配置
func (p *UnifiedMarkdownProcessor) applyLayoutUnified(htmlText string, layout map[string]interface{}) string {
	maxWidth, _ := layout["max_width"].(string)
	if maxWidth == "" {
		maxWidth = "800px"
	}
	margin, _ := layout["margin"].(string)
	if margin == "" {
		margin = "0 auto"
	}
	padding, _ := layout["padding"].(string)
	if padding == "" {
		padding = "20px"
	}

	pattern := `(<section[^>]*style="[^"]*")`
	re := regexp.MustCompile(pattern)

	layoutStyles := fmt.Sprintf("max-width: %s; margin: %s; padding: %s;", maxWidth, margin, padding)

	return re.ReplaceAllStringFunc(htmlText, func(match string) string {
		return match + layoutStyles
	})
}

// renderTemplate 渲染模板，将变量替换为实际值
func (p *UnifiedMarkdownProcessor) renderTemplate(template string, variables map[string]interface{}) string {
	result := template
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		var strValue string
		if valueMap, ok := value.(map[string]interface{}); ok {
			// 如果是map，转换为样式字符串
			styles := []string{}
			for k, v := range valueMap {
				styles = append(styles, fmt.Sprintf("%s: %v", k, v))
			}
			strValue = strings.Join(styles, "; ")
		} else {
			strValue = fmt.Sprintf("%v", value)
		}
		result = strings.ReplaceAll(result, placeholder, strValue)
	}
	return result
}

// loadFromDatabase 从数据库中加载模板配置
func (p *UnifiedMarkdownProcessor) loadFromDatabase(configName string) (map[string]interface{}, *float64, string, error) {
	// 优先从官方模板查找
	var publicTemplate models.PublicTemplate
	if err := repository.DB.Where("template_id = ?", configName).First(&publicTemplate).Error; err == nil {
		if publicTemplate.TemplateData != nil {
			var config map[string]interface{}
			if err := json.Unmarshal([]byte(*publicTemplate.TemplateData), &config); err == nil {
				config = p.ensureConfigStructure(config)
				var updatedAt *float64
				if !publicTemplate.UpdatedAt.IsZero() {
					ts := float64(publicTemplate.UpdatedAt.Unix())
					updatedAt = &ts
				}
				return config, updatedAt, "public", nil
			}
		}
	}

	// 尝试从用户模板查找
	var userTemplate models.UserTemplate
	if err := repository.DB.Where("template_id = ?", configName).First(&userTemplate).Error; err == nil {
		if userTemplate.TemplateData != nil {
			var config map[string]interface{}
			if err := json.Unmarshal([]byte(*userTemplate.TemplateData), &config); err == nil {
				config = p.ensureConfigStructure(config)
				var updatedAt *float64
				if !userTemplate.UpdatedAt.IsZero() {
					ts := float64(userTemplate.UpdatedAt.Unix())
					updatedAt = &ts
				}
				return config, updatedAt, "user", nil
			}
		}
	}

	return nil, nil, "default", gorm.ErrRecordNotFound
}

// getTemplateUpdatedAt 获取模板的更新时间戳
func (p *UnifiedMarkdownProcessor) getTemplateUpdatedAt(configName string, templateType string) *float64 {
	// 优先从官方模板查找
	var publicTemplate models.PublicTemplate
	if err := repository.DB.Where("template_id = ?", configName).First(&publicTemplate).Error; err == nil {
		if !publicTemplate.UpdatedAt.IsZero() {
			ts := float64(publicTemplate.UpdatedAt.Unix())
			return &ts
		}
	}

	// 如果官方模板不存在，再查用户模板
	if templateType == "user" {
		var userTemplate models.UserTemplate
		if err := repository.DB.Where("template_id = ?", configName).First(&userTemplate).Error; err == nil {
			if !userTemplate.UpdatedAt.IsZero() {
				ts := float64(userTemplate.UpdatedAt.Unix())
				return &ts
			}
		}
	}

	return nil
}

// ensureConfigStructure 确保模板配置包含必要的结构
func (p *UnifiedMarkdownProcessor) ensureConfigStructure(templateData map[string]interface{}) map[string]interface{} {
	if _, ok := templateData["components"]; !ok {
		templateData["components"] = make(map[string]interface{})
	}
	if _, ok := templateData["rules"]; !ok {
		templateData["rules"] = make(map[string]interface{})
	}
	if _, ok := templateData["layout"]; !ok {
		templateData["layout"] = make(map[string]interface{})
	}
	return templateData
}

// getDefaultConfig 获取默认配置
func getDefaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"meta": map[string]interface{}{
			"name":        "默认统一配置",
			"description": "默认的统一配置，包含基础样式和简单装饰",
			"version":     "1.0.0",
			"type":        "unified",
		},
		"base": map[string]interface{}{
			"text-align":  "left",
			"line-height": "1.75",
			"color":       "#333",
			"font-family": "Arial, sans-serif",
			"font-size":   "16px",
		},
		"block": map[string]interface{}{
			"h1": map[string]interface{}{
				"font-size":   "24px",
				"font-weight": "bold",
				"color":       "#333",
				"margin":      "1em 0",
			},
			"h2": map[string]interface{}{
				"font-size":   "20px",
				"font-weight": "bold",
				"color":       "#333",
				"margin":      "1em 0",
			},
			"p": map[string]interface{}{
				"margin":      "1em 0",
				"line-height": "1.6",
			},
		},
		"inline": map[string]interface{}{
			"strong": map[string]interface{}{
				"font-weight": "bold",
			},
			"em": map[string]interface{}{
				"font-style": "italic",
			},
		},
		"components": make(map[string]interface{}),
		"rules":      make(map[string]interface{}),
	}
}

// convertMarkdownToHTML 将 Markdown 转换为 HTML（简化版）
func convertMarkdownToHTML(markdownText string, themeConfig map[string]interface{}) string {
	if markdownText == "" {
		return ""
	}

	// 简化版：基本的markdown转换
	// 这里可以使用更完整的markdown库，但为了简化，先实现基本功能
	lines := strings.Split(markdownText, "\n")
	var result strings.Builder
	result.WriteString("<section style=\"text-align: left; line-height: 1.75; color: #333;\">\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 处理标题
		if strings.HasPrefix(line, "# ") {
			content := strings.TrimPrefix(line, "# ")
			result.WriteString(fmt.Sprintf("<h1>%s</h1>\n", escapeHTML(content)))
		} else if strings.HasPrefix(line, "## ") {
			content := strings.TrimPrefix(line, "## ")
			result.WriteString(fmt.Sprintf("<h2>%s</h2>\n", escapeHTML(content)))
		} else if strings.HasPrefix(line, "### ") {
			content := strings.TrimPrefix(line, "### ")
			result.WriteString(fmt.Sprintf("<h3>%s</h3>\n", escapeHTML(content)))
		} else if strings.HasPrefix(line, "> ") {
			// 引用
			content := strings.TrimPrefix(line, "> ")
			result.WriteString(fmt.Sprintf("<blockquote>%s</blockquote>\n", processInlineMarkdown(content)))
		} else if strings.HasPrefix(line, "- ") {
			// 列表
			content := strings.TrimPrefix(line, "- ")
			result.WriteString(fmt.Sprintf("<ul><li>%s</li></ul>\n", processInlineMarkdown(content)))
		} else {
			// 段落
			result.WriteString(fmt.Sprintf("<p>%s</p>\n", processInlineMarkdown(line)))
		}
	}

	result.WriteString("</section>")
	return result.String()
}

// processInlineMarkdown 处理行内markdown（粗体、斜体、代码等）
func processInlineMarkdown(text string) string {
	// 处理粗体 **text**
	text = regexp.MustCompile(`\*\*(.*?)\*\*`).ReplaceAllString(text, "<strong>$1</strong>")
	// 处理斜体 *text*
	text = regexp.MustCompile(`\*(.*?)\*`).ReplaceAllString(text, "<em>$1</em>")
	// 处理代码 `code`
	text = regexp.MustCompile("`([^`]+)`").ReplaceAllString(text, "<code>$1</code>")
	return text
}

// escapeHTML 转义HTML
func escapeHTML(text string) string {
	text = strings.ReplaceAll(text, "&", "&amp;")
	text = strings.ReplaceAll(text, "<", "&lt;")
	text = strings.ReplaceAll(text, ">", "&gt;")
	text = strings.ReplaceAll(text, "\"", "&quot;")
	text = strings.ReplaceAll(text, "'", "&#39;")
	return text
}

// GetThumbnailContent 获取缩略图预览内容
func GetThumbnailContent() string {
	// 尝试从配置文件读取
	configPath := filepath.Join("configs", "docs", "theme_thumbnail_content.md")
	if content, err := os.ReadFile(configPath); err == nil {
		return string(content)
	}
	// 默认内容
	return "# 主题预览\n\n这是**主题效果**的展示。"
}
