## JWT 解决什么问题？
JWT的主要目的是在服务端和客户端之间以安全的方式来转移声明。主要的应用场景如下所示：
- 认证 `Authentication`；
- 授权 `Authorization`；
- 联合识别；
- 客户端会话（无状态的会话）；
- 客户端机密。

## 名词解释
- `JWS`：`Signed JWT` 签名过的 `jwt`
- `JWE`：`Encrypted JWT` 部分 `payload` 经过加密的 `jwt`；目前加密 `payload` 的操作不是很普及；
- `JWK`：`JWT` 的密钥，也就是我们常说的 `scret`；
- `JWKset`：`JWT key set`在非对称加密中，需要的是密钥对而非单独的密钥，在后文中会阐释；
- `JWA`：当前 `JWT` 所用到的密码学算法；
- `nonsecure JWT`：当头部的签名算法被设定为 `none` 的时候，该 `JWT` 是不安全的；因为签名的部分空缺，所有人都可以修改。

## JWT 组成
一个通常你看到的jwt，由以下三部分组成，它们分别是：
- `header`：主要声明了 `JWT` 的签名算法；
- `payload`：主要承载了各种声明并传递明文数据；
- `signture`：拥有该部分的 `JWT` 被称为 `JWS`，也就是签了名的 `JWS`；没有该部分的 `JWT` 被称为 `nonsecure JWT` 也就是不安全的 `JWT`，此时 `header` 中声明的签名算法为 `none`。

**Header**
```json
{  
  "typ": "JWT",  # 类型
  "alg": "none",  # 算法
  "jti": "4f1g23a12aa"  # JWT ID，代表了正在使用JWT的编号，这个编号在对应服务端应当唯一
} 
```

**Payload**
```json
{  
  "iss": "http://shaobaobaoer.cn",  
  "aud": "http://shaobaobaoer.cn/webtest/jwt_auth/",  
  "jti": "4f1g23a12aa",  
  "iat": 1534070547,  
  "nbf": 1534070607,  
  "exp": 1534074147,  
  "uid": 1,  
  "data": {  
    "uname": "shaobao",  
    "uEmail": "shaobaobaoer@126.com",  
    "uID": "0xA0",  
    "uGroup": "guest"  
  }  
} 
```
- `payload` 通常由三个部分组成，分别是 
  - `Registered Claims`; 
  - `Public Claims`; 
  - `Private Claims`;
- `Registered Claims`
  - `iss`  【`issuer`】发布者的url地址
  - `sub`  【`subject`】该JWT所面向的用户，用于处理特定应用，不是常用的字段
  - `aud`  【`audience`】接受者的url地址
  - `exp`  【`expiration`】该jwt销毁的时间；unix时间戳
  - `nbf`  【`not before`】该jwt的使用时间不能早于该时间；unix时间戳
  - `iat`  【`issued at`】该jwt的发布时间；unix 时间戳
  - `jti`  【`JWT ID`】该jwt的唯一ID编号
- `Public Claims` 这些可以由使用 `JWT` 的那些标准化组织根据需要定义，应当参考文档 `IANA JSON Web Token Registry`。
- `Private Claims` 这些是为在同意使用它们的各方之间共享信息而创建的自定义声明，既不是注册声明也不是公开声明。上面的 `payload` 中，没有 `public claims` 只有 `private claims`。

## Signature
- `Signature` 部分是对前两部分的签名，防止数据篡改。
- 首先，需要指定一个密钥（`secret`）。这个密钥只有服务器才知道，不能泄露给用户。然后，使用 `Header` 里面指定的签名算法（默认是 `HMAC SHA256`），按照下面的公式产生签名。
```js
HMACSHA256(
  base64UrlEncode(header) + "." +
  base64UrlEncode(payload),
  secret)
```
- 算出签名以后，把 `Header`、`Payload`、`Signature` 三个部分拼成一个字符串，每个部分之间用"点"（.）分隔，就可以返回给用户。