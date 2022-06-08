## 为什么需要这些技术

互联网交换数据的过程中，存在着如下四种问题：

### 窃听

![窃听](https://github.com/com-wushuang/goBasic/blob/main/image/窃听.png)

A和B之间通过互联网发送数据，网络中可能存在第三者窃听了数据内容。“加密技术”，能够解决这个问题。

### 篡改

![篡改](https://github.com/com-wushuang/goBasic/blob/main/image/篡改.png)

网络中的第三者不仅窃听了A和B之间的通信数据，甚至有可能将数据的内容篡改了。“消息认证码”、“数字签名技术” 能够解决这个问题。

### 欺骗

![欺骗](https://github.com/com-wushuang/goBasic/blob/main/image/欺骗.png)

A发送给B数据时，网络中有可能存在第三者冒充了B，让A以为他是在和B通信。相反A也可能被第三者冒充。“消息认证码”、“数字签名技术” 能够解决这个问题。

### 否认

A给B发送完数据之后，A放坚持说这不是他发送的信息。“数字签名技术” 能够解决这个问题。

### 小结

| 问题 | 解决                   |
| ---- | ---------------------- |
| 窃听 | 加密                   |
| 篡改 | 消息认证码 或 数字签名 |
| 欺骗 | 消息认证码 或 数字签名 |
| 否认 | 数字签名               |

**注：“数字签名”和“数字证书”是有区别的，数字证书是用来解决“数字签名”技术中存在的“公钥持有人无法识别”这一问题。**

## 加密技术

### 对称加密

![对称加密](https://github.com/com-wushuang/goBasic/blob/main/image/对称加密.png)

简介：

使用相同的密钥（加密算法）进行加密和解密的加密方法，发送方用密钥将数据加密后，接收方收到数据用相同的密钥进行解密。常见的加密算法是AES、DES、OTP等。

过程：

1. 发送方将密钥发送给接收方
2. 发送方加密数据
3. 接受方用密钥解密
4. 反之亦然

存在问题：

在通信双方建立连接后，发送方需要将双方共用的密钥（不是公钥）发送给接收方，这个时候，网络中的第三者可能会窃听到该密钥，从而通过该密钥解开后续通信过程中的加密数据，“对称加密”需要一种方法来安全的传送密钥。

![对称加密的密钥传送问题](https://github.com/com-wushuang/goBasic/blob/main/image/对称加密的密钥传送问题.png)

### 非对称加密

![非对称加密](https://github.com/com-wushuang/goBasic/blob/main/image/非对称加密.png)

简介：

用于加密的密钥称为“公开密钥” public key，用户解密的密钥成为“私密密钥” private key，常见的算法包括RSA。

过程：

1. 接收方创建一个公钥和一个私钥
2. 接收方将公钥发送给发送方，发送方用接收方的公钥对数据进行加密，形成密文
3. 接收方用自己的私钥解开密文

这样就完成一次从单向数据交换，现实中往往通信是双向的：

1. 通信双方建立连接后，交换公钥，互相持有对方的公钥，当要想对方发送数据是，用对方的公钥来加密数据，形成密文
2. 各自的私钥自己小心保存，用来解密密文，一旦泄漏，那么发送给你的数据可能会被窃听
3. 因为公钥和密文是通过互联网发送的，因此就可能被第三者截取，但是由于没有私钥，所以他无法解开密文
4. 因此非对称加密不会存在对称加密的“密钥传送问题”（在非对称加密中，私钥并不会在网络中传输，通信的双方各自持有）

存在问题：

- 加密和解密都需要时间，通过混合加密来解决该问题。
- 公钥的可靠性，双方建立通信时，交换公钥的过程中，网络的第三者x可能会用自己的公钥来替换真实的公钥。这样子，用x的公钥加密过的密文就能够被x的私钥解密。问题的根源在于通信的双发无法确认他们收到的公要是否是真实可靠的。为了解决这个问题，使用了“数字证书”系统。

![非对称加密的公钥持有人问题](https://github.com/com-wushuang/goBasic/blob/main/image/非对称加密的公钥持有人问题.png)

### 混合加密

对称加密存在“密钥传输过程中被窃取”的问题；非对称加密存在时间复杂度高的问题；混合加密将两者结合，弥补他们的缺点。

- 通信双方发送数据的过程使用对称加密
- 建立通信时，共用的密钥通过非对称加密方式加密后，传递给对方

## 数字签名

在非对称加密中，公钥用来加密，私钥用来解密。如果将这个过程反过来，假设你用私钥对数据进行加密，用公钥进行解密，由于任何人都可能持有你的公钥，因此任何人都可能解开你的密文，作为一种加密形式，这个绝对没有任何意义。但是能够证明的是该密文是你创建的（因为用你的公钥可以解开该密文），这就是数字签名的基础。

在“数字签名”中，只有你本人能够创建的密文可以用作“签名”，严格来说，签名的创建可能是与“加密”不同的计算方法。”数字签名“的思想是，使用本人私钥来创建签名， 并使用公钥对签名进行验证。

数字签名机制：

![数字签名](https://github.com/com-wushuang/goBasic/blob/main/image/数字签名.jpeg)

## 数字证书

“非对称加密”和“数字签名”系统存在不能保证公钥属于谁的问题。因为当A试图将公钥发送给B时，网络中的第三方可能会将公钥替换成自己的公钥。通过“数字证书”系统，我们可以保证谁是公钥的创建者。

![公钥持有人难题](https://github.com/com-wushuang/goBasic/blob/main/image/公钥持有人难题.jpeg)

数字证书的机制：

1. A方有一对公钥（PA）和私钥（SA）
2. A方想要将公钥发送给B方
3. A方需要请求认证机构颁发证书，证明他是PA的所有者
4. 认证机构有自己的公钥（PC）和私钥（SC）
5. A方准备自己的信息（个人电子邮件）+公钥，发送给认证机构
6. 认证机构确认完毕后，用自己的私钥（SC）并基于A的数据创建一个数字签名（注意这个签名时认证机构制作的）
7. 认证机构将数字签名+A方信息+A方公钥，制作成一个文件，发送给A方，这个文件就是A方的数字证书
8. 代替公钥，A方将数字证书发送给B方
9. B方从证书中取出A方信息+A方公钥，并用认证机构的公钥（PC）验证数字证书中的签名来自认证机构
10. 如果验证结果没有问题，那么数字证书无疑是由认证机构签发的，所以数字证书中的PA是可信的
11. 由此，从A方到B方的公开密钥的交付过程完成

![数字证书](https://github.com/com-wushuang/goBasic/blob/main/image/数字证书.jpeg)

在上面的机制中，存在一个问题，B方收到的认证机构的公钥PC，是否真的就是认证机构的公钥呢，因为依然存在被第三方替换的可能性。实际上，认证机构的公钥PC也是作为数字证书的方式被交付的，给认证机构颁发证书的，是更高级别的认证机构，认证机构形成了一个树结构，高级别权威机构为较低级别的机构创建证书。

![CA认证树](https://github.com/com-wushuang/goBasic/blob/main/image/CA认证树.jpeg)

客户端服务器场景流程：

1. 通过从网站接收的带有公钥的数字证书，您可以确认该网站未被第三方欺骗
2. 该证书被称为服务器证书，也是由认证机构颁发的
3. 在个人场景下，数字证书被绑定到一个电子邮件，但是在服务器证书的情况下，它绑定到一个域。由此来防止第三方欺骗

![服务器证书](https://github.com/com-wushuang/goBasic/blob/main/image/服务器证书.jpeg)

## 哈希函数

"哈希函数"是将给定数据转化为固定长度的不规则值的函数。常见的哈希函数算法有：MD4、MD5、SHA-0、SHA-1、SHA-2等。有如下特征：

- 输出值的长度不变，输出数据的长度取决于哈希函数本身。例如，在SHA-1的情况下，它的固定为20个字节。即使输入非常大的数据，输出的哈希值的数据长度也不会改变。同样，不管输出的数据是有多小，哈希值的数据长度也是相同的。
- 相同输入的输出必定相同。
- 相似输入的输出差别很大。
- 输入不同，输出会有很小的概率相同，称之为“哈希值的冲突”。
- 不可能通过哈希值输出逆向求解输入。
- 算法相对简单。

**哈希函数，可以将输入的数据摘要输出，在很多情况下被使用到。**

## 应用场景

- 在互联网上通信的双方，利用数字签名和数字证的认证功能，来确保不被第三方欺骗（中间人攻击）。
- SSH，利用公钥和私钥，非对称加密的安全性。