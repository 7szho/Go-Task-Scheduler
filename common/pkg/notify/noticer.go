package notify

type Noticer interface {
	SendMsg(*Message)
}

type Message struct {
	Type      int
	IP        string
	Subject   string
	Body      string
	To        []string
	OccurTime string
}

var msgQueue chan *Message

func Init(mail *Mail, web *WebHook) {

}

func Send(msg *Message) {
	msgQueue <- msg
}

func Serve() {

}
