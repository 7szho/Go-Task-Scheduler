package utils

import "errors"

type fmsState int

const (
	stateArgumentOutside fmsState = iota // 在参数
	stateArgumentStart
	stateArgumentEnd
)

var errEndOfLine = errors.New("end of line")

type cmdArgumentParser struct {
	s            string   // 要解析的输入字符串
	i            int      // 当前解析的字符索引
	length       int      // 输入字符串的总长度
	state        fmsState // 当前有限状态机的状态
	startToken   byte     // 如果参数被引号包裹, 记录起始印好; 如果是非引号参数, 则为0
	shouldEscape bool     // 标志位: 指示下一个字符是否应该被转义处理
	currArgument []byte   // 当前正在构建的参数内容
	err          error    // 记录解析过程中遇到的错误
}

// 创建并初始化一个 cmdArgumentParser 实例
func newCmdArgumentParser(s string) *cmdArgumentParser {
	return &cmdArgumentParser{
		s:            s,                   // 待解析的字符串
		i:            -1,                  // 初始化索引为 -1, 以便在第一次调用 next() 时指向第一个字符
		length:       len(s),              // 字符串的总长度
		currArgument: make([]byte, 0, 16), // 初始化一个容量为16的字节切片, 用于存储当前参数参数
	}
}

// 命令行参数解析的主循环
func (cap *cmdArgumentParser) parse() (arguments []string) {
	for {
		cap.next()

		if cap.err != nil {
			if cap.shouldEscape {
				cap.currArgument = append(cap.currArgument, '\\')
			}

			if len(cap.currArgument) > 0 {
				arguments = append(arguments, string(cap.currArgument))
			}

			return
		}

		switch cap.state {
		case stateArgumentOutside:
			cap.detectStartToken()
		case stateArgumentStart:
			if !cap.detectEnd() {
				cap.detectContent()
			}
		case stateArgumentEnd:
			cap.state = stateArgumentOutside
			arguments = append(arguments, string(cap.currArgument))
			cap.currArgument = cap.currArgument[:0]
		}
	}
}

// 递减 i, 实现字符串字符的遍历
func (cap *cmdArgumentParser) previous() {
	if cap.i >= 0 {
		cap.i--
	}
}

// 递增 i, 实现字符串字符的遍历
func (cap *cmdArgumentParser) next() {
	if cap.length-cap.i == 1 {
		cap.err = errEndOfLine
		return
	}
	cap.i++
}

// 识别新参数的起始
func (cap *cmdArgumentParser) detectStartToken() {
	c := cap.s[cap.i]
	if c == ' ' {
		return
	}

	switch c {
	case '\\':
		cap.startToken = 0
		cap.shouldEscape = true
	case '"', '\'':
		cap.startToken = c
	default:
		cap.startToken = 0
		cap.previous()
	}
	cap.state = stateArgumentStart
}

// 处理参数内部的字符
func (cap *cmdArgumentParser) detectContent() {
	c := cap.s[cap.i]

	if cap.shouldEscape {
		switch c {
		case ' ', '\\', cap.startToken:
			cap.currArgument = append(cap.currArgument, c)
		default:
			cap.currArgument = append(cap.currArgument, '\\', c)
		}
		cap.shouldEscape = false
		return
	}

	if c == '\\' {
		cap.shouldEscape = true
	} else {
		cap.currArgument = append(cap.currArgument, c)
	}
}

// 检查当前参数是否结束
func (cap *cmdArgumentParser) detectEnd() (detected bool) {
	c := cap.s[cap.i]

	if cap.startToken == 0 {
		if c == ' ' && !cap.shouldEscape {
			cap.state = stateArgumentEnd
			cap.previous()
			return true
		}
		return false
	}

	if c == cap.startToken && !cap.shouldEscape {
		cap.state = stateArgumentEnd
		return true
	}

	return false
}

// getter 函数
func ParseCmdArguments(s string) (arguments []string) {
	return newCmdArgumentParser(s).parse()
}
