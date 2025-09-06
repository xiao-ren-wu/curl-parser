package curl_parser

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// HTTPRequest 表示解析后的HTTP请求结构
type HTTPRequest struct {
	Method string
	// example: https://www.example.foo/bar?a=1&b=2
	URL string
	// example: https://www.example.foo/
	BaseURL string
	// example: /bar
	Path    string
	Headers map[string]string
	Body    string
	Query   map[string]string
}

// CurlParser curl解析器
type CurlParser struct {
	curlCommand string
}

// NewCurlParser 创建新的curl解析器
func NewCurlParser(curlCommand string) *CurlParser {
	return &CurlParser{
		curlCommand: curlCommand,
	}
}

// Parse 解析curl命令并返回HTTPRequest结构
func (cp *CurlParser) Parse() (*HTTPRequest, error) {
	req := &HTTPRequest{
		Headers: make(map[string]string),
		Query:   make(map[string]string),
	}

	// 清理curl命令，移除多余的空白字符和换行符
	cmd := strings.ReplaceAll(cp.curlCommand, "\\\n", " ")
	cmd = strings.ReplaceAll(cmd, "\\", "")
	cmd = strings.TrimSpace(cmd)

	// 移除开头的curl
	if strings.HasPrefix(cmd, "curl ") {
		cmd = strings.TrimPrefix(cmd, "curl ")
	}

	// 解析URL
	urlStr, err := cp.extractURL(cmd)
	if err != nil {
		return nil, fmt.Errorf("解析URL失败: %v", err)
	}
	req.URL = urlStr

	// 解析BaseURL和Path
	parsedURL, err := url.Parse(urlStr)
	if err == nil {
		req.BaseURL = fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
		req.Path = parsedURL.Path
	}

	// 解析HTTP方法
	req.Method = cp.extractMethod(cmd)

	// 解析Headers
	cp.extractHeaders(cmd, req)

	// 解析Body
	req.Body = cp.extractBody(cmd)

	// 解析Query参数
	cp.extractQueryParams(req)

	return req, nil
}

// extractURL 提取URL
func (cp *CurlParser) extractURL(cmd string) (string, error) {
	// 查找第一个URL（以http://或https://开头）
	// 改进的正则表达式，更好地处理引号包围的URL
	urlRegex := regexp.MustCompile(`(?:"|'|)(https?://[^\s"']+)("|'|)`)
	matches := urlRegex.FindStringSubmatch(cmd)
	if len(matches) > 1 {
		return matches[1], nil
	}

	// 如果上面的方法失败，尝试更宽松的匹配
	urlRegex = regexp.MustCompile(`(https?://[^\s"']+)`)
	matches = urlRegex.FindStringSubmatch(cmd)
	if len(matches) > 1 {
		return matches[1], nil
	}

	return "", fmt.Errorf("未找到有效的URL")
}

// extractMethod 提取HTTP方法
func (cp *CurlParser) extractMethod(cmd string) string {
	// 检查是否有-X参数指定方法
	methodRegex := regexp.MustCompile(`-X\s+(\w+)`)
	matches := methodRegex.FindStringSubmatch(cmd)
	if len(matches) > 1 {
		return strings.ToUpper(matches[1])
	}

	// 检查是否有--request参数
	requestRegex := regexp.MustCompile(`--request\s+(\w+)`)
	matches = requestRegex.FindStringSubmatch(cmd)
	if len(matches) > 1 {
		return strings.ToUpper(matches[1])
	}

	// 检查是否有特定参数（表示POST请求）
	if strings.Contains(cmd, "--data") || strings.Contains(cmd, "-d") {
		return "POST"
	}

	// 检查是否有文件上传相关参数
	if strings.Contains(cmd, "--form") || strings.Contains(cmd, "-F") {
		return "POST"
	}

	// 默认返回GET
	return "GET"
}

// extractHeaders 提取请求头
func (cp *CurlParser) extractHeaders(cmd string, req *HTTPRequest) {
	// 匹配 -H 或 --header 参数，支持多种格式
	// 1. 单引号包围: -H 'Content-Type: application/json'
	// 2. 双引号包围: -H "Content-Type: application/json"
	// 3. 无引号: -H Content-Type:application/json
	headerRegex := regexp.MustCompile(`(?:-H|--header)\s+(?:'([^']*)'|"([^"]*)"|([^\s]+))`)
	matches := headerRegex.FindAllStringSubmatch(cmd, -1)

	for _, match := range matches {
		if len(match) > 3 {
			// 获取非空的匹配组
			header := ""
			for i := 1; i <= 3; i++ {
				if match[i] != "" {
					header = match[i]
					break
				}
			}

			if header != "" {
				parts := strings.SplitN(header, ":", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					req.Headers[key] = value
				}
			}
		}
	}
}

// parseCookieString 解析cookie字符串
func (cp *CurlParser) parseCookieString(cookieStr string, cookies map[string]string) {
	// 分割cookie字符串
	pairs := strings.Split(cookieStr, ";")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			cookies[key] = value
		} else if len(kv) == 1 && kv[0] != "" {
			// 只有键没有值的情况
			cookies[strings.TrimSpace(kv[0])] = ""
		}
	}
}

// extractBody 提取请求体 - 改进版本
func (cp *CurlParser) extractBody(cmd string) string {
	// 先尝试匹配单引号包围的JSON
	singleQuoteRegex := regexp.MustCompile(`(?:--data|-d)\s+'([^']*)'`)
	matches := singleQuoteRegex.FindStringSubmatch(cmd)
	if len(matches) > 1 {
		return matches[1]
	}

	// 再尝试匹配双引号包围的JSON
	doubleQuoteRegex := regexp.MustCompile(`(?:--data|-d)\s+"([^"]*)"`)
	matches = doubleQuoteRegex.FindStringSubmatch(cmd)
	if len(matches) > 1 {
		return matches[1]
	}

	// 匹配 --data-raw 参数
	dataRawRegex := regexp.MustCompile(`--data-raw\s+['"]?(.*?)['"]?$`)
	matches = dataRawRegex.FindStringSubmatch(cmd)
	if len(matches) > 1 {
		return matches[1]
	}

	// 匹配无引号的数据
	noQuoteRegex := regexp.MustCompile(`(?:--data|-d)\s+([^\s]+)`)
	matches = noQuoteRegex.FindStringSubmatch(cmd)
	if len(matches) > 1 {
		return matches[1]
	}

	// 匹配 --form 参数
	formRegex := regexp.MustCompile(`(?:--form|-F)\s+(?:'([^']*)'|"([^"]*)"|([^\s]+))`)
	formMatches := formRegex.FindAllStringSubmatch(cmd, -1)
	if len(formMatches) > 0 {
		// 构建form数据
		var formData []string
		for _, match := range formMatches {
			if len(match) > 3 {
				formDataStr := ""
				for i := 1; i <= 3; i++ {
					if match[i] != "" {
						formDataStr = match[i]
						break
					}
				}
				if formDataStr != "" {
					formData = append(formData, formDataStr)
				}
			}
		}
		return strings.Join(formData, "&")
	}

	return ""
}

// extractQueryParams 从URL中提取查询参数
func (cp *CurlParser) extractQueryParams(req *HTTPRequest) {
	parsedURL, err := url.Parse(req.URL)
	if err != nil {
		return
	}

	query := parsedURL.Query()
	for key, values := range query {
		if len(values) > 0 {
			req.Query[key] = values[0]
		}
	}
}
