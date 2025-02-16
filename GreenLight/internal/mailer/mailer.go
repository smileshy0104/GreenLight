package mailer

import (
	"bytes"
	"embed"
	"github.com/go-mail/mail/v2"
	"html/template"
	"time"
)

// 嵌入静态文件
//
//go:embed "templates"
var templateFS embed.FS

// 声明Mailer结构体
type Mailer struct {
	dialer *mail.Dialer
	sender string
}

// New 创建并初始化一个新的Mailer实例。
// 这个函数需要主机地址、端口号、用户名、密码和发件人地址作为参数。
// 主机地址(host)是邮件服务器的地址。
// 端口号(port)是邮件服务器使用的端口。
// 用户名(username)和密码(password)用于邮件服务器的身份验证。
// 发件人地址(sender)是邮件的发送方地址。
// 函数返回一个新的Mailer实例。
func New(host string, port int, username, password, sender string) Mailer {
	// 初始化Dialer
	dialer := mail.NewDialer(host, port, username, password)
	// 设置Dialer的超时时间为5秒，以防止连接耗时过长
	dialer.Timeout = 5 * time.Second

	// 返回Mailer实例，包含初始化的Dialer和发件人地址
	return Mailer{
		dialer: dialer,
		sender: sender,
	}
}

// Send 方法用于发送邮件。
// 它根据提供的模板文件和数据生成邮件内容，并将其发送给指定的收件人。
// 参数:
//
//	recipient: 收件人的电子邮件地址。
//	templateFile: 用于生成邮件内容的模板文件路径。
//	data: 用于填充模板的数据。
//
// 返回值:
//
//	如果邮件发送成功，则返回nil；否则返回错误。
func (m Mailer) Send(recipient, templateFile string, data interface{}) error {
	// 创建一个mail.NewMessage()实例
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	// 生成邮件主题
	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	// 生成纯文本邮件正文
	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	// 生成HTML邮件正文
	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

	// 创建邮件消息并设置相关头信息
	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())

	// 尝试发送邮件，最多重试3次
	for i := 1; i <= 3; i++ {
		err = m.dialer.DialAndSend(msg)
		if nil == err {
			return nil
		}

		// 如果发送失败，等待500毫秒后重试
		time.Sleep(500 * time.Millisecond)
	}

	// 如果经过重试后仍然发送失败，则返回错误
	return err
}
