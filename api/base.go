package api

// BaseRes 所有回复的基类
type BaseRes struct {
	Code    int64       `json:"code"`
	Msg     string      `json:"msg"`
	Payload interface{} `json:"payload"`
}
