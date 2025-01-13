package __Unintended_variable_shadowing

import (
	"log"
	"net/http"
	"testing"
)

// TODO: 01、意外的变量隐藏

// Test_01 测试在启用或禁用跟踪模式下创建客户端的功能。
// 该测试展示了如何使用闭包和变量遮盖来根据跟踪标志创建不同类型的 HTTP 客户端。
func Test_01(t *testing.T) {
	// 启用或禁用跟踪模式。
	tracing := true

	// 定义一个闭包，用于创建和测试客户端。
	// 注意：闭包中包含一个被遮盖的变量 'client'，该变量在 if 和 else 块中重新声明。
	func() error {
		// 被遮盖的变量————默认情况下，变量 'client' 是未初始化的。
		var client *http.Client

		// 如果启用了跟踪模式，创建带有跟踪的客户端。
		if tracing {
			// 创建带有跟踪的客户端。
			client, err := createClientWithTracing()
			if err != nil {
				return err
			}
			log.Println(client)
		} else {
			// 创建默认的客户端，不带跟踪。
			client, err := createDefaultClient()
			if err != nil {
				return err
			}
			log.Println(client)
		}
		return nil
	}()
}

func Test_02(t *testing.T) {
	// 启用或禁用跟踪模式。
	tracing := true

	func() error {
		var client *http.Client

		if tracing {
			// 利用中间变量
			c, err := createClientWithTracing()
			if err != nil {
				return err
			}
			// 进行赋值
			client = c
			log.Println(client)
		} else {
			c, err := createDefaultClient()
			if err != nil {
				return err
			}
			client = c
			log.Println(client)
		}
		return nil
	}()
}

func Test_03(t *testing.T) {
	// 启用或禁用跟踪模式。
	tracing := true

	func() error {
		// 提前声明对应变量
		var client *http.Client
		var err error

		if tracing {
			// 直接给变量赋值
			client, err = createClientWithTracing()
			log.Println(client)
		} else {
			client, err = createDefaultClient()
			log.Println(client)
		}
		if err != nil {
			return err
		}
		return nil
	}()
}

func createClientWithTracing() (*http.Client, error) {
	return nil, nil
}
func createDefaultClient() (*http.Client, error) {
	return nil, nil
}
