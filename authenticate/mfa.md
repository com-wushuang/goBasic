## MFA必要性
- 为了提高安全性，提出了两步认证（2-Step Verification，又称多因素认证，Multi-Factor Authentication）方式。
- 除了使用密码认证外，再增加一个认证因素，只有两步认证都通过，用户身份的认证过程才算完成。
- 第二种认证因素的形态和传输渠道与密码差异很大，如银行常见的有通过短信发送认证码，定时变化的数字token(Time-based One-Time Password)等。
- 增加了一种认证因素，增加了攻击者的难度。

## totp原理
多因素认证中，使用最方便的就是TOTP,服务器侧认证用户身份的工作过程和原理如下：

**前提条件**
- 服务器侧和用户的TOTP设备预先有个双方约定的同一个密钥K(每个人的均不同)和一个算法
- 算法可以根据时间戳和密钥K计算出6位数字 (RFC6238 TOTP: Time-Based One-Time Password Algorithm)

**验证过程**
- TOTP设备: 根据时间戳和密钥K计算出6位数字，显示给用户
- 用户: 将这6位数字交给服务器
- 服务器: 使用同样的算法计算出6位数字，如果与用户提交的相同，用户认证成功，否则认证失败
- 考虑到双方时间可能有偏差，用户输入也需要时间，因此服务器在验证时往往会计算当前时刻前后几分钟的6位数字，只要有一个相同即认为认证成功。

## totp 工程实践
开源的totp算法的实现可参考：`https://github.com/pquerna/otp`

**生成key**
```go
// Generate a new TOTP Key.
func Generate(opts GenerateOpts) (*otp.Key, error) {
	// url encode the Issuer/AccountName
	if opts.Issuer == "" {
		return nil, otp.ErrGenerateMissingIssuer
	}

	if opts.AccountName == "" {
		return nil, otp.ErrGenerateMissingAccountName
	}

	if opts.Period == 0 {
		opts.Period = 30
	}

	if opts.SecretSize == 0 {
		opts.SecretSize = 20
	}

	if opts.Digits == 0 {
		opts.Digits = otp.DigitsSix
	}

	if opts.Rand == nil {
		opts.Rand = rand.Reader
	}

	// otpauth://totp/Example:alice@google.com?secret=JBSWY3DPEHPK3PXP&issuer=Example

	v := url.Values{}
	if len(opts.Secret) != 0 {
		v.Set("secret", b32NoPadding.EncodeToString(opts.Secret))
	} else {
		secret := make([]byte, opts.SecretSize)
		_, err := opts.Rand.Read(secret)
		if err != nil {
			return nil, err
		}
		v.Set("secret", b32NoPadding.EncodeToString(secret))
	}

	v.Set("issuer", opts.Issuer)
	v.Set("period", strconv.FormatUint(uint64(opts.Period), 10))
	v.Set("algorithm", opts.Algorithm.String())
	v.Set("digits", opts.Digits.String())

	u := url.URL{
		Scheme:   "otpauth",
		Host:     "totp",
		Path:     "/" + opts.Issuer + ":" + opts.AccountName,
		RawQuery: v.Encode(),
	}

	return otp.NewKeyFromURL(u.String())
}
```
- 生成的key信息格式类似：`otpauth://totp/Example.com:bob@example.com?algorithm=SHA1&digits=6&issuer=Example.com&period=30&secret=JBSWY3DPEHPK3PXP`
- 一般可将该key信息保存到一个二维码中，用于被Authenticator application扫描，application会将二维码中的信息进行存储，包括secret和algorithm，达到secret和algorithm共享的目的。

**生成动态码**

动态码一般情况下，由客户端生成
![动态码生成](https://github.com/com-wushuang/goBasic/blob/main/image/totp.png)
```go
// GenerateCodeCustom uses a counter and secret value and options struct to create a passcode.
func GenerateCodeCustom(secret string, counter uint64, opts ValidateOpts) (passcode string, err error) {
	// As noted in issue #10 and #17 this adds support for TOTP secrets that are
	// missing their padding.
	secret = strings.TrimSpace(secret)
	if n := len(secret) % 8; n != 0 {
		secret = secret + strings.Repeat("=", 8-n)
	}

	// As noted in issue #24 Google has started producing base32 in lower case,
	// but the StdEncoding (and the RFC), expect a dictionary of only upper case letters.
	secret = strings.ToUpper(secret)

	secretBytes, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return "", otp.ErrValidateSecretInvalidBase32
	}

	buf := make([]byte, 8)
	mac := hmac.New(opts.Algorithm.Hash, secretBytes)
	binary.BigEndian.PutUint64(buf, counter)
	if debug {
		fmt.Printf("counter=%v\n", counter)
		fmt.Printf("buf=%v\n", buf)
	}

	mac.Write(buf)
	sum := mac.Sum(nil)

	// "Dynamic truncation" in RFC 4226
	// <http://tools.ietf.org/html/rfc4226#section-5.4>
	offset := sum[len(sum)-1] & 0xf
	value := int64(((int(sum[offset]) & 0x7f) << 24) |
		((int(sum[offset+1] & 0xff)) << 16) |
		((int(sum[offset+2] & 0xff)) << 8) |
		(int(sum[offset+3]) & 0xff))

	l := opts.Digits.Length()
	mod := int32(value % int64(math.Pow10(l)))

	if debug {
		fmt.Printf("offset=%v\n", offset)
		fmt.Printf("value=%v\n", value)
		fmt.Printf("mod'ed=%v\n", mod)
	}

	return opts.Digits.Format(mod), nil
}
```

