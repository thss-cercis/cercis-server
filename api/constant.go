package api

// MsgUnknown 成功信息
const MsgUnknown = "未知错误"

// MsgSuccess 成功信息
const MsgSuccess = "ok"

// MsgWrongParam 成功信息
const MsgWrongParam = "参数错误"

// MsgNotLogin 未登录的信息
const MsgNotLogin = "未登录"

// MsgUserNotFound 用户无法找到的信息
const MsgUserNotFound = "未找到指定用户"

// MsgSMSError sms 服务异常的信息
const MsgSMSError = "SMS 服务异常"

// MsgSMSTooOften sms 服务异常的信息
const MsgSMSTooOften = "SMS 服务调用频率过快"

// MsgSMSWrong SMS 验证码错误
const MsgSMSWrong = "验证码错误"

// MsgUserAlreadyExist 用户已经存在
const MsgUserAlreadyExist = "用户已经存在"

// MsgChatError 聊天服务异常
const MsgChatError = "聊天服务异常"

// MsgChatCreateFail 创建聊天失败
const MsgChatCreateFail = "创建聊天失败"

// MsgChatDeleteFail 创建聊天失败
const MsgChatDeleteFail = "删除聊天失败"

// MsgChatMemberAddFail 添加群聊成员失败
const MsgChatMemberAddFail = "添加聊天成员失败"

// CodeFailure 未知错误
const CodeFailure = -1

// CodeSuccess 成功
const CodeSuccess = 0

// CodeBadParam 无效参数或缺少参数
const CodeBadParam = 1

// CodeNotLogin 未登录
const CodeNotLogin = 100

// CodeUserBadPassword 密码错误
const CodeUserBadPassword = 101

// CodeUserAlreadyLogin 已经登陆
const CodeUserAlreadyLogin = 102

// CodeUserIDNotFound 找不到用户 id
const CodeUserIDNotFound = 103

// CodeUserAlreadyExist 用户已经存在
const CodeUserAlreadyExist = 104

// CodeSMSError SMS 服务异常
const CodeSMSError = 200

// CodeSMSTooOften SMS 服务过快
const CodeSMSTooOften = 201

// CodeSMSWrong SMS 验证码错误
const CodeSMSWrong = 202

// CodeChatError 聊天服务异常
const CodeChatError = 300

// CodeChatCreateFail 创建聊天失败
const CodeChatCreateFail = 301

// CodeChatDeleteFail 删除聊天失败
const CodeChatDeleteFail = 302

// CodeChatMemberAddFail 添加群聊成员失败
const CodeChatMemberAddFail = 303

/*
 * WebSocket Type code
 */

// TypePong 新好友请求
const TypePong = 2

// TypeNewFriendApply 新好友请求
const TypeNewFriendApply = 100

// TypeFriendListUpdate 好友列表更新
const TypeFriendListUpdate = 101

// TypeAddNewMessage 新消息
const TypeAddNewMessage = 200

// TypeWithdrawMessage 撤回消息
const TypeWithdrawMessage = 200
