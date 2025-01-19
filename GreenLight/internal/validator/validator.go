package validator

import "regexp"

// TODO 封装各种的校验器（使用正则表达式）
var (
	// EmailRX 用于验证电子邮件地址格式的正则表达式。
	// 正则表达式模式来源于 https://html.spec.whatwg.org/#valid-e-mail-address。
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

// Validator 结构体包含一个验证错误映射。
type Validator struct {
	Errors map[string]string
}

// New 创建一个新的 Validator 实例，并初始化一个空的错误映射。
//
// 返回值:
// - *Validator: 指向新创建的 Validator 实例的指针。
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// Valid 如果错误映射中没有任何条目，则返回 true。
//
// 返回值:
// - bool: 如果没有错误条目，返回 true；否则返回 false。
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// AddError 向映射中添加错误消息（前提是给定键不存在）。
//
// 参数:
// - key: 错误键。
// - message: 错误消息。
func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// Check 如果验证检查不通过，则向映射中添加错误消息。
//
// 参数:
// - ok: 验证是否通过的布尔值。
// - key: 错误键。
// - message: 错误消息。
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// In 如果特定值存在于字符串列表中，则返回 true。
//
// 参数:
// - value: 要检查的值。
// - list: 字符串列表。
//
// 返回值:
// - bool: 如果值存在于列表中，返回 true；否则返回 false。
func In(value string, list ...string) bool {
	for i := range list {
		if value == list[i] {
			return true
		}
	}
	return false
}

// Matches 如果字符串值匹配特定的正则表达式模式，则返回 true。
//
// 参数:
// - value: 要匹配的字符串值。
// - rx: 正则表达式对象。
//
// 返回值:
// - bool: 如果字符串值匹配正则表达式模式，返回 true；否则返回 false。
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// Unique 如果切片中的所有字符串值都是唯一的，则返回 true。
//
// 参数:
// - values: 字符串切片。
//
// 返回值:
// - bool: 如果所有字符串值都是唯一的，返回 true；否则返回 false。
func Unique(values []string) bool {
	uniqueValues := make(map[string]bool)

	for _, value := range values {
		uniqueValues[value] = true
	}

	return len(values) == len(uniqueValues)
}
