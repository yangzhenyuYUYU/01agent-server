package tools

import (
	"encoding/json"
	"regexp"
	"strings"
)

// ExtractFirstImageFromMarkdown 从 markdown 内容或 HTML 中提取第一张图片的 URL
// 支持:
// 1. Markdown格式: ![替代文本](图片URL "可选标题")
// 2. HTML img标签: <img src="图片URL" ... />
func ExtractFirstImageFromMarkdown(content string) string {
	if content == "" {
		return ""
	}

	// 首先尝试提取 HTML img 标签中的 src 属性
	htmlImgPattern := regexp.MustCompile(`<img[^>]+src\s*=\s*["']([^"']+)["']`)
	if htmlMatch := htmlImgPattern.FindStringSubmatch(content); htmlMatch != nil {
		url := strings.TrimSpace(htmlMatch[1])
		if url != "" {
			return url
		}
	}

	// 如果没有找到 HTML 图片，尝试匹配 Markdown 格式
	// 匹配有标题的格式: ![alt](url "title")
	titlePattern := regexp.MustCompile(`!\[([^\]]*)\]\(([^"]+?)\s+"[^"]*"\)`)
	if titleMatch := titlePattern.FindStringSubmatch(content); titleMatch != nil {
		url := strings.TrimSpace(titleMatch[2])
		if url != "" {
			return url
		}
	}

	// 匹配简单格式: ![alt](url)
	simplePattern := regexp.MustCompile(`!\[([^\]]*)\]\(([^)]*)\)`)
	if matches := simplePattern.FindAllStringSubmatch(content, -1); len(matches) > 0 {
		for _, match := range matches {
			fullURLPart := strings.TrimSpace(match[2])
			if fullURLPart == "" {
				continue
			}

			// 检查是否包含标题（以引号开始的部分）
			if idx := strings.LastIndex(fullURLPart, `"`); idx > 0 {
				fullURLPart = strings.TrimSpace(fullURLPart[:idx])
			}

			if fullURLPart != "" {
				return fullURLPart
			}
		}
	}

	return ""
}

// ExtractThumbnailFromImages 从 images 字段中提取第一张图片的 URL
// 支持多种数据结构格式
func ExtractThumbnailFromImages(images interface{}) string {
	if images == nil {
		return ""
	}

	switch v := images.(type) {
	case []interface{}:
		if len(v) > 0 {
			if dict, ok := v[0].(map[string]interface{}); ok {
				if imageURL, exists := dict["imageUrl"]; exists {
					if url, ok := imageURL.(string); ok {
						return url
					}
				}
			} else if str, ok := v[0].(string); ok {
				return str
			}
		}

	case map[string]interface{}:
		// 处理嵌套字典的情况
		for _, value := range v {
			if arr, ok := value.([]interface{}); ok && len(arr) > 0 {
				if dict, ok := arr[0].(map[string]interface{}); ok {
					if imageURL, exists := dict["imageUrl"]; exists {
						if url, ok := imageURL.(string); ok {
							return url
						}
					}
				}
			}
		}

	case string:
		// 尝试解析 JSON 字符串
		var imagesList interface{}
		if err := json.Unmarshal([]byte(v), &imagesList); err == nil {
			return ExtractThumbnailFromImages(imagesList)
		}
	}

	return ""
}

// ProcessSingleRecordThumbnail 处理单条记录的缩略图提取
func ProcessSingleRecordThumbnail(recordContent string, recordImages interface{}) string {
	// 首先尝试从 markdown/HTML 内容中提取图片
	if recordContent != "" {
		if thumbnail := ExtractFirstImageFromMarkdown(recordContent); thumbnail != "" {
			return thumbnail
		}
	}

	// 如果从内容中提取不到图片，再从 images 字段中获取
	return ExtractThumbnailFromImages(recordImages)
}

// IsDefaultWelcomeContent 检查内容是否为默认的欢迎内容
func IsDefaultWelcomeContent(content string) bool {
	if content == "" {
		return false
	}

	// 去除首尾空白字符
	content = strings.TrimSpace(content)

	// 检查是否包含关键的默认内容标识
	defaultKeywords := []string{
		"# 欢迎使用01Editor",
		"这是一个智能创作编辑器",
		"支持Markdown格式和多种样式",
		"## 快速开始",
		"在左侧选择创作选题获取灵感",
		"使用内容创作面板进行写作",
		"通过样式排版美化文档",
		"导出为多种格式分享",
		"开始您的创作之旅吧！✨",
	}

	// 如果内容包含多个默认关键词，则判断为默认内容
	keywordCount := 0
	for _, keyword := range defaultKeywords {
		if strings.Contains(content, keyword) {
			keywordCount++
		}
	}

	// 如果包含 3 个或以上的关键词，则认为是默认内容
	if keywordCount >= 3 {
		return true
	}

	// 如果以"# 欢迎使用01Editor"开头，直接判断为默认内容
	if strings.HasPrefix(content, "# 欢迎使用01Editor") {
		return true
	}

	return false
}
