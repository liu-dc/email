# Email 包

一个功能强大、易于使用的Go语言邮件发送库，支持普通SMTP和TLS加密邮件发送，提供并发发送、HTML支持、域名配置管理等高级特性。

## 功能特性

- ✅ **多种发送方式**：支持普通SMTP和TLS加密邮件发送
- ✅ **并发发送**：每个收件人使用独立goroutine处理，大幅提高批量发送效率
- ✅ **格式支持**：同时支持纯文本和HTML格式邮件
- ✅ **智能配置**：基于域名的配置管理，自动选择对应SMTP服务器
- ✅ **安全保障**：可配置TLS证书验证，支持TLS 1.2+版本
- ✅ **错误处理**：完善的错误收集和处理机制，支持批量错误返回
- ✅ **配置验证**：自动验证SMTP配置的有效性，提前发现问题

## 安装

```bash
go get github.com/liu-dc/email
```

## 基本使用

### 配置

```go
import "github.com/liu-dc/email"

// 配置映射 - 支持多个域名的不同SMTP配置
configMapper := map[string]*email.ConfigMapper{
    "default": {
        TLS:            true,           // true: TLS加密, false: 普通SMTP
        Host:           "smtp.example.com",
        Port:           465,            // TLS通常使用465端口
        Username:       "username@example.com",
        Password:       "password",
        SkipTLSVerify:  false,          // 生产环境建议设置为false
    },
    "example.org": {
        TLS:            false,          // 使用普通SMTP
        Host:           "smtp.example.org",
        Port:           25,             // 普通SMTP通常使用25端口
        Username:       "username@example.org",
        Password:       "password",
        SkipTLSVerify:  false,
    },
    "internal.company.com": {
        TLS:            true,
        Host:           "smtp.internal.company.com",
        Port:           587,            // 有些TLS使用587端口
        Username:       "noreply@company.com",
        Password:       "internalpass",
        SkipTLSVerify:  true,           // 内部服务器可能使用自签名证书
    },
}

// 创建Email实例 - 会自动验证配置有效性
emailClient := email.New(configMapper)
```

### 发送纯文本邮件

```go
import (
    "fmt"
    "net/mail"
    "github.com/liu-dc/email"
)

// 准备收件人列表
toList := []mail.Address{
    {
        Name:    "收件人1",
        Address: "recipient1@example.com",
    },
    {
        Name:    "收件人2",
        Address: "recipient2@example.org",
    },
}

// 创建Email实例（配置略）
emailClient := email.New(configMapper)

// 发送纯文本邮件
errs := emailClient.Send(
    "发件人名称",      // 发件人显示名称
    toList,           // 收件人列表
    "重要通知",       // 邮件主题
    "这是一封纯文本邮件的内容。\n\n此致\n敬礼"
)

// 处理发送错误
if len(errs) > 0 {
    fmt.Printf("共 %d 封邮件发送失败:\n", len(errs))
    for i, err := range errs {
        fmt.Printf("  失败 %d: %v\n", i+1, err)
    }
} else {
    fmt.Println("所有邮件发送成功!")
}
```

### 发送HTML格式邮件

```go
import (
    "net/mail"
    "github.com/liu-dc/email"
)

// HTML邮件内容 - 支持完整的HTML标签
htmlContent := `<html>
<head>
    <meta charset="UTF-8">
    <title>欢迎使用Email包</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #f0f0f0; padding: 10px; border-radius: 5px; }
        .content { margin: 20px 0; }
        .footer { font-size: 12px; color: #666; margin-top: 30px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>欢迎使用 Email 包</h1>
        </div>
        <div class="content">
            <p>亲爱的用户：</p>
            <p>这是一封使用 <strong>Email 包</strong> 发送的HTML格式邮件。</p>
            <p>您可以在邮件中使用各种HTML元素，包括：</p>
            <ul>
                <li>格式化文本</li>
                <li>列表</li>
                <li>表格</li>
                <li>图片（需要使用绝对URL）</li>
                <li>样式表</li>
            </ul>
        </div>
        <div class="footer">
            <p>此邮件由系统自动发送，请勿直接回复。</p>
        </div>
    </div>
</body>
</html>`

// 发送HTML邮件（最后一个参数为true表示HTML格式）
errs := emailClient.Send(
    "系统通知",
    toList,
    "HTML邮件功能演示",
    htmlContent,
    true
)
```

## 配置说明

### ConfigMapper 字段说明

| 字段名 | 类型 | 说明 | 默认值 | 约束条件 |
|-------|------|------|-------|---------|
| TLS | bool | 邮件发送方式：true=TLS加密，false=普通SMTP | true | 布尔值 |
| Host | string | SMTP服务器地址 | 必填 | 不能为空字符串 |
| Port | int | SMTP服务器端口 | 必填 | 1-65535之间 |
| Username | string | 发件人用户名（通常是邮箱地址） | 必填 | 不能为空字符串 |
| Password | string | 发件人密码或授权码 | 必填 | 不能为空字符串 |
| SkipTLSVerify | bool | 是否跳过TLS证书验证 | false | 建议生产环境设置为false |

### 常用SMTP端口参考

| 端口 | 用途 | 加密方式 |
|------|------|---------|
| 25 | 普通SMTP | 不加密 |
| 465 | SMTP over SSL | TLS加密 |
| 587 | 邮件提交协议 | STARTTLS |

## API文档

### New(mapper map[string]*ConfigMapper) *Email
创建一个新的Email实例，并自动验证配置有效性。

**参数：**
- mapper: 基于域名的配置映射表，键为域名（或"default"），值为对应SMTP配置

**返回值：**
- *Email: 初始化后的Email实例

**注意：**
- 如果配置验证失败，会输出警告信息但仍然创建实例
- 建议至少配置一个"default"默认配置

### (m *Email) Send(fromName string, toList []mail.Address, subject, content string, isHTML ...bool) []error
并发发送邮件给多个收件人。

**参数：**
- fromName: 发件人显示名称
- toList: 收件人列表，包含名称和邮箱地址
- subject: 邮件主题
- content: 邮件正文内容
- isHTML: 可选参数，是否发送HTML格式邮件（默认false）

**返回值：**
- []error: 发送失败的错误列表，成功则返回空切片

**特性：**
- 每个收件人使用独立goroutine发送，提高发送效率
- 自动根据收件人域名选择对应的SMTP配置
- 支持批量错误收集，不影响其他邮件发送

### (m *Email) GetMapper(email string) (*ConfigMapper, bool)
根据邮箱地址智能获取对应的SMTP配置。

**参数：**
- email: 邮箱地址（如"user@example.com"）

**返回值：**
- *ConfigMapper: 匹配到的SMTP配置
- bool: 是否找到有效配置

**工作原理：**
1. 首先尝试解析邮箱地址，提取域名部分
2. 查找该域名对应的配置
3. 如果未找到，使用默认配置（"default"）
4. 如果默认配置也不存在，返回nil和false

## 高级功能

### 智能配置选择

```go
// 不同域名的收件人会自动使用对应配置
mixedRecipients := []mail.Address{
    {Name: "用户1", Address: "user1@example.com"},     // 使用example.com配置
    {Name: "用户2", Address: "user2@example.org"},     // 使用example.org配置  
    {Name: "用户3", Address: "user3@unknown.com"},     // 使用default配置
}

// 发送邮件 - 自动为每个收件人选择合适的SMTP服务器
errs := emailClient.Send("发件人", mixedRecipients, "主题", "内容")
```

### 并发发送与错误处理

```go
// 批量发送大量邮件
largeRecipients := make([]mail.Address, 100)
for i := 0; i < 100; i++ {
    largeRecipients[i] = mail.Address{
        Name:    fmt.Sprintf("用户%d", i),
        Address: fmt.Sprintf("user%d@example.com", i),
    }
}

// 并发发送，提高效率
errs := emailClient.Send("批量通知", largeRecipients, "批量邮件主题", "批量邮件内容")

// 统计发送结果
if len(errs) > 0 {
    fmt.Printf("发送完成：%d 成功，%d 失败\n", len(largeRecipients)-len(errs), len(errs))
    
    // 分析失败原因
    failureReasons := make(map[string]int)
    for _, err := range errs {
        failureReasons[err.Error()]++
    }
    
    fmt.Println("失败原因统计：")
    for reason, count := range failureReasons {
        fmt.Printf("  %s: %d次\n", reason, count)
    }
} else {
    fmt.Println("所有邮件发送成功！")
}
```

### 配置验证机制

```go
// 故意创建无效配置
invalidConfig := map[string]*email.ConfigMapper{
    "default": {
        MT:             0,
        Host:           "",            // 空主机地址 - 会被验证发现
        Port:           99999,         // 无效端口 - 会被验证发现
        Username:       "",            // 空用户名 - 会被验证发现
        Password:       "",            // 空密码 - 会被验证发现
        SkipTLSVerify:  false,
    },
}

// 创建实例时会输出警告信息
emailClient := email.New(invalidConfig)
```

## 最佳实践

### 1. 安全配置

```go
// 生产环境推荐配置
safeConfig := map[string]*email.ConfigMapper{
    "default": {
        MT:             0,              // 始终使用TLS加密
        Host:           "smtp.secure.com",
        Port:           465,
        Username:       "secure@example.com",
        Password:       "strong_password123",
        SkipTLSVerify:  false,          // 绝不跳过证书验证
    },
}
```

### 2. 密码保护

```go
// 建议从环境变量或配置文件加载密码
import "os"

password := os.Getenv("SMTP_PASSWORD")
if password == "" {
    panic("SMTP_PASSWORD environment variable is required")
}

config := &email.ConfigMapper{
    // ...其他配置
    Password: password,
}
```

### 3. 错误重试策略

```go
// 实现简单的重试机制
maxRetries := 3
var lastErrs []error

for i := 0; i < maxRetries; i++ {
    errs := emailClient.Send(fromName, toList, subject, content)
    if len(errs) == 0 {
        fmt.Println("邮件发送成功！")
        return
    }
    lastErrs = errs
    fmt.Printf("发送失败（尝试 %d/%d）: %v\n", i+1, maxRetries, errs)
    time.Sleep(time.Second * time.Duration(i+1)) // 指数退避
}

fmt.Printf("所有重试都失败: %v\n", lastErrs)
```

## 测试

运行所有测试：

```bash
go test ./...
```

运行测试并显示详细输出：

```bash
go test -v ./...
```

## 注意事项

1. **TLS安全**：
   - 生产环境中请勿设置`SkipTLSVerify: true`，这会带来安全风险
   - 自签名证书的内部服务器可以考虑跳过验证，但建议在可能的情况下使用有效证书

2. **并发限制**：
   - 虽然支持并发发送，但不同SMTP服务器可能有发送频率限制
   - 大量发送时建议实现速率控制或分批发送

3. **错误处理**：
   - Send方法返回所有失败的错误列表，请务必处理这些错误
   - 可以根据错误类型实现不同的处理逻辑

4. **配置管理**：
   - 建议将SMTP配置存储在环境变量或安全的配置文件中
   - 定期更新SMTP密码，避免使用弱密码

5. **邮件内容**：
   - HTML邮件中使用的图片和链接建议使用绝对URL
   - 避免发送过大的邮件内容，可能会被SMTP服务器拒绝

## 故障排除

### 常见问题及解决方案

**问题1：连接超时**
- 检查SMTP服务器地址和端口是否正确
- 确认网络连接是否正常，防火墙是否允许该端口
- 验证SMTP服务器是否要求特定IP段访问

**问题2：认证失败**
- 检查用户名和密码是否正确
- 某些邮箱服务需要使用授权码而非登录密码
- 确认SMTP服务是否已在邮箱设置中启用

**问题3：TLS握手失败**
- 检查TLS配置是否正确
- 确认SMTP服务器是否支持TLS 1.2+版本
- 对于内部服务器，可以尝试设置`SkipTLSVerify: true`

**问题4：收件人不接收邮件**
- 检查收件人邮箱地址是否正确
- 邮件可能被标记为垃圾邮件，建议优化邮件内容
- 确认发件人域名是否配置了正确的SPF和DKIM记录

## 版本历史

### v1.0.0 (2025-12-16)
- 初始版本发布
- 支持TLS和普通SMTP发送
- 并发发送功能
- 纯文本和HTML格式支持
- 基于域名的配置管理
- 完善的错误处理和配置验证

## License

MIT License - 详见LICENSE文件

## 贡献

欢迎提交Issue和Pull Request！

### 贡献指南
1. Fork本仓库
2. 创建特性分支（git checkout -b feature/AmazingFeature）
3. 提交更改（git commit -m 'Add some AmazingFeature'）
4. 推送到分支（git push origin feature/AmazingFeature）
5. 开启Pull Request

## 联系

如有问题或建议，欢迎通过以下方式联系：
- 提交Issue：GitHub Issues页面
- 发送邮件：maintainer@example.com

---

**Email包** - 让Go语言邮件发送变得简单高效！
