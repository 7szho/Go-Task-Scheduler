package notify

type WebHook struct {
	Kind string
	Url  string
}

var _defaultWebHook *WebHook

func (w *WebHook) SendMsg(msg *Message) {
	switch _defaultWebHook.Kind {
	}
}
