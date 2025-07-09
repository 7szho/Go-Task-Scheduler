package notify

import (
	"crony/common/pkg/httpclient"
	"crony/common/pkg/logger"
	"encoding/json"
	"fmt"
	"strings"
)

// 定义了一个 WebHook 通知所需的配置信息
type WebHook struct {
	Kind string // WebHook 的类型
	Url  string // WebHook 的接收地址
}

// 一个包级别的私有变量, 用于存储默认的 WebHook 配置
var _defaultWebHook *WebHook

// SendMsg 是 WebHook 类型的一个方法, 实现了 Noticer 接口
func (w *WebHook) SendMsg(msg *Message) {
	switch _defaultWebHook.Kind {
	case "feishu":
		// 1. 加载预定义的飞书卡片消息模板
		var sendData = feiShuTemplateCard
		// 2. 逐个替换模板中的展位符
		sendData = strings.Replace(sendData, "timeSlot", msg.OccurTime, 1)
		sendData = strings.Replace(sendData, "ipSlot", msg.IP, 1)

		// 3. 投诉处理 @-mention 用户
		// 飞书的卡片消息中, @某人需要特定的 <at> 标签
		userSlot := ""
		// 遍历所有收件人
		for _, to := range msg.To {
			// 将每个收件人构成一个 <at> 标签并拼接到字符串中
			userSlot += fmt.Sprintf("<at email='' >%s</at>", to)
		}
		// 将拼接好的 @-mention 字符替换到模板中
		sendData = strings.Replace(sendData, "userSlot", userSlot, 1)
		sendData = strings.Replace(sendData, "msgSlot", msg.Body, 1)
		sendData = strings.Replace(sendData, "subjectSlot", msg.Subject, 1)
		// 4. 调用 httpclient 包, 将最终生成的 JSON 字符串作为请求体发送到飞书的 WebHook URL
		_, err := httpclient.PostJson(_defaultWebHook.Url, sendData, 0)
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("feishu send msg[%+v] err: %s", msg, err.Error()))
		}
	default:
		// 1. 将整个 Message 结构体序列化为 JSON
		b, err := json.Marshal(msg)
		if err != nil {
			return
		}
		// 2. 发送 HTTP POST 请求
		// 将序列化后的 JSON 字符串发送到配置的 WebHook URL
		_, err = httpclient.PostJson(_defaultWebHook.Url, string(b), 0)
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("web hook api send msg[%+v] err: %s", msg, err.Error()))
		}
	}
}
