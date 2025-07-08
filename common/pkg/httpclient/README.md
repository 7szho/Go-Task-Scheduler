实现了一个简单的HTTP客户端工具包, 封装了三种常见的HTTP请求操作, 让其他代码可以方便地调用, 这三种操作分别是:  
1. 发起 HTTP GET 请求
2. 发起 HTTP POST 请求, 并以表单(from)格式提交数据
3. 发起 HTTP POST 请求, 并以 JSON 格式提交数据  

---

#### `Get(url string, timeout int64)` 函数  
- 作用: 发起一个标准的 HTTP GET 请求
- 输入:
    1. `url`: 要请求的目标网址
    2. `timeout`: 请求超时时间
- 流程:
    1. 创建一个 HTTP 客户端
    2. 设置超时时间
    3. 发送 GET 请求
    4. 检查服务器返回状态码是否为200(表示成功). 如果不是, 返回错误
    5. 读取服务器返回数据
- 输出:
    1. `result`: 服务器返回的响应内容(字符串格式)
    2. `err`: 错误信息

#### `PostParams(url string, params string, timeout int64) (result string, err error)` 函数   
- 作用: 模拟网页提交表单的行为, 发起一个 POST 请求
- 输入:
    1. `url`: 要请求的目标网址
    2. `params`: 要提交的表单数据, 格式为"key1=value1&key2=value2"
    3. `timeout`: 请求超时时间
- 流程:
    1. 创建一个 HTTP 客户端
    2. 将 params 字符串作为请求体
    3. 在请求体中设置 Content-Type 为`application/x-www-form-urlencoded`, 告诉服务器这是一个表单提交
    4. 发送 POST 请求
    5. 处理响应
- 输出:
    1. `result`: 服务器返回的响应内容(字符串格式)
    2. `err`: 错误信息

#### `PostJson(url string, body string, timeout int64) (result string, err error)` 函数  
- 作用: 向服务器发送 JSON 数据, 常用于调用API接口
- 输入:
    1. `url`: 要请求的目标网址
    2. `body`: 要提交的 JSON 数据, 格式为标准的 JSON 字符串
    3. `timeout`: 请求超时时间
- 流程:
    1. 创建一个 HTTP 客户端
    2. 将 body JSON 字符串作为请求体
    3. 在请求体中设置 Content-Type 为`appplication/json`, 告诉服务器这是一个表单提交
    4. 发送 POST 请求
    5. 处理响应
- 输出:
    1. `result`: 服务器返回的响应内容(字符串格式)
    2. `err`: 错误信息