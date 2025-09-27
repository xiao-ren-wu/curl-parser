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
	// 原始Cookie字符串，例如: "name1=value1; name2=value2"
	RawCookie string
	// 解析后的Cookie键值对
	ParsedCookies map[string]string
	// User-Agent字符串
	UserAgent string
	// 认证信息 (username:password)
	Auth string
	// Referer头
	Referer string
	// 代理服务器
	Proxy string
	// 连接超时时间（秒）
	ConnectTimeout int
	// 最大请求时间（秒）
	MaxTime int
	// SSL选项
	Insecure bool
	// CA证书文件
	CACert string
	// Cookie文件路径
	CookieJar string
	// 是否跟随重定向
	FollowRedirects bool
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
		Headers:       make(map[string]string),
		Query:         make(map[string]string),
		ParsedCookies: make(map[string]string),
	}

	// 清理curl命令，移除多余的空白字符和换行符
	// cmd := strings.ReplaceAll(cp.curlCommand, "\\\n", " ")
	// cmd = strings.ReplaceAll(cmd, "\\", "")
	// cmd = strings.TrimSpace(cmd)
	cmd := cp.curlCommand

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

	// 解析Cookie
	cp.extractCookies(req)

	// 解析其他参数
	cp.extractUserAgent(cmd, req)
	cp.extractAuth(cmd, req)
	cp.extractReferer(cmd, req)
	cp.extractProxy(cmd, req)
	cp.extractTimeouts(cmd, req)
	cp.extractSSLOptions(cmd, req)
	cp.extractCookieJar(cmd, req)
	cp.extractFollowRedirects(cmd, req)

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

// extractCookies 从Headers中提取并解析Cookie
func (cp *CurlParser) extractCookies(req *HTTPRequest) {
	// 首先尝试从 -b 或 --cookie 参数中提取Cookie
	cookieData := cp.extractCookieFromParams(req)

	// 如果没有从参数中找到，则从Headers中获取Cookie头
	if cookieData == "" {
		cookieHeader, exists := req.Headers["Cookie"]
		if !exists {
			// 也检查小写的cookie头
			cookieHeader, exists = req.Headers["cookie"]
		}
		cookieData = cookieHeader
	}

	if cookieData == "" {
		return
	}

	// 保存原始Cookie字符串
	req.RawCookie = cookieData

	// 解析Cookie键值对
	// Cookie格式: "name1=value1; name2=value2; name3=value3"
	cookiePairs := strings.Split(cookieData, ";")

	for _, pair := range cookiePairs {
		// 去除前后空白字符
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		// 分割键值对
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			req.ParsedCookies[key] = value
		}
	}
}

// extractCookieFromParams 从 -b 或 --cookie 参数中提取Cookie数据
func (cp *CurlParser) extractCookieFromParams(req *HTTPRequest) string {
	cmd := cp.curlCommand

	// 匹配 -b 或 --cookie 参数
	// 支持多种格式：
	// 1. -b "name1=value1; name2=value2"
	// 2. -b 'name1=value1; name2=value2'
	// 3. --cookie "name1=value1; name2=value2"
	// 4. --cookie 'name1=value1; name2=value2'
	// 5. -b name1=value1;name2=value2 (无引号)

	// 匹配单引号包围的cookie
	singleQuoteRegex := regexp.MustCompile(`(?:-b|--cookie)\s+'([^']*)'`)
	matches := singleQuoteRegex.FindStringSubmatch(cmd)
	if len(matches) > 1 {
		return matches[1]
	}

	// 匹配双引号包围的cookie
	doubleQuoteRegex := regexp.MustCompile(`(?:-b|--cookie)\s+"([^"]*)"`)
	matches = doubleQuoteRegex.FindStringSubmatch(cmd)
	if len(matches) > 1 {
		return matches[1]
	}

	// 匹配无引号的cookie（到下一个参数或行尾）
	noQuoteRegex := regexp.MustCompile(`(?:-b|--cookie)\s+([^\s-][^\s]*(?:\s+[^\s-][^\s]*)*?)(?:\s+-|$|\s+https?://)`)
	matches = noQuoteRegex.FindStringSubmatch(cmd)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	return ""
}

// extractUserAgent 提取User-Agent
func (cp *CurlParser) extractUserAgent(cmd string, req *HTTPRequest) {
	// 匹配 -A 或 --user-agent 参数
	// 支持格式: -A "Mozilla/5.0" 或 --user-agent 'Custom Agent'
	userAgentRegex := regexp.MustCompile(`(?:-A|--user-agent)\s+(?:'([^']*)'|"([^"]*)"|([^\s-][^\s]*(?:\s+[^\s-][^\s]*)*?)(?:\s+-|$|\s+https?://))`)
	matches := userAgentRegex.FindStringSubmatch(cmd)
	if len(matches) > 3 {
		// 获取非空的匹配组
		for i := 1; i <= 3; i++ {
			if matches[i] != "" {
				req.UserAgent = strings.TrimSpace(matches[i])
				break
			}
		}
	}
}

// extractAuth 提取认证信息
func (cp *CurlParser) extractAuth(cmd string, req *HTTPRequest) {
	// 匹配 -u 或 --user 参数
	// 支持格式: -u "username:password" 或 --user admin:secret
	authRegex := regexp.MustCompile(`(?:-u|--user)\s+(?:'([^']*)'|"([^"]*)"|([^\s-][^\s]*(?:\s+[^\s-][^\s]*)*?)(?:\s+-|$|\s+https?://))`)
	matches := authRegex.FindStringSubmatch(cmd)
	if len(matches) > 3 {
		// 获取非空的匹配组
		for i := 1; i <= 3; i++ {
			if matches[i] != "" {
				req.Auth = strings.TrimSpace(matches[i])
				break
			}
		}
	}
}

// extractReferer 提取Referer
func (cp *CurlParser) extractReferer(cmd string, req *HTTPRequest) {
	// 匹配 --referer 参数
	// 支持格式: --referer "https://example.com" 或 --referer 'https://example.com'
	refererRegex := regexp.MustCompile(`--referer\s+(?:'([^']*)'|"([^"]*)"|([^\s-][^\s]*(?:\s+[^\s-][^\s]*)*?)(?:\s+-|$))`)
	matches := refererRegex.FindStringSubmatch(cmd)
	if len(matches) > 3 {
		// 获取非空的匹配组
		for i := 1; i <= 3; i++ {
			if matches[i] != "" {
				req.Referer = strings.TrimSpace(matches[i])
				break
			}
		}
	}
}

// extractProxy 提取代理信息
func (cp *CurlParser) extractProxy(cmd string, req *HTTPRequest) {
	// 匹配 --proxy 参数
	// 支持格式: --proxy "http://proxy:8080" 或 --proxy 'socks5://proxy:1080'
	proxyRegex := regexp.MustCompile(`--proxy\s+(?:'([^']*)'|"([^"]*)"|([^\s-][^\s]*(?:\s+[^\s-][^\s]*)*?)(?:\s+-|$))`)
	matches := proxyRegex.FindStringSubmatch(cmd)
	if len(matches) > 3 {
		// 获取非空的匹配组
		for i := 1; i <= 3; i++ {
			if matches[i] != "" {
				req.Proxy = strings.TrimSpace(matches[i])
				break
			}
		}
	}
}

// extractTimeouts 提取超时设置
func (cp *CurlParser) extractTimeouts(cmd string, req *HTTPRequest) {
	// 匹配 --connect-timeout 参数
	connectTimeoutRegex := regexp.MustCompile(`--connect-timeout\s+(\d+)`)
	matches := connectTimeoutRegex.FindStringSubmatch(cmd)
	if len(matches) > 1 {
		if timeout, err := fmt.Sscanf(matches[1], "%d", &req.ConnectTimeout); err == nil && timeout == 1 {
			// 成功解析
		}
	}

	// 匹配 --max-time 参数
	maxTimeRegex := regexp.MustCompile(`--max-time\s+(\d+)`)
	matches = maxTimeRegex.FindStringSubmatch(cmd)
	if len(matches) > 1 {
		if timeout, err := fmt.Sscanf(matches[1], "%d", &req.MaxTime); err == nil && timeout == 1 {
			// 成功解析
		}
	}
}

// extractSSLOptions 提取SSL选项
func (cp *CurlParser) extractSSLOptions(cmd string, req *HTTPRequest) {
	// 检查 --insecure 参数
	if strings.Contains(cmd, "--insecure") {
		req.Insecure = true
	}

	// 匹配 --cacert 参数
	cacertRegex := regexp.MustCompile(`--cacert\s+(?:'([^']*)'|"([^"]*)"|([^\s-][^\s]*(?:\s+[^\s-][^\s]*)*?)(?:\s+-|$|\s+https?://))`)
	matches := cacertRegex.FindStringSubmatch(cmd)
	if len(matches) > 3 {
		// 获取非空的匹配组
		for i := 1; i <= 3; i++ {
			if matches[i] != "" {
				req.CACert = strings.TrimSpace(matches[i])
				break
			}
		}
	}
}

// extractCookieJar 提取Cookie文件路径
func (cp *CurlParser) extractCookieJar(cmd string, req *HTTPRequest) {
	// 匹配 -c 或 --cookie-jar 参数
	cookieJarRegex := regexp.MustCompile(`(?:-c|--cookie-jar)\s+(?:'([^']*)'|"([^"]*)"|([^\s-][^\s]*(?:\s+[^\s-][^\s]*)*?)(?:\s+-|$|\s+https?://))`)
	matches := cookieJarRegex.FindStringSubmatch(cmd)
	if len(matches) > 3 {
		// 获取非空的匹配组
		for i := 1; i <= 3; i++ {
			if matches[i] != "" {
				req.CookieJar = strings.TrimSpace(matches[i])
				break
			}
		}
	}
}

// extractFollowRedirects 提取重定向设置
func (cp *CurlParser) extractFollowRedirects(cmd string, req *HTTPRequest) {
	// 检查 -L 或 --location 参数
	if strings.Contains(cmd, "-L") || strings.Contains(cmd, "--location") {
		req.FollowRedirects = true
	}
}
