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

**authorization code 的作用？为什么需要授权码？**
- 用户授权的过程全程为 `URL` 跳转，额外的信息只能附着在 `url` 后缀中，所以在用户确认后，授权服务器跳转到应用服务器时，不直接携带 `token`，而是携带 `code`，让 `client` 的后台应用去使用 `code` 获取 `token`，降低了 `token` 的泄露风险。
- 通常情况下，`server` 是无法确认 `client` 身份的，此方案中，`client` 需要向 `server` 发送 `secret` 来确认身份。
- `code` 仅能使用一次并且使用时基于一次浏览器 `session`。

## OIDC
总的来说，`OAuth 2.0` 协议只提供了授权认证，并没有身份认证的功能，而这一缺陷就由 `OIDC` 协议补上了。`OIDC` 的登录过程与 `OAuth` 相比，最主要的扩展就是提供了 `ID Token`。

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

## 常见的 token 对比
**access_token**
- `Access Token` 的格式可以是 `JWT` 也可以是一个随机字符串。
- 当携带 `Access Token` 访问受保护的 `API` 接口，`API` 接口检验 `Access Token` 中的 `scope` 权限项目决定是否返回资源，所以 `Access Token` 用于调用接口，而不是用作用户认证。
- 绝对不要使用 `Access Token` 做认证，`Access Token` 本身不能标识用户是否已经认证，`Access Token` 中只包含了用户 `id`，在 `sub` 字段。
- 在你开发的应用中，应该将 `Access Token` 视为一个随机字符串，不要试图从中解析信息，token_data(需要理解 token 和 token_data 之间的区别) 如下:
```
{
  "jti": "YEeiX17iDgNwHGmAapjSQ",
  "sub": "601ad46d0a3d171f611164ce", // subject 的缩写，为用户 ID
  "iat": 1612415013,
  "exp": 1613624613,
  "scope": "openid profile offline_access",
  "iss": "https://yelexin-test1.authing.cn/oidc",
  "aud": "601ad382d02a2ba94cf996c4" // audience 的缩写，为应用 ID
}
```  
- 你希望通过 Access Token 获取更多的用户信息，可以携带 `Access Token` 调用授权服务器的用户信息端点(user_info)来获取完整的用户信息

**id_token**
- `Id Token` 的格式为 `JWT` ，`Id Token` 仅适用于认证场景。
- 不推荐使用 `Id Token` 来进行 `API` 的访问鉴权。
- 每个 `Id Token` 的受众（`aud` 参数）是发起认证授权请求的应用的 `ID`（或编程访问账号的 `AK`）。
```
{
  "sub": "601ad46d0a3d171f611164ce", // subject 的缩写，为用户 ID
  "birthdate": null,
  "family_name": null,
  "gender": "U",
  "given_name": null,
  "locale": null,
  "middle_name": null,
  "name": null,
  "nickname": null,
  "picture": "https://files.authing.co/authing-console/default-user-avatar.png",
  "preferred_username": null,
  "profile": null,
  "updated_at": "2021-02-04T05:02:25.932Z",
  "website": null,
  "zoneinfo": null,
  "at_hash": "xnpHKuO1peDcJzbB8xBe4w",
  "aud": "601ad382d02a2ba94cf996c4", // audience 的缩写，为应用 ID
  "exp": 1613624613,
  "iat": 1612415013,
  "iss": "https://oidc1.authing.cn/oidc"
}
```

**refresh_token**
- `AccessToken` 和 `IdToken` 是 `JSON Web Token`，有效时间通常较短。通常用户在获取资源的时候需要携带 `AccessToken`，当 `AccessToken` 过期后，用户需要获取一个新的 `AccessToken`。
- `Refresh Token` 用于获取新的 `AccessToken`。这样可以缩短 `AccessToken` 的过期时间保证安全，同时又不会因为频繁过期重新要求用户登录。
- 用户在初次认证时，`Refresh Token` 会和 `AccessToken`、`IdToken` 一起返回。你的应用必须安全地存储 `Refresh Token`，它的重要性和密码是一样的，因为 `Refresh Token` 能够一直让用户保持登录。
- 应用携带 `Refresh Token` 向 `Token` 端点发起请求时，授权服务器每次都会返回相同的 `Refresh Token` 和新的 `AccessToken`、`IdToken`，直到 `Refresh Token` 过期。
```
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6InIxTGtiQm8zOTI1UmIyWkZGckt5VTNNVmV4OVQyODE3S3gwdmJpNmlfS2MifQ.eyJqdGkiOiJ4R01uczd5cmNFckxiakNRVW9US1MiLCJzdWIiOiI1YzlmNzVjN2NjZjg3YjA1YTkyMWU5YjAiLCJpc3MiOiJodHRwczovL2F1dGhpbmcuY24iLCJpYXQiOjE1NTQ1Mzc4NjksImV4cCI6MTU1NDU0MTQ2OSwic2NvcGUiOiJvcGVuaWQgcHJvZmlsZSBvZmZsaW5lX2FjY2VzcyBwaG9uZSBlbWFpbCIsImF1ZCI6IjVjYTc2NWUzOTMxOTRkNTg5MWRiMTkyNyJ9.wX05OAgYuXeYM7zCxhrkvTO_taqxrCTG_L2ImDmQjMml6E3GXjYA9EFK0NfWquUI2mdSMAqohX-ndffN0fa5cChdcMJEm3XS9tt6-_zzhoOojK-q9MHF7huZg4O1587xhSofxs-KS7BeYxEHKn_10tAkjEIo9QtYUE7zD7JXwGUsvfMMjOqEVW6KuY3ZOmIq_ncKlB4jvbdrduxy1pbky_kvzHWlE9El_N5qveQXyuvNZVMSIEpw8_y5iSxPxKfrVwGY7hBaF40Oph-d2PO7AzKvxEVMamzLvMGBMaRAP_WttBPAUSqTU5uMXwMafryhGdIcQVsDPcGNgMX6E1jzLA",
  "expires_in": 3600,
  "id_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6InIxTGtiQm8zOTI1UmIyWkZGckt5VTNNVmV4OVQyODE3S3gwdmJpNmlfS2MifQ.eyJzdWIiOiI1YzlmNzVjN2NjZjg3YjA1YTkyMWU5YjAiLCJub25jZSI6IjIyMTIxIiwiYXRfaGFzaCI6Ik5kbW9iZVBZOEFFaWQ2T216MzIyOXciLCJzaWQiOiI1ODM2NzllNC1lYWM5LTRjNDEtOGQxMS1jZWFkMmE5OWQzZWIiLCJhdWQiOiI1Y2E3NjVlMzkzMTk0ZDU4OTFkYjE5MjciLCJleHAiOjE1NTQ1NDE0NjksImlhdCI6MTU1NDUzNzg2OSwiaXNzIjoiaHR0cHM6Ly9hdXRoaW5nLmNuIn0.IQi5FRHO756e_eAmdAs3OnFMU7QuP-XtrbwCZC1gJntevYJTltEg1CLkG7eVhdi_g5MJV1c0pNZ_xHmwS0R-E4lAXcc1QveYKptnMroKpBWs5mXwoOiqbrjKEmLMaPgRzCOdLiSdoZuQNw_z-gVhFiMNxI055TyFJdXTNtExt1O3KmwqanPNUi6XyW43bUl29v_kAvKgiOB28f3I0fB4EsiZjxp1uxHQBaDeBMSPaRVWQJcIjAJ9JLgkaDt1j7HZ2a1daWZ4HPzifDuDfi6_Ob1ZL40tWEC7xdxHlCEWJ4pUIsDjvScdQsez9aV_xMwumw3X4tgUIxFOCNVEvr73Fg",
  "refresh_token": "WPsGJbvpBjqXz6IJIr1UHKyrdVF",
  "scope": "openid profile offline_access phone email",
  "token_type": "Bearer"
}
```