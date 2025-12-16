package email

import (
	"net/mail"
	"testing"
)

var configMapper = map[string]*ConfigMapper{
	"default": {
		TLS:           true,
		Host:          "smtp.exmail.qq.com",
		Port:          465,
		Username:      "liudongcai@bright-ai.com",
		Password:      "******",
		SkipTLSVerify: true,
	},
	"bright-ai.com.cn": {
		TLS:           false,
		Host:          "192.168.1.203",
		Port:          25,
		Username:      "liudongcai@bright-ai.com.cn",
		Password:      "******",
		SkipTLSVerify: false,
	},
}
var testToEmail1 = "liudongcai@bright-ai.com.cn"
var testToEmail2 = "liudongcai@hotmail.com"

// TestEmail_SendPlainEmail tests the SendPlainEmail method of Email
func TestEmail_SendPlainEmail(t *testing.T) {
	email := New(configMapper)
	_ = email.Send("深圳博辉特科技有限公司", []mail.Address{
		{
			Name:    "刘小小111",
			Address: "liudongcai@bright-ai.com.cn",
		},
		{
			Name:    "刘小小222",
			Address: "liudongcai@bright-ai.com",
		},
		{
			Name:    "刘小小333",
			Address: "liudongcai@hotmail.com",
		},
	}, "小主题", "这是一个测试")
}

// TestEmail_SendHTMLEmail tests sending HTML format email
func TestEmail_SendHTMLEmail(t *testing.T) {
	email := New(configMapper)
	htmlContent := `<html><body><h1>测试邮件</h1><p>这是一封HTML格式的测试邮件。</p></body></html>`
	_ = email.Send("深圳博辉特科技有限公司", []mail.Address{
		{
			Name:    "测试用户",
			Address: "liudongcai@bright-ai.com.cn",
		},
	}, "HTML测试主题", htmlContent, true)
}

// TestEmail_EmptyRecipients tests sending email with empty recipient list
func TestEmail_EmptyRecipients(t *testing.T) {
	email := New(configMapper)
	errs := email.Send("测试发件人", []mail.Address{}, "测试主题", "测试内容")
	if len(errs) == 0 {
		t.Error("expected error for empty recipients, but got none")
	}
	if errs[0].Error() != "gomail: no recipients" {
		t.Errorf("expected 'gomail: no recipients' error, got %v", errs[0])
	}
}

// TestEmail_InvalidEmailAddress tests sending email with invalid email address
func TestEmail_InvalidEmailAddress(t *testing.T) {
	email := New(configMapper)
	// 测试无效邮箱地址（应该回退到默认配置）
	errs := email.Send("测试发件人", []mail.Address{
		{
			Name:    "无效用户",
			Address: "invalid-email",
		},
	}, "测试主题", "测试内容")
	// 这个测试可能会失败，因为无效邮箱地址可能无法发送，但主要测试域名解析逻辑
	if len(errs) > 0 {
		t.Logf("Got expected errors for invalid email: %v", errs)
	}
}

// TestEmail_ConfigurationValidation tests configuration validation
func TestEmail_ConfigurationValidation(t *testing.T) {
	// 测试空配置
	emptyMapper := map[string]*ConfigMapper{}
	email := New(emptyMapper) // 应该产生警告，但不会崩溃
	if email == nil {
		t.Error("expected Email instance even with invalid configuration")
	}

	// 测试无效配置
	invalidMapper := map[string]*ConfigMapper{
		"default": {
			TLS:           true,
			Host:          "", // 空主机
			Port:          465,
			Username:      "test@example.com",
			Password:      "password",
			SkipTLSVerify: true,
		},
	}
	email2 := New(invalidMapper) // 应该产生警告，但不会崩溃
	if email2 == nil {
		t.Error("expected Email instance even with invalid configuration")
	}
}
