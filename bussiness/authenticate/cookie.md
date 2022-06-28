## http状态管理难题

早期Web开发面临的最大问题之一是如何管理状态。因为http是无状态的协议，因此服务器端没有办法知道两个请求是否来自于同一个浏览器。那时的办法是在请求的页面中插入一个token，并且在下一次请求中将这个token返回至服务器。这就需要在form中插入一个包含token的隐藏表单域，或着在URL的qurey字符串中传递该token。这两种办法都强调手工操作并且极易出错。

- 早期Web开发中，通过注入token来管理状态，实现的方式是：服务器生成一个token插入到请求界面中，并在下一次请求时将该token返回至服务器。
- 基于cookie管理状态的方式是：服务器设置cookie，浏览器存储cookie，并在以后每次请求服务期时都将cookie发送给服务器。

以上两种实现方式的本质原理都是一样的，服务器生成信息（用户标识、token、sessionID），客户端存储该信息，并且在以后请求服务器时携带该信息，然后服务器通过解析这些信息来判断这些请求来自同一用户或浏览器。

## cookie是什么

HTTP Cookie（也叫 Web Cookie 或浏览器 Cookie）是服务器发送到用户浏览器并保存在本地的一小块数据，它会在浏览器下次向同一服务器再发起请求时被携带并发送到服务器上。通常，它用于告知服务端两个请求是否来自同一浏览器，如保持用户的登录状态。Cookie 使基于无状态的HTTP协议记录稳定的状态信息成为了可能。

Cookie 主要用于以下三个方面：

- 会话状态管理（如用户登录状态、购物车、游戏分数或其它需要记录的信息）
- 个性化设置（如用户自定义设置、主题等）
- 浏览器行为跟踪（如跟踪分析用户行为等）

## 创建cookie

### 服务端创建cookie

通过HTTP的Set-Cookie消息头（Response Headers），Web服务器可以指定存储一个cookie。Set-Cookie消息的格式如下面的字符串:

```
Set-Cookie:value [ ;expires=date][ ;domain=domain][ ;path=path][ ;secure]
```

消息头的第一部分，value部分，通常是一个name=value格式的字符串（或者一个不包含等号的字符串）。

### 客户端发送cookie

当一个cookie存在，并且可选条件允许的话，该cookie的值会在接下来的每个请求中被发送至服务器。cookie的值被存储在名为Cookie的HTTP消息头中（Request Headers），并且只包含了cookie的值，cookie的选项会被去除。例如:

```
Cookie:value
```

如果在指定的请求中有多个cookies，那么它们会被分号和空格分开，例如：

```
Cookie:value1 ; value2 ; name3=value3
```

### 选项

#### Expires选项

expires指定了cookie何时不会再被发送到服务器端的，因此该cookie可能会被浏览器删掉。格式为`Wdy, DD-Mon--YYYY HH:MM:SS GMT`的值，例如：

```
Set-Cookie:name=Nicholas;expires=Sat, 02 May 2009 23:38:25 GMT
```

没有expires选项时，cookie的寿命仅限于单一的会话中。浏览器的关闭意味这一次会话的结束，所以会话cookie只存在于浏览器保持打开的状态之下。这就是为什么当你登录到一个web应用时经常看到一个checkbox，询问你是否选择存储你的登录信息：如果你选择是的话，那么一个expires选项会被附加到登录的cookie中。如果expires选项设置了一个过去的时间点，那么这个cookie会被立即删除。

#### Max-Age选项

expires是 http/1.0协议中的选项，在新的http/1.1协议中expires已经由 max-age选项代替，两者的作用都是限制cookie 的有效时间。expires的值是一个时间点（cookie失效时刻= expires），而max-age的值是一个以`秒`为单位时间段。默认（缺省）情况下，有效期为session。

#### Domain选项

domain指示cookie将要发送到哪个域或那些域中。默认情况下，domain会被设置为创建该cookie的页面所在的域名。domain选项被用来扩展cookie值所要发送域的数量。例如:

```
Set-Cookie:name=Nicholas;domain=nczonline.net
```

想象诸如[Yahoo](http://www.yahoo.com/)[！](http://www.yahoo.com/)这样的大型网站都会有许多以name.yahoo.com(例如：[my.yahoo.com](http://my.yahoo.com/)、[finance.yahoo.com](http://finance.yahoo.com/)等等)为格式的站点。将cookie的domain选项设置为yahoo.com，就能够让浏览器发送该cookie到所有这些站点。浏览器会对domain的值与请求所要发送至的域名，做一个尾部比较（即从字符串的尾部开始比较），并且在匹配后发送一个Cookie消息头。

####  Path选项

path也是用来控制何时发送cookie。将path属性值与请求的URL头部比较。如果字符匹配，则发送Cookie消息头，例如：

```
Set-Cookie:name=Nicholas;path=/blog
```

 在这个例子中，path选项值会与/blog,/blogrool等等相匹配；任何以/blog开头的选项都是合法的。只有在domain选项核实完毕之后才会对path属性进行比较。

#### Secure选项

不像其它选项，secure选项只是一个标记并且没有其它的值。一个secure cookie只有当请求是通过SSL和HTTPS创建时，才会发送到服务器端。例如：

```
Set-Cookie:name=Nicholas;secure
```

实际应用中，机密且敏感的信息绝不应该在cookies中存储或传输，因为cookies的整个机制都是原本不安全的。默认情况下，在HTTPS链接上传输的cookies都会被自动添加上secure选项。

#### HttpOnly选项

微软的IE6在cookies中引入了一个新的选项：HttpOnly。HttpOnly意思是告之浏览器该cookie绝不应该通过 Javascript 的document.cookie属性访问。设计该特征意在提供一个安全措施来帮助阻止通过 Javascript 发起的跨站脚本攻击(XSS)窃取cookie的行为。要创建一个HttpOnly cookie，只要向你的cookie中添加一个HttpOnly标记即可：

```
Set-Cookie: name=Nicholas; HttpOnly
```

一旦设定这个标记，通过documen.coookie则不能再访问该cookie。

#### SameSite选项

Chrome51开始，浏览器的 Cookie 新增加了一个SameSite选项，用来防止 CSRF 攻击和用户追踪。SameSite属性用来限制第三方 Cookie。他可以被设置为三个值：

- Strict
- Lax
- None

##### Strict

`Strict`最为严格，完全禁止第三方 Cookie，跨站点时，任何情况下都不会发送 Cookie。换言之，只有当前网页的 URL 与请求目标一致，才会带上 Cookie。

```bash
Set-Cookie: CookieName=CookieValue; SameSite=Strict;
```

这个规则过于严格，可能造成非常不好的用户体验。比如，当前网页有一个 GitHub 链接，用户点击跳转就不会带有 GitHub 的 Cookie，跳转过去总是未登陆状态。

##### Lax

`Lax`规则稍稍放宽，大多数情况也是不发送第三方 Cookie，但是导航到目标网址的 Get 请求除外。

```markup
Set-Cookie: CookieName=CookieValue; SameSite=Lax;
```

导航到目标网址的 GET 请求，只包括三种情况：链接，预加载请求，GET 表单。详见下表。

| 请求类型  |                 示例                 |    正常情况 | Lax         |
| :-------- | :----------------------------------: | ----------: | :---------- |
| 链接      |         `<a href="..."></a>`         | 发送 Cookie | 发送 Cookie |
| 预加载    | `<link rel="prerender" href="..."/>` | 发送 Cookie | 发送 Cookie |
| GET 表单  |  `<form method="GET" action="...">`  | 发送 Cookie | 发送 Cookie |
| POST 表单 | `<form method="POST" action="...">`  | 发送 Cookie | 不发送      |
| iframe    |    `<iframe src="..."></iframe>`     | 发送 Cookie | 不发送      |
| AJAX      |            `$.get("...")`            | 发送 Cookie | 不发送      |
| Image     |          `<img src="...">`           | 发送 Cookie | 不发送      |

设置了`Strict`或`Lax`以后，基本就杜绝了 CSRF 攻击。当然，前提是用户浏览器支持 SameSite 属性。

##### None

Chrome 计划将`Lax`变为默认设置。这时，网站可以选择显式关闭`SameSite`属性，将其设为`None`。不过，前提是必须同时设置`Secure`属性（Cookie 只能通过 HTTPS 协议发送），否则无效。

下面的设置无效。

```bash
Set-Cookie: widget_session=abc123; SameSite=None
```

下面的设置有效。

```bash
Set-Cookie: widget_session=abc123; SameSite=None; Secure
```


## 修改cookie

假如服务器设置了如下cookie：

```
Set-Cookie:name=Nicholas; domain=nczonline.net; path=/blog
```

要想在将来改变这个cookie的值，需要发送另一个具有相同cookie name,domain,path的Set-Cookie消息头。例如：

```
Set-Cooke:name=Greg; domain=nczonline.net; path=/blog
```

这将以一个新的值来覆盖原来cookie的值。**注：仅仅只是改变这些选项的某一个，浏览器会创建一个完全不同的cookie，而不是覆盖原来的cookie**

```
Set-Cookie:name=Nicholas; domain=nczonline.net; path=/
```

两个同时拥有“name”的不同的cookie。如果你访问在 www.nczonline.net/blog 下的一个页面，以下的消息头将被包含进来：

```
Cookie：name=Greg;name=Nicholas
```

## 删除cookie

cookie会被浏览器自动删除，通常存在以下几种原因：

- 会话cookie（Session cookie）在会话结束时（浏览器关闭）会被删除
- 持久化cookie（Persistent cookie）在到达失效日期时会被删除
- 如果浏览器中的cookie限制到达，那么cookies会被删除以为新建cookies创建空间。

**注意：服务器设置完cookie后是无法删除客户端浏览器上的cookie的**

## Cookie的限制

在cookies上存在了诸多限制，来阻止cookie滥用并保护浏览器和服务器免受一些负面影响。有两种cookies的限制条件：cookies的属性和cookies的总大小。

- 原始的规范中限定每个域名下不超过20个cookies，早期的浏览器都遵循该规范。在IE7中增加cookie的限制到50个，与此同时Opera限定cookies个数为30。Safari和Chrome对与每个域名下的cookies个数没有限制。
- 发向服务器的所有cookies的最大数量（空间）仍旧维持原始规范中所指出的：4KB。所有超出该限制的cookies都会被截掉并且不会发送至服务器。

