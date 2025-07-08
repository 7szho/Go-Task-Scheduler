package notify

import (
	"bytes"
	"crony/common/pkg/logger"
	"fmt"
	"text/template"

	"github.com/go-gomail/gomail"
)

// 定义通知类型的常量
const (
	NotifyTypeMail    = 1
	NotifyTypeWebHook = 2
)

var mailTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
    <title></title>
    <meta charset="utf-8"/>

</head>
<body>
<div class="cap" style="
            border: 2px solid black;
            background-color: whitesmoke;
            height: 500px"
>
    <div class="content" style="
            background-color: white;
            background-clip: padding-box;
            color:black;
            font-size: 13px;
            padding: 25px 25px;
            margin: 25px 25px;
        ">
        <div class="hello" style="text-align: center; color: #FF3333;font-size: 18px;font-weight: bolder">
            {{.Subject}}
        </div>
        <br>
        <div>
            <table border="1"  bordercolor="black" cellspacing="0px" cellpadding="4px" style="margin: 0 auto;">
                <tr >
                    <td>告警主机</td>
                    <td >{{.IP}}</td>
                </tr>

                <tr>
                    <td>告警时间</td>
                    <td>{{.OccurTime}}</td>
                </tr>

                <tr>
                    <td>告警信息</td>
                    <td>{{.Body}}</td>
                </tr>

            </table>
        </div>
        <br><br>
    </div>
</div>
<br>

</body>
</html>
`
var _defaultMail *Mail

type Mail struct {
	Port     int
	From     string
	Host     string
	Secret   string
	Nickname string
}

// SendMsg 方法用于发送一封邮件
// msg: 一个指向 Message 结构体的指针
func (mail *Mail) SendMsg(msg *Message) {
	// 创建一个新的 gomail 消息对象
	m := gomail.NewMessage()

	// 设置邮件头 "From"(发件人), 并使用 FormatAddress 添加昵称
	m.SetHeader("From", m.FormatAddress(_defaultMail.From, _defaultMail.Nickname)) //这种方式可以添加别名，即“XX官方”
	// 设置邮件头 "To"(收件人)
	m.SetHeader("To", msg.To...)
	// 设置邮件主题
	m.SetHeader("Subject", msg.Subject)
	// 调用 parseMailTemplate 函数, 将模板和数据结合, 生成 HTML 邮件正文
	msgData := parseMailTemplate(msg)

	// 设置邮件正文, 并指定内容类型为 "text/html"
	m.SetBody("text/html", msgData)

	// 创建一个 Dialer 对象
	d := gomail.NewDialer(_defaultMail.Host, _defaultMail.Port, _defaultMail.From, _defaultMail.Secret)
	// 连接到 SMTP 服务器并发送邮件
	if err := d.DialAndSend(m); err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("smtp send msg[%+v] err: %s", msg, err.Error()))
	}
}

// parseMailTemplate 函数封装解析邮件模板并填充邮件
// msg: 包含模板所需数据的 Message 对象
// 返回值: 填充数据后生成的最终 HTML 字符串
func parseMailTemplate(msg *Message) string {
	tmpl, err := template.New("notify").Parse(mailTemplate)
	if err != nil {
		return fmt.Sprintf("Failed to parse the notification template error: %s", err.Error())
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, msg)
	if err != nil {
		return fmt.Sprintf("Failed to parse the notification template execute error: %s", err.Error())
	}
	return buf.String()
}
