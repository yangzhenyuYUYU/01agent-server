package tools

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"net/url"

	"encoding/hex"

	"github.com/google/uuid"
)

/*
* accessKey 联系技术支持获取
* accessSecret 联系技术支持获取
* method 请求方法：POST
* path URI，比如 /v1/open/visitor/register
* query URL 地址上的参数 a=xx&b=xx
* body 请求体 json 字符串的二进制数据
 */
func signRequest(accessKey, accessSecret, method, path string, query url.Values, body []byte) (headers map[string]string) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := uuid.New().String()

	// 构造 param string（含 query + auth params）
	params := make(url.Values)
	for k, v := range query {
		params[k] = v
	}
	params.Set("accessKey", accessKey)
	params.Set("timestamp", timestamp)
	params.Set("nonce", nonce)

	// 排序
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var paramStr strings.Builder
	for i, k := range keys {
		if i > 0 {
			paramStr.WriteString("&")
		}
		paramStr.WriteString(fmt.Sprintf("%s=%s", k, params.Get(k)))
	}

	// Body SHA256
	bodyHash := sha256Hex(body)

	stringToSign := fmt.Sprintf("%s\n%s\n%s\n%s", method, path, paramStr.String(), bodyHash)
	mac := hmac.New(sha256.New, []byte(accessSecret))
	mac.Write([]byte(stringToSign))
	signature := hex.EncodeToString(mac.Sum(nil))

	return map[string]string{
		"X-ACCESS-KEY": accessKey,
		"X-Timestamp":  timestamp,
		"X-Nonce":      nonce,
		"X-Signature":  signature,
	}
}

// 计算 SHA256 摘要（十六进制小写）
func sha256Hex(data []byte) string {
	h := sha256.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
