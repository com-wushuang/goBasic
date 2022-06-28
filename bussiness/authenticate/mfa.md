## MFA必要性
- 为了提高安全性，提出了两步认证（`2-Step Verification`，又称多因素认证，`Multi-Factor Authentication`）方式。
- 除了使用密码认证外，再增加一个认证因素，只有两步认证都通过，用户身份的认证过程才算完成。
- 第二种认证因素的形态和传输渠道与密码差异很大，如银行常见的有通过短信发送认证码，定时变化的数字 `token` (`Time-based One-Time Password`)等。
- 增加了一种认证因素，增加了攻击者的难度。

## totp原理
多因素认证中，使用最方便的就是 `TOTP` ,服务器侧认证用户身份的工作过程和原理如下：

**前提条件**
- 服务器侧和用户的 `TOTP` 设备预先有个双方约定的同一个密钥 `K` (每个人的均不同)和一个算法
- 算法可以根据时间戳和密钥K计算出 `6` 位数字 (`RFC6238 TOTP: Time-Based One-Time Password Algorithm`)

**验证过程**
- `TOTP` 设备: 根据时间戳和密钥K计算出 `6` 位数字，显示给用户。
- 用户: 将这 `6` 位数字交给服务器。
- 服务器: 使用同样的算法计算出 `6` 位数字，如果与用户提交的相同，用户认证成功，否则认证失败。
- 考虑到双方时间可能有偏差，用户输入也需要时间，因此服务器在验证时往往会计算当前时刻前后几分钟的 `6` 位数字，只要有一个相同即认为认证成功。

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
- 生成的 `key` 信息格式类似：`otpauth://totp/Example.com:bob@example.com?algorithm=SHA1&digits=6&issuer=Example.com&period=30&secret=JBSWY3DPEHPK3PXP`
- 一般可将该 `key` 信息保存到一个二维码中，用于被 `Authenticator application` 扫描，`application` 会将二维码中的信息进行存储，包括 `secret` 和 `algorithm` ，达到 `secret` 和 `algorithm` 共享的目的。

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

**服务端验证**
```go
// ValidateCustom validates a TOTP given a user specified time and custom options.
// Most users should use Validate() to provide an interpolatable TOTP experience.
func ValidateCustom(passcode string, secret string, t time.Time, opts ValidateOpts) (bool, error) {
	if opts.Period == 0 {
		opts.Period = 30
	}

	counters := []uint64{}
	counter := int64(math.Floor(float64(t.Unix()) / float64(opts.Period)))

	counters = append(counters, uint64(counter))
	for i := 1; i <= int(opts.Skew); i++ {
		counters = append(counters, uint64(counter+int64(i)))
		counters = append(counters, uint64(counter-int64(i)))
	}

	for _, counter := range counters {
		rv, err := hotp.ValidateCustom(passcode, counter, secret, hotp.ValidateOpts{
			Digits:    opts.Digits,
			Algorithm: opts.Algorithm,
		})

		if err != nil {
			return false, err
		}

		if rv == true {
			return true, nil
		}
	}

	return false, nil
}


// ValidateCustom validates an HOTP with customizable options. Most users should
// use Validate().
func ValidateCustom(passcode string, counter uint64, secret string, opts ValidateOpts) (bool, error) {
	passcode = strings.TrimSpace(passcode)

	if len(passcode) != opts.Digits.Length() {
		return false, otp.ErrValidateInputInvalidLength
	}

	otpstr, err := GenerateCodeCustom(secret, counter, opts)
	if err != nil {
		return false, err
	}

	if subtle.ConstantTimeCompare([]byte(otpstr), []byte(passcode)) == 1 {
		return true, nil
	}

	return false, nil
}
```
- 因为证明方和校验方都是基于时间来计算 `OTP` ，如果证明方在一个时间片段的最后时刻发送 `OTP` ，在请求达到校验方时，已经进入下一个时间片段，如果校验方使用当前时间来计算 `OTP` ，肯定会匹配失败，这样会导致一定的失败率，影响可用性。
- 校验方应该不仅仅以接收请求的时间，还应该用上一个时间片段来计算 `TOTP` ，增强容错性。不过，容错窗口越长，被攻击风险越高，“后向兼容”一般推荐不超过一个时间片段。
- 其中 `opts.Skew` 便是这个向后兼容性的设置，如果 `opts.Skew` 被设置为 `1` ，则有效的 `code` 码包括当前 `code` 码，前一个 `code` 码以及下一个 `code` 码。

## 参考
- https://bg6cq.github.io/ITTS/security/mfa/
- https://bbs.huaweicloud.com/blogs/205528