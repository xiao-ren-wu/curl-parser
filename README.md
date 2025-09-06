# curl-parser

[![Go Report Card](https://goreportcard.com/badge/github.com/xiao-ren-wu/curl-parser)](https://goreportcard.com/report/github.com/xiao-ren-wu/curl-parser)
[![GoDoc](https://pkg.go.dev/badge/github.com/xiao-ren-wu/curl-parser?utm_source=godoc)](https://pkg.go.dev/github.com/xiao-ren-wu/curl-parser)
![GitHub](https://img.shields.io/github/license/xiao-ren-wu/curl-parser)

一个用 Go 语言编写的 curl 命令解析器，可以将 curl 命令转换为 Go 的 HTTP 请求结构体。

## 功能特性

- 解析 curl 命令中的 URL
- 提取 HTTP 方法（GET、POST、PUT、DELETE 等）
- 解析请求头（Headers）
- 解析请求体（Body）
- 解析 URL 查询参数
- 支持多种 curl 参数格式
- 支持多行 curl 命令
- 支持 Cookie 解析

## 安装

```bash
go get github.com/xiao-ren-wu/curl-parser
```

## 使用方法

### 基本用法

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/xiao-ren-wu/curl-parser"
)

func main() {
    curlCommand := `curl -X POST -H "Content-Type: application/json" -d '{"key":"value"}' https://httpbin.org/post`
    
    parser := curl_parser.NewCurlParser(curlCommand)
    request, err := parser.Parse()
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Method: %s\n", request.Method)
    fmt.Printf("URL: %s\n", request.URL)
    fmt.Printf("BaseURL: %s\n", request.BaseURL)
    fmt.Printf("Path: %s\n", request.Path)
    fmt.Printf("Body: %s\n", request.Body)
    
    // 打印 Headers
    for key, value := range request.Headers {
        fmt.Printf("Header: %s = %s\n", key, value)
    }
    
    // 打印 Query 参数
    for key, value := range request.Query {
        fmt.Printf("Query: %s = %s\n", key, value)
    }
}
```

### 支持的 curl 参数

| 参数 | 描述 | 示例 |
|------|------|------|
| `-X`, `--request` | 指定 HTTP 方法 | `curl -X POST https://httpbin.org/post` |
| `-H`, `--header` | 添加请求头 | `curl -H "Content-Type: application/json" https://httpbin.org/get` |
| `-d`, `--data` | 发送数据体 | `curl -d "key=value" https://httpbin.org/post` |
| `--data-raw` | 发送原始数据 | `curl --data-raw '{"key": "value"}' https://httpbin.org/post` |
| `-F`, `--form` | 发送表单数据 | `curl -F "key=value" https://httpbin.org/post` |

### 多行命令支持

```go
curlCommand := `curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"key":"value"}' \
  https://httpbin.org/post`

parser := curl_parser.NewCurlParser(curlCommand)
request, err := parser.Parse()
```

## 返回结构

解析器返回一个 [HTTPRequest](file:///Users/erik/Desktop/curl-parser/parser.go#L12-L22) 结构体：

```go
type HTTPRequest struct {
    Method  string            // HTTP 方法
    URL     string            // 完整 URL
    BaseURL string            // 基础 URL (协议 + 主机)
    Path    string            // 路径
    Headers map[string]string // 请求头
    Body    string            // 请求体
    Query   map[string]string // 查询参数
}
```

## 测试

运行测试：

```bash
go test -v
```

## 贡献

欢迎提交 Issue 和 Pull Request。

## 许可证

[MIT](LICENSE)