package handler

import "crony/common/models"

// Handler 是一个接口，定义了所有具体的任务处理器需要实现的方法
type Handler interface {
	// Run方法接收一个Job指针，执行任务，并返回执行结果和错误
	Run(job *Job) (string, error)
}

// CreateHandler 函数用于创建一个Handler实例
// 接收一个Job对象，并根据Job的类型返回一个实现Handler接口的处理器
func CreateHandler(j *Job) Handler {
	var handler Handler = nil // 初始化handler变量为nil
	// 使用switch豫剧检查作业的类型
	switch j.Type {
	// 如果是命令类型（JobTypeCmd）的作业
	case models.JobTypeCmd:
		// 创建一个新的CMDHanlder实例
		handler = new(CMDHandler)
	// 如果是HTTP请求类型（JobTypeHttp）的作业
	case models.JobTypeHttp:
		// 创建一个新的HTTPHanlder实例
		handler = new(HTTPHandler)
	}
	// 返回创建好的具体处理器实例
	return handler

}
