package email

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"net/mail"
	"net/smtp"
	"strings"
	"sync"
)

type NotAuth struct {
	Host     string
	Username string
	Password string
}

func (a *NotAuth) Start(server *smtp.ServerInfo) (proto string, toServer []byte, err error) {
	if !server.TLS {
		advertised := false
		for _, mechanism := range server.Auth {
			if mechanism == "LOGIN" {
				advertised = true
				break
			}
		}
		if !advertised {
			return "", nil, errors.New("gomail: unencrypted connection")
		}
	}
	//if server.Name != a.Host {
	//    return "", nil, errors.New("gomail: wrong Host name")
	//}
	return "LOGIN", nil, nil
}
func (a *NotAuth) Next(fromServer []byte, more bool) (toServer []byte, err error) {
	if !more {
		return nil, nil
	}

	switch {
	case bytes.Equal(fromServer, []byte("Username:")):
		return []byte(a.Username), nil
	case bytes.Equal(fromServer, []byte("Password:")):
		return []byte(a.Password), nil
	default:
		return nil, fmt.Errorf("gomail: unexpected server challenge: %s", fromServer)
	}
}

type ConfigMapper struct {
	TLS           bool
	Host          string
	Port          int
	Username      string
	Password      string
	SkipTLSVerify bool // 是否跳过TLS证书验证，默认false
}
type Email struct {
	mapper map[string]*ConfigMapper
}

// validateConfig 验证配置的有效性
func validateConfig(mapper map[string]*ConfigMapper) error {
	if len(mapper) == 0 {
		return errors.New("empty configuration mapper")
	}

	// 检查是否有默认配置
	if _, hasDefault := mapper["default"]; !hasDefault {
		// 如果没有默认配置，确保所有配置都有效
		for domain, config := range mapper {
			if err := validateSingleConfig(config); err != nil {
				return fmt.Errorf("invalid configuration for domain %s: %v", domain, err)
			}
		}
	} else {
		// 如果有默认配置，确保默认配置有效
		if err := validateSingleConfig(mapper["default"]); err != nil {
			return fmt.Errorf("invalid default configuration: %v", err)
		}

		// 其他配置可以部分有效（会回退到默认配置）
		for domain, config := range mapper {
			if domain != "default" && config != nil {
				_ = validateSingleConfig(config) // 非默认配置的验证失败不影响整体
			}
		}
	}

	return nil
}

// validateSingleConfig 验证单个配置项的有效性
func validateSingleConfig(config *ConfigMapper) error {
	if config == nil {
		return errors.New("nil configuration")
	}

	if config.Host == "" {
		return errors.New("empty host")
	}

	if config.Port <= 0 || config.Port > 65535 {
		return errors.New("invalid port")
	}

	if config.Username == "" {
		return errors.New("empty username")
	}

	if config.Password == "" {
		return errors.New("empty password")
	}

	return nil
}

// New 创建一个新的Email实例
func New(mapper map[string]*ConfigMapper) *Email {
	// 验证配置
	if err := validateConfig(mapper); err != nil {
		// 配置验证失败时记录警告，但仍然创建实例（允许后续修复配置）
		fmt.Printf("Warning: %v\n", err)
	}

	return &Email{
		mapper: mapper,
	}
}

// GetMapper 根据邮箱地址获取对应的配置
func (m *Email) GetMapper(email string) (*ConfigMapper, bool) {
	// 解析邮箱地址，提取域名
	addr, err := mail.ParseAddress(email)
	var domain = "default"
	if err == nil {
		// 从解析后的邮箱地址中提取域名
		if idx := strings.LastIndex(addr.Address, "@"); idx != -1 {
			domain = addr.Address[idx+1:]
		}
	} else {
		// 解析失败时回退到简单分割
		domainSplit := strings.Split(email, "@")
		if len(domainSplit) >= 2 {
			domain = domainSplit[len(domainSplit)-1]
		}
	}

	// 首先查找域名对应的配置
	if mapper, ok := m.mapper[domain]; ok {
		return mapper, true
	}

	// 如果域名没有配置，使用默认配置
	if mapper, ok := m.mapper["default"]; ok {
		return mapper, true
	}

	return nil, false
}

// buildMessage 构建邮件消息
func buildMessage(from, to mail.Address, subject, contentType, content string) []byte {
	return []byte(fmt.Sprintf("To: %s\r\nFrom: %s\r\nSubject: %s\r\n%s\r\n\r\n%s",
		to.String(), from.String(), subject, contentType, content))
}

// sendPlainMail 发送普通SMTP邮件
func sendPlainMail(config *ConfigMapper, from mail.Address, to string, message []byte) error {
	auth := &NotAuth{
		Host:     config.Host,
		Username: config.Username,
		Password: config.Password,
	}
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	return smtp.SendMail(addr, auth, config.Username, []string{to}, message)
}

// sendTLSMail 发送TLS加密邮件
func sendTLSMail(config *ConfigMapper, from mail.Address, to mail.Address, message []byte) error {
	// TLS配置
	tlsConfig := &tls.Config{
		InsecureSkipVerify: config.SkipTLSVerify,
		ServerName:         config.Host,
		MinVersion:         tls.VersionTLS12, // 只支持TLS 1.2及以上版本
	}
	// 建立TLS连接
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port), tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to create TLS connection: %v", err)
	}
	defer func(conn *tls.Conn) {
		_ = conn.Close()
	}(conn)
	// 创建SMTP客户端
	smtpClient, err := smtp.NewClient(conn, config.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %v", err)
	}
	defer func(smtpClient *smtp.Client) {
		_ = smtpClient.Quit()
	}(smtpClient)

	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)
	// 身份验证
	if err = smtpClient.Auth(auth); err != nil {
		return fmt.Errorf("authentication failed: %v", err)
	}
	// 发送邮件
	if err = smtpClient.Mail(from.Address); err != nil {
		return fmt.Errorf("failed to set sender: %v", err)
	}
	if err = smtpClient.Rcpt(to.Address); err != nil {
		return fmt.Errorf("failed to set recipient: %v", err)
	}
	wc, err := smtpClient.Data()
	if err != nil {
		return fmt.Errorf("failed to send data: %v", err)
	}
	_, err = wc.Write(message)
	if err != nil {
		return fmt.Errorf("failed to write message: %v", err)
	}
	if err = wc.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %v", err)
	}
	return nil
}

// Send 发送邮件
// isHTML: 是否发送HTML格式邮件，默认false（纯文本）
func (m *Email) Send(fromName string, toList []mail.Address, subject, content string, isHTML ...bool) []error {
	if len(toList) == 0 {
		return []error{errors.New("gomail: no recipients")}
	}
	var errs []error
	var mutex sync.Mutex
	var wg sync.WaitGroup

	// 确定邮件格式
	html := false
	if len(isHTML) > 0 {
		html = isHTML[0]
	}

	// 设置内容类型
	contentType := "Content-Type: text/plain; charset=UTF-8"
	if html {
		contentType = "Content-Type: text/html; charset=UTF-8"
	}

	// 并发发送邮件
	for _, toAddr := range toList {
		wg.Add(1)
		go func(addr mail.Address) {
			defer wg.Done()

			config, ok := m.GetMapper(addr.Address)
			if !ok {
				return
			}

			from := mail.Address{
				Name:    fromName,
				Address: config.Username,
			}
			message := buildMessage(from, addr, subject, contentType, content)
			var err error

			if !config.TLS {
				// TLS=false时发送普通SMTP邮件
				err = sendPlainMail(config, from, addr.Address, message)
			} else {
				// TLS=true时发送TLS加密邮件
				err = sendTLSMail(config, from, addr, message)
			}

			if err != nil {
				mutex.Lock()
				errs = append(errs, err)
				mutex.Unlock()
			}
		}(toAddr)
	}

	// 等待所有邮件发送完成
	wg.Wait()
	return errs
}
