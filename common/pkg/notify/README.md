notify 包是项目中负责发送通知的核心模块. 它的主要职责是接收来自系统其他部分的通知请求, 并同通过不同的渠道将这些消息发送给指定的用户. 这个包的设计采用了经典的生产者-消费者模式, 通过一个带缓冲的 channel 作为消息队列, 实现了通知发送的异步化和解耦. 这确保了用于主流程不会因为发送通知的网络延迟而被阻塞. 在项目中的作用:

---

#### `type Noticer interface {SendMsg(*Message)}` 接口  
- 作用: 定义了一个通知器的行为契约

#### `Init(mail *Mail, web *WebHook)` 函数  
- 作用: 初始化整个 notify 包
- 输入:
    1. `mail`: 包含 SMTP 配置的 Mail 结构体指针
    2. `web`: 包含 WebHook 配置的 WebHook 结构体指针
- 流程:
    1. 根据传入的 mail 参数, 创建一个新的 Mail 实例并赋值给包内私有变量 _defaultMail
    2. 根据传入的 web 参数, 创建一个新的 WebHook 实例并赋值给包内私有变量 _defaultWebHook
    3. 使用 make(chan *Message, 64) 创建一个容量为 64 的带缓冲 channel, 并赋值给包内私有变量 msgQueue

#### `Send(msg *Message)` 函数
- 作用:  作为“生产者”, 为系统其他部分提供一个发送通知的入口. 这个函数是异步的, 它将消息放入队列后会立即返回，不会等待通知被实际发送
- 输入: 
    1. `msg`: 一个包含了完整通知消息的对象指针
- 流程: 
    1. 将传入的 msg 指针发送到 msgQueue channel 中

#### `Serve()` 函数
- 作用: 作为“消费者”, 在一个独立的 goroutine 中运行, 持续地从消息队列中取出任务并分发给对应的处理器
- 流程: 
    1. 使用 for msg := range msgQueue 循环, 持续地,阻塞地从msgQueue channel中接收消息
    2. 对接收到的 msg 进行 nil 检查
    3. 调用 msg.Check() 方法对消息进行预处理
    4. 使用 switch msg.Type 语句判断消息类型
    5. case NotifyTypeMail: 同步调用 _defaultMail.SendMsg(msg) 发送邮件
    6. case NotifyTypeWebHook: 启动一个新的goroutine (go _defaultWebHook.SendMsg(msg)) 来发送 WebHook, 这可以防止慢速 HTTP 请求阻塞整个消费者循环

#### `(m *Message) Check` 方法
- 作用: 对消息进行发送前的标准化处理和数据清洗
- 输入: 
    1. `m`: Message 对象的接收者指针
- 流程: 
    1. 检查 m.OccurTime 字段是否为空. 如果为空, 则用当前时间(格式化为秒)填充它
    2. 使用 strings.ReplaceAll 将 m.Body 中的所有双引号 " 替换为 单引号 '
    3. 使用 strings.ReplaceAll 将 m.Body 中的所有换行符 \n 删除

#### `(mail *Mail) SendMsg(msg *Message)` 方法
- 作用: Mail 结构体存储 SMTP 配置, SendMsg 方法实现了发送邮件的具体逻辑, 它封装了使用 gomail 库的复杂性
- 输入: 
    1. `msg`: 要发送的消息对象
- 流程: 
    1. 创建一个 gomail.NewMessage() 对象
    2. 设置邮件的 From, To, Subject 等头部信息
    3. 调用 parseMailTemplate(msg) 将 HTML 模板和消息数据结合, 生成最终的邮件正文
    4. 将生成的 HTML 设置为邮件的 Body
    5. 创建一个 gomail.NewDialer(), 配置好 SMTP 服务器地址,端口,和认证信息
    6. 调用 d.DialAndSend(m) 连接服务器并发送邮件
    7. 如果发送失败, 记录一条警告日志

#### `parseMailTemplate(msg *Message)` 函数
- 作用: 一个辅助函数, 专门负责解析 HTML 邮件模板并用真实数据填充
- 输入: 
    1. `msg`: 包含模板所需数据 (如 Subject, IP, Body 等) 的消息对象
- 流程: 
    1. 使用 text/template 包解析预定义的 mailTemplate 字符串
    2. 创建一个 bytes.Buffer 作为写入目标
    3. 调用 tmpl.Execute(), 将 msg 的数据填充到模板中, 结果写入 Buffer
    4. 处理可能的错误
- 输出:
    1. `string`:  填充数据后生成的最终 HTML 字符串

#### `(w *WebHook) SendMsg(msg *Message)` 方法
- 作用: WebHook 结构体存储 Webhook 配置, SendMsg 方法实现了发送 Webhook 通知的具体逻辑. 它采用了策略模式, 能根据 Kind 处理不同类型的 Webhook
- 输入: 
    1. `msg`: 要发送的消息对象
- 流程: 
    1. 使用 switch _defaultWebHook.Kind 对 Webhook 类型进行判断
    2. case "feishu":
        - 使用 strings.Replace 将模板中的 timeSlot, ipSlot, msgSlot, subjectSlot 等占位符替换成 msg 中的实际数据
        - 特殊处理 userSlot: 遍历 msg.To 列表, 为每个用户生成飞书的 \<at\> 标签并拼接成字符串, 然后替换到模板中
        - 调用 httpclient.PostJson 将最终的卡片 JSON 发送到飞书的 WebHook URL
        - 如果失败, 记录错误日志
    3. default:
        - 使用 json.Marshal(msg) 将整个 Message 对象序列化为通用的 JSON 字符串
        - 调用 httpClient.PostJson 将这个 JSON 字符串发送到指定的 Webhook URL
        - 如果失败, 记录错误日志
- 输出:
    1. `string`:  填充数据后生成的最终 HTML 字符串
