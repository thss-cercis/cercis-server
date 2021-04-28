package redis

import "time"

// TagSMSSignUp sms 用户注册服务的 tag
const TagSMSSignUp = "SMS_SignUp"

// TagSMSSignUpRetry sms 用户注册冷却期的 tag
const TagSMSSignUpRetry = "SMS_SignUp_Retry"

// TagSMSRecover sms 密码找回服务的 tag
const TagSMSRecover = "SMS_Recover"

// TagSMSRecoverRetry sms 密码找回冷却期的 tag
const TagSMSRecoverRetry = "SMS_Recover_Retry"

// ExpSMSSignUp sms 用户注册服务的键值对有效期
const ExpSMSSignUp = 10 * time.Minute

// ExpSMSSignUpRetry sms 用户注册冷却期的键值对有效期
const ExpSMSSignUpRetry = 58 * time.Second

// ExpSMSRecover sms 密码找回服务的键值对有效期
const ExpSMSRecover = 10 * time.Minute

// ExpSMSRecoverRetry sms 密码找回冷却期的键值对有效期
const ExpSMSRecoverRetry = 58 * time.Second
