package httpclient

import (
	"bytes"
	"crony/common/pkg/logger"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Get 函数用于发起一个 HTTP GET 请求
// url: 目标请求的 URL
// timeout: 请求的超时时间
func Get(url string, timeout int64) (result string, err error) {
	// 创建一个 HTTP 客户端实例
	var client = &http.Client{}
	// 使用 "GET" 方法和指定的 URL 创建一个新的 HTTP 请求对象
	// 第三个参数时请求体
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	// 如果传入的 timeout 大于 0, 则为这个客户端设置超时时间
	if timeout > 0 {
		client.Timeout = time.Duration(timeout) * time.Second
	}
	// 使用客户端发送上面创建的请求
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	// 检查响应的状态码
	if resp.StatusCode != 200 {
		err = fmt.Errorf("response status code is not 200")
		return
	}
	// 读取响应体的所有内容
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		// 如果读取响应体时发生错误, 记录一条警告日志并返回错误
		logger.GetLogger().Warn(fmt.Sprintf("http get api url:%s send err: %s", url, err.Error()))
		return
	}
	// 将读取到的字节切片 ([]bytes) 转换为字符串
	result = string(data)
	return
}

// PostParams 函数用于发起一个 HTTP POST 请求, 内容类型为 "application/x-www-form-urlencoded"
// url: 目标请求的 URL
// params: POST 请求的表单参数, 格式应为 "key1=value&key2=value2"
// timeout: 请求的超时时间
func PostParams(url string, params string, timeout int64) (result string, err error) {
	// 为每个请求创建一个新的 Client
	var client = &http.Client{}
	// 将参数字符串 `params` 包装成一个 `bytes.Buffer`,
	// 实现了 `io.Reader` 接口, 可以作为请求体
	buf := bytes.NewBufferString(params)
	// 使用 "POST" 方法, URL, 和请求体来创建一个新的 HTTP 请求对象
	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return
	}
	// 设置请求头(Header), 指明请求体的内容类型
	// "application/x-www-form-urlencoded" 是网页表单提交时最常用的格式
	req.Header.Set("Content-type", "application/x-www-form-urlencoded")
	if timeout > 0 {
		client.Timeout = time.Duration(timeout) * time.Second
	}
	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	// 检查响应状态码
	if resp.StatusCode != 200 {
		err = fmt.Errorf("response status code is not 200")
		return
	}
	// 读取响应体
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("http post api url:%s send send: %s", url, err.Error()))
		return
	}
	// 将响应体字节切片转换为字符串
	result = string(data)
	return
}

func PostJson(url string, body string, timeout int64) (result string, err error) {
	// 创建一个新的 http.Client 实例
	var client = &http.Client{}
	// 将输入的 JSON 字符串 `body` 转换成一个 `io.Reader`
	buf := bytes.NewBufferString(body)
	// 使用 "POST" 方法, URL, 和请求体(buf)创建一个新的 HTTP 请求对象
	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return
	}
	// 设置请求头(Header), 指明请求头的内容类型
	req.Header.Set("Content-type", "application/json")
	if timeout > 0 {
		client.Timeout = time.Duration(timeout) * time.Second
	}
	// 使用客户端发送请求, 并获得响应
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = fmt.Errorf("response status code is not 200")
		return
	}
	// 读取响应体
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("http post api url: %s send err: %s", url, err.Error()))
		return
	}
	result = string(data)
	return
}
