package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// TODO 自定义类型Runtime，同时是实现了自定义MarshalJSON和UnmarshalJSON
// ErrInvalidRuntimeFormat 在无法解析或转换 JSON 字符串时返回错误。
// 该错误在 Runtime.UnmarshalJSON() 方法中使用。
var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

// Runtime 表示电影的时长，使用 int32 类型。
// 自定义类型
type Runtime int32

// MarshalJSON 实现了 json.Marshaler 接口 interface，用于将 Runtime 类型编码为 JSON。
// 它返回一个格式为 "<runtime> mins" 的 JSON 编码字符串。
//
// 返回值:
// - []byte: JSON 编码后的字节切片
// - error: 如果发生错误则返回非空错误
// r Runtime 使用值接收器是因为：值方法可以在指针和值上调用，但指针方法只能在指针上调用！这极大增加了对应的灵活性
func (r Runtime) MarshalJSON() ([]byte, error) {
	// 生成符合要求格式的字符串
	jsonValue := fmt.Sprintf("%d mins", r)

	// 使用 strconv.Quote() 函数将字符串包裹在双引号中，使其成为————有效的 JSON 字符串
	// 否则，它将不会被解释为一个JSON字符串
	quotedJSONValue := strconv.Quote(jsonValue)

	// 将带引号的字符串转换为字节切片并返回
	return []byte(quotedJSONValue), nil
}

// UnmarshalJSON 实现了 json.Unmarshaler 接口，用于将 JSON 解码为 Runtime 类型。
// 注意：因为 UnmarshalJSON() 需要修改接收者（Runtime 类型），所以必须使用指针接收者。
// 否则，我们只会修改一个副本（当方法返回时该副本会被丢弃）。
//
// 参数:
// - jsonValue: 包含 JSON 数据的字节切片
//
// 返回值:
// - error: 如果发生错误则返回非空错误
func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
	// 去除字符串周围的双引号。如果无法取消引用，则返回 ErrInvalidRuntimeFormat 错误
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// 拆分字符串以隔离数字部分
	parts := strings.Split(unquotedJSONValue, " ")

	// 检查字符串是否符合预期格式。如果不符，则返回 ErrInvalidRuntimeFormat 错误
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}

	// 将包含数字的字符串解析为 int32。如果解析失败，则返回 ErrInvalidRuntimeFormat 错误
	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// 将 int32 转换为 Runtime 类型并赋值给接收者
	*r = Runtime(i)

	return nil
}
