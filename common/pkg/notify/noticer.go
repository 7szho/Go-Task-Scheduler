package notify

import (
	"crony/common/pkg/utils"
	"strings"
	"time"
)

// 定义了所有通知方式需要实现的方法
type Noticer interface {
	SendMsg(*Message)
}

// 定义了一个通知消息所包含的所有数据
type Message struct {
	Type      int      // 消息类型, 区分 Mail 还是 WebHook
	IP        string   // 相关的 IP 地址
	Subject   string   // 消息主题
	Body      string   // 消息正文
	To        []string // 收件人列表
	OccurTime string   // 事件发生时间
}

// msgQueue 是一个包级别的 channel, 作为消息队列使用
var msgQueue chan *Message

// Init 是通知包的初始化函数
func Init(mail *Mail, web *WebHook) {
	// 初始化默认的 Mail 设置
	_defaultMail = &Mail{
		Port:     mail.Port,
		From:     mail.From,
		Host:     mail.Host,
		Secret:   mail.Secret,
		Nickname: mail.Nickname,
	}
	// 初始化默认的 WebHook 设置
	_defaultWebHook = &WebHook{
		Kind: web.Kind,
		Url:  web.Url,
	}
	// 创建消息队列 channel, 并设置缓冲区大小为 64
	msgQueue = make(chan *Message, 64)
}

// Send 是一个暴露给外部调用的函数, 用于发送一条通知
func Send(msg *Message) {
	// 将信息放入消息队列, 然后立即返回, 不会等待消息被真正发送
	msgQueue <- msg
}

// 这个函数会阻塞并持续地从消息队列中读取消息, 然后分发给相应的处理器
func Serve() {
	// 使用 for range 循环遍历 channel
	for msg := range msgQueue {
		if msg == nil {
			continue
		}

		// 使用 switch 语句根据消息类型进行分支
		switch msg.Type {
		case NotifyTypeMail:
			// Mail:
			// 1. 调用 Check 方法对消息进行预处理和格式化
			msg.Check()
			// 2. 调用邮件发送器发送消息
			// 这是一个同步调用, Serve 循环会等待发送完成
			_defaultMail.SendMsg(msg)
		case NotifyTypeWebHook:
			// webhook:
			// 1. 调用 Check 方法对消息进行预处理和格式化
			msg.Check()
			// 2. 使用 go 关键字启动一个新的 goroutine 来发送消息
			// 这是一个异步调用, 意味着 Serve 循环不会等待 WebHook 发送完
			// 而是可以立即处理队列中的下一条消息
			go _defaultWebHook.SendMsg(msg)
		}
	}
}

// Check 是 Message 类型的一个方法, 用于发送前对消息数据进行检查和标准化
func (m *Message) Check() {
	// 如果消息中没有指定发送时间, 则自动设置为当前时间
	if m.OccurTime == "" {
		m.OccurTime = time.Now().Format(utils.TimeFormatSecond)
	}
	// Remove the transfer character
	// 对消息正文进行清理, 移除可能导致问题的特殊字符
	m.Body = strings.ReplaceAll(m.Body, "\"", "'")
	m.Body = strings.ReplaceAll(m.Body, "\n", "")
}
