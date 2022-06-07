---
title: cors
date: 2021-01-14 16:31:50
tags:
---

![跨域请求示例](CORS_example.png)

运行在 `http://domain-a.com` 的JavaScript代码使用[`XMLHttpRequest`](https://developer.mozilla.org/zh-CN/docs/Web/API/XMLHttpRequest)来发起一个到 `https://domain-b.com/data.json` 的请求。

出于安全性，浏览器限制脚本内发起的跨源HTTP请求。 XMLHttpRequest遵循同源策略，这意味着应用程序只能从加载应用程序的同一个域请求HTTP资源，除非响应报文包含了正确CORS响应头。

## 简介

**跨源资源共享**是一种基于[HTTP](https://developer.mozilla.org/zh-CN/docs/Glossary/HTTP) 头的机制，该机制通过允许服务器标示除了它自己以外的其它[origin](https://developer.mozilla.org/zh-CN/docs/Glossary/源)（域，协议和端口），这样浏览器可以访问加载这些资源。CORS需要浏览器和服务器同时支持。目前，所有浏览器都支持该功能。整个CORS通信过程，都是浏览器自动完成，不需要用户参与。对于开发者来说，CORS通信与同源的AJAX通信没有差别，代码完全一样。浏览器一旦发现AJAX请求跨源，就会自动添加一些附加的头信息，有时还会多出一次附加的请求，但用户不会有感觉。**因此，实现CORS通信的关键是服务器。只要服务器实现了CORS接口，就可以跨源通信。**

## CORS应用场景

- 前文提到的由 [`XMLHttpRequest`](https://developer.mozilla.org/zh-CN/docs/Web/API/XMLHttpRequest) 或 [Fetch](https://developer.mozilla.org/en-US/docs/Web/API/Fetch_API) 发起的跨源 HTTP 请求
- Web 字体 (CSS 中通过` @font-face `使用跨源字体资源)。网站可以发布 TrueType 字体资源，并只允许已授权网站进行跨站调用
- 使用 `drawImage` 将 Images/video 画面绘制到 canvas

## 访问控制场景

### 简单请求

某些请求不会触发 CORS 预检请求。本文称这样的请求为“简单请求”。简单请求满足的条件，本文略过，有很多官方资料介绍。请求的过程如下：

![简单请求过程](simple-req-updated.png)

- 请求Header字段 [`Origin`](https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Headers/Origin) 表明该请求来源于 `http://foo.example`
- 服务端返回的 `Access-Control-Allow-Origin: *` 表明，该资源可以被**任意**外域访问。

如果服务端仅允许来自 `http://foo.example` 的访问，应返回：Access-Control-Allow-Origin: `http://foo.example`。这样，除了 `http://foo.example`，其它外域均不能访问该资源

### 预检请求

与简单请求不同，“需预检的请求”要求必须首先使用 [`OPTIONS`](https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Methods/OPTIONS)  方法发起一个预检请求到服务器，以获知服务器是否允许该实际请求。

![预检请求的场景](preflight_correct.png)

1.浏览器检测到，从 JavaScript 中发起的请求需要被预检。首先发起了一个使用 `OPTIONS `方法的预检请求。

请求
```http
OPTIONS /resources/post-here/ HTTP/1.1 #OPTIONS是HTTP/1.1 协议中定义的方法，用以从服务器获取更多信息。
Host: bar.other #请求服务器domain
Origin: http://foo.example #请求源
Access-Control-Request-Method: POST #告知服务器，实际请求将使用POST方法
Access-Control-Request-Headers: X-PINGOTHER, Content-Type #请求将携带两个自定义的请求头
```
响应
```http
HTTP/1.1 200 OK
Access-Control-Allow-Origin: http://foo.example #表明服务器允许访问的源
Access-Control-Allow-Methods: POST, GET, OPTIONS #表明服务器允许客户端使用POST、GET、OPTIONS方法发起请求
Access-Control-Allow-Headers: X-PINGOTHER, Content-Type #表明服务器允许请求中携带字段 X-PINGOTHER 与 Content-Type
Access-Control-Max-Age: 86400 #表明该响应的有效时间为 86400 秒。在有效时间内，浏览器无须为同一请求再次发起预检请求。
```

2.预检请求完成之后，发送实际请求。

### 附带身份凭证的请求

一般而言，对于跨源 [`XMLHttpRequest`](https://developer.mozilla.org/zh-CN/docs/Web/API/XMLHttpRequest) 或 [Fetch](https://developer.mozilla.org/en-US/docs/Web/API/Fetch_API) 请求，浏览器**不会**发送身份凭证信息。如果要发送凭证信息，需要设置 `XMLHttpRequest `的某个特殊标志位。

本例中，`http://foo.example` 的某脚本向 `http://bar.other` 发起一个GET 请求，并设置 Cookies:
```javascript
var invocation = new XMLHttpRequest();
var url = 'http://bar.other/resources/credentialed-content/';

function callOtherDomain(){
  if(invocation) {
    invocation.open('GET', url, true);
    invocation.withCredentials = true; # 标志设置为 true，从而向服务器发送 Cookies。
    invocation.send();
  }
}
```

因为这是一个简单 GET 请求，所以浏览器不会对其发起“预检请求”。但是，如果服务器端的响应中未携带 `Access-Control-Allow-Credentials: true` ，浏览器将不会把响应内容返回给请求的发送者。

过程如下：

![携带cookie的跨域请求](cred-req-updated.png)

请求
```http
GET /resources/access-control-with-credentials/ HTTP/1.1
Host: bar.other
Origin: http://foo.example
Cookie: pageAccess=2
```

响应
```http
HTTP/1.1 200 OK
Set-Cookie: pageAccess=3; expires=Wed, 31-Dec-2008 01:34:53 GMT
Access-Control-Allow-Origin: http://foo.example
Access-Control-Allow-Credentials: true
```

对于附带身份凭证的请求，服务器不得设置 `Access-Control-Allow-Origin` 的值为“`*`”。这是因为请求的首部中携带了 `Cookie` 信息，如果 `Access-Control-Allow-Origin` 的值为“`*`”，请求将会失败。而将 `Access-Control-Allow-Origin` 的值设置为 `http://foo.example`，则请求将成功执行。另外，响应首部中也携带了 Set-Cookie 字段，尝试对 Cookie 进行修改。如果操作失败，将会抛出异常。

## gin-cors

```go
package main

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	// CORS for https://foo.com and https://github.com origins, allowing:
	// - PUT and PATCH methods
	// - Origin header
	// - Credentials share
	// - Preflight requests cached for 12 hours
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://foo.com"},
		AllowMethods:     []string{"PUT", "PATCH"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "https://github.com"
		},
		MaxAge: 12 * time.Hour,
	}))
	router.Run()
}
```

