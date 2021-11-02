package alert

import (
	"crypto/tls"
	"fmt"
	"github.com/itnotebooks/fireye/config"
	"log"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// RenderMailTemp 渲染邮件模板
func RenderMailTemp(app config.LogDirField, body string) string {
	global_config := config.GLOBAL_CONFIG

	privateIP := global_config.PIP
	path := app.Path

	if global_config.Platform == "container" {
		privateIP = app.PodName
	}

	body = fmt.Sprintf("--------ERROR Message--------\n"+
		"Project: %v\n"+
		"Mode: %v\n"+
		"Global IP: %v\n"+
		"Private IP: %v\n"+
		"Log File: %v\n"+
		"---------------------------\n\n"+
		"%v\n", global_config.Project, app.Name,
		global_config.GIP, privateIP, path, body)
	return body
}

// 发送错误日志邮件
func send_mail(subject, body string) (string, error) {
	// 判断是否有开启送信
	if !config.GLOBAL_CONFIG.SMTP.SMTP_ENABLE {
		fmt.Println("未启用SMTP送信")
		return "未启用SMTP送信", nil
	}
	// 获取SMTP配置
	config := config.GLOBAL_CONFIG
	smtp_config := config.SMTP

	// 认证
	auth := smtp.PlainAuth("",
		smtp_config.SMTP_USERNAME,
		smtp_config.SMTP_PASSWORD,
		smtp_config.SMTP_SERVER)

	// 渲染邮件体
	header := make(map[string]string)
	header["From"] = smtp_config.SMTP_USERNAME
	header["To"] = strings.Join(config.MailTo, ";")
	header["Cc"] = strings.Join(config.MailCC, ";")
	header["Subject"] = subject
	header["Content-Type"] = "text/plain; charset=UTF-8"

	msg := ""
	for k, v := range header {
		msg += fmt.Sprintf("%s:%s\r\n", k, v)
	}

	msg += "\r\n" + body

	// 拼接SMTP地址
	smtpServer := fmt.Sprintf("%v:%v", smtp_config.SMTP_SERVER, smtp_config.SMTP_PORT)

	// 送信
	err := SendMailUsingTLS(
		smtpServer,
		auth,
		smtp_config.SMTP_USERNAME,
		config.MailTo,
		[]byte(msg))

	if err != nil {
		fmt.Printf("smtp error: %s\n", err)
		return "failed", err
	} else {
		return "success", err
	}

}

func SendMail(app config.LogDirField, body string) {
	log.Println("========>>>>>>>>")
	global_config := config.GLOBAL_CONFIG
	// 渲染邮件模板
	body = RenderMailTemp(app, body)

	// 发送邮件
	subject := fmt.Sprintf("ERROR: %v webapps:%v (%v)", global_config.Project, app.Name, time.Now().Format("2006-01-02 15:04:05"))
	ret, err := send_mail(subject, body)
	fmt.Printf("%v: Email Send %v..., msg: %v\n", app.Name, ret, err)
	config.WG.Done()
}

func SendMailUsingTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) (err error) {
	c, err := Dial(addr)
	if err != nil {
		return err
	}
	defer c.Close()
	if auth != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err = c.Auth(auth); err != nil {
				return err
			}
		}
	}
	if err = c.Mail(from); err != nil {
		return err
	}

	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			fmt.Print(err)
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}

func Dial(addr string) (*smtp.Client, error) {
	conn, err := tls.Dial("tcp", addr, nil)
	if err != nil {
		return nil, err
	}

	// 分解主机端口字符串
	host, _, _ := net.SplitHostPort(addr)
	return smtp.NewClient(conn, host)
}
