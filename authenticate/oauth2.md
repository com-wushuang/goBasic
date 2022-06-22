## Oauth2.0 设计思想 
- `OAuth` 在"客户端"与"服务提供商"之间，设置了一个授权层（`authorization layer`）。
- "客户端"不能直接登录"服务提供商"，只能登录授权层，以此将用户与客户端区分开来。
- "客户端"登录授权层所用的令牌（`token`），与用户的密码不同。用户可以在登录的时候，指定授权层令牌的权限范围和有效期。
- "客户端"登录授权层以后，"服务提供商"根据令牌的权限范围和有效期，向"客户端"开放用户储存的资料。

## 整体流程
![oauth_flow](https://github.com/com-wushuang/goBasic/blob/main/image/oauth_flow.png)
- （A）用户打开客户端以后，客户端要求用户给予授权。
- （B）用户同意给予客户端授权。
- （C）客户端使用上一步获得的授权，向认证服务器申请令牌。
- （D）认证服务器对客户端进行认证以后，确认无误，同意发放令牌。
- （E）客户端使用令牌，向资源服务器申请获取资源。
- （F）资源服务器确认令牌无误，同意向客户端开放资源。

不难看出来，上面六个步骤之中，B是关键，即用户怎样才能给于客户端授权。有了这个授权以后，客户端就可以获取令牌，进而凭令牌获取资源。

## 授权模式
客户端必须得到用户的授权（`authorization grant`），才能获得令牌（`access token`）。OAuth 2.0定义了四种授权方式:
- 授权码模式（`authorization code`）
- 简化模式（`implicit`）
- 密码模式（`resource owner password credentials`）
- 客户端模式（`client credentials`）

## 授权码模式
授权码模式（authorization code）是功能最完整、流程最严密的授权模式。它的特点就是通过客户端的后台服务器，与"服务提供商"的认证服务器进行互动。
![authorization_code](https://github.com/com-wushuang/goBasic/blob/main/image/authorization_code.png)
- （A）用户访问客户端，后者将前者导向认证服务器(一般是一个登录用户认证的交互界面)。
- （B）用户认证通过后，选择是否给予客户端授权。
- （C）假设用户给予授权，认证服务器将用户导向客户端事先指定的"重定向 `URI` "（`redirection URI`），同时附上一个授权码。
- （D）客户端收到授权码，附上早先的"重定向 `URI` "，向认证服务器申请令牌。这一步是在客户端的后台的服务器上完成的，对用户不可见。
- （E）认证服务器核对了授权码和重定向 `URI`，确认无误后，向客户端发送访问令牌（`access token`）和更新令牌（`refresh token`）。

**重要的参数**

A步骤中，客户端申请认证的URI，包含以下参数：
- response_type：表示授权类型，必选项，此处的值固定为"code"。
- client_id：表示客户端的ID，必选项。
- redirect_uri：表示重定向URI，可选项。
- scope：表示申请的权限范围，可选项。
- state：表示客户端的当前状态，可以指定任意值，认证服务器会原封不动地返回这个值。
```http request
GET /authorize?response_type=code&client_id=s6BhdRkqt3&state=xyz
        &redirect_uri=https%3A%2F%2Fclient%2Eexample%2Ecom%2Fcb HTTP/1.1
Host: server.example.com
```
C步骤中，服务器回应客户端的URI，包含以下参数：
- code：表示授权码，必选项。该码的有效期应该很短，通常设为10分钟，客户端只能使用该码一次，否则会被授权服务器拒绝。该码与客户端ID和重定向URI，是一一对应关系。
- state：如果客户端的请求中包含这个参数，认证服务器的回应也必须一模一样包含这个参数。(避免中间人攻击)
```http request
HTTP/1.1 302 Found
Location: https://client.example.com/cb?code=SplxlOBeZQQYbYS6WxSbIA
          &state=xyz
```
D步骤中，客户端向认证服务器申请令牌的HTTP请求，包含以下参数：
- grant_type：表示使用的授权模式，必选项，此处的值固定为"authorization_code"。
- code：表示上一步获得的授权码，必选项。
- redirect_uri：表示重定向URI，必选项，且必须与A步骤中的该参数值保持一致。
- client_id：表示客户端ID，必选项。
```http request
POST /token HTTP/1.1
Host: server.example.com
Authorization: Basic czZCaGRSa3F0MzpnWDFmQmF0M2JW # 客户端认证(client_id + client_secret)
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code&code=SplxlOBeZQQYbYS6WxSbIA
&redirect_uri=https%3A%2F%2Fclient%2Eexample%2Ecom%2Fcb
```

E步骤中，认证服务器发送的HTTP回复，包含以下参数：
- access_token：表示访问令牌，必选项。
- token_type：表示令牌类型，该值大小写不敏感，必选项，可以是bearer类型或mac类型。
- expires_in：表示过期时间，单位为秒。如果省略该参数，必须其他方式设置过期时间。
- refresh_token：表示更新令牌，用来获取下一次的访问令牌，可选项。
- scope：表示权限范围，如果与客户端申请的范围一致，此项可省略。
```http request
     HTTP/1.1 200 OK
     Content-Type: application/json;charset=UTF-8
     Cache-Control: no-store
     Pragma: no-cache

     {
       "access_token":"2YotnFZFEjr1zCsicMWpAA",
       "token_type":"example",
       "expires_in":3600,
       "refresh_token":"tGzv3JOkF0XG5Qx2TlKWIA",
       "example_parameter":"example_value"
     }
```

## OIDC
总的来说，`OAuth 2.0` 协议只提供了授权认证，并没有身份认证的功能，而这一缺陷就由 `OIDC` 协议补上了。

`OIDC` 的登录过程与 `OAuth` 相比，最主要的扩展就是提供了 `ID Token`。

**ID Token**

`ID Token` 是一个安全令牌，其数据格式满足 `JWT` 格式，在 `JWT` 的 `Payload` 中由服务器提供一组用户信息。其主要信息包括：
- `iss`(`Issuer Identifier`)：必须。提供认证信息者的唯一标识。一般是一个 `https` 的 `url`（不包含`querystring` 和 `fragment` 部分）；
- `sub`(`Subject Identifier`)：必须。`iss` 提供的用户标识，在 iss 范围内唯一，它有时也会被客户端用来标识唯一的用户。最长为 255 个 ASCII 字符；
- `aud`(`Audiences`)：必须。标识 `ID Token` 的受众。必须包含 `OAuth2` 的 `client_id`；
- `exp`(`Expiration time`)：必须。过期时间，超过此时间的 `ID Token` 会作废；
- `iat`(`Issued At Time`)：必须。JWT 的构建时间；
- `auth_time`(`AuthenticationTime`)：用户完成认证的时间；
- `nonce`：客户端发送请求的时候提供的随机字符串，用来减缓重放攻击，也可以来关联 `ID Token` 和客户端本身的 `Session` 信息；
- `acr`(`Authentication Context Class Reference`)：可选。表示一个认证上下文引用值，可以用来标识认证上下文类；
- `amr`(`Authentication Methods References`)：可选。表示一组认证方法；
- `azp`(`Authorized party`)：可选。结合 aud 使用。只有在被认证的一方和受众（`aud`）不一致时才使用此值，一般情况下很少使用。
- 除了上述这些，ID Token 的用户信息还可以包含其他信息，由服务器端配置。
- 另外 `ID Token` 必须进行 `JWS` 签名和 `JWE` 加密，从而保证认证的完整性、不可否认性以及可选的保密性。

## JWT
**什么是jwt**
- 一个JWT，应该是如下形式的：
```shell
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.  
eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.  
TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ 
```
- 以下三部分组成，它们分别是：
  - `header`：主要声明了JWT的签名算法；
  - `payload`：主要承载了各种声明并传递明文数据；
  - `signture`：拥有该部分的JWT被称为JWS，也就是签了名的JWS；没有该部分的JWT被称为nonsecure JWT 也就是不安全的JWT，此时header中声明的签名算法为none。



