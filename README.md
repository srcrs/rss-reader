# 简述

RSS可以将信息聚合，曾寻找过一些RSS客户端，但觉得都太过于复杂，会需要登陆、保存历史消息、
使用缓存加快响应速度，但我想要看到的是，打开页面看到关注网站的热点消息即可（一般通过RSS订阅获取到的数据即是热点），
看到有感兴趣的信息，可以跳转过去再详细的了解。

# 配置文件

配置文件位于config.json，sources是RSS订阅链接，示例如下

```json
{
    "sources": [
        "https://rsshub.asksowhat.cn/v2ex/topics/latest",
        "https://rsshub.asksowhat.cn/36kr/news/latest",
        "https://rsshub.asksowhat.cn/aliyun/developer/group/alitech",
        "https://rsshub.asksowhat.cn/blogread/newest",
        "https://rsshub.asksowhat.cn/juejin/category/backend",
        "https://rsshub.asksowhat.cn/edrawsoft/mindmap/8/PV/DESC/CN/1",
        "https://hostloc.com/forum.php?mod=rss&amp;fid=45&amp;auth=3ba611tSbtZSmrvt5Zo2lBgahajeORVteWbX8IarKV66xIEkPiuIRFG2g5x0tQ"
    ]
}
```

# 使用方式

该项目仅仅只有一个html文件，将RSS订阅链接抽离为JSON配置文件，可以部署在GitHub Pages、VerCel等支持静态网页平台，也可
在自己搭建的web容器中如Nginx。无论使用哪种都需要解决跨域问题，我所设想的最好方案不需要依赖外部的网络，但是没能找到合适
的，当前是通过[cors-anywhere](https://github.com/Rob--W/cors-anywhere)来解决的，需要在自己的服务器部署，当前内置的cors代理
是在我自己服务器部署的，这部分最好能自己去搭建。

## Docker部署(非自建cors代理)

环境要求：Git、Docker、Docker-Compose

克隆项目

```bash
git clone https://github.com/asksowhat/ownrss.git
```

进入ownrss文件夹，运行项目

```bash
docker-compose up -d ownrss
```

最后通过ip+端口号访问即可

## Docker部署(自建cors代理，推荐)

```bash
git clone https://github.com/asksowhat/ownrss.git
```

进入ownrss文件夹，运行项目，这里于上一步有区别

```bash
docker-compose up -d
```

更改index.html中的代理

```html
const CORS_PROXY = 'https://cors.asksowhat.cn/'
```

替换为如下形式，IP是你服务器的ip，PORT即是cors-anywhere容器对外暴漏的端口，在项目中，该端口为10015

```html
const CORS_PROXY = 'http:IP:PORT/'
```

更改完之后保存，最后通过ip+端口号访问


## 非Docker部署

这里不通过Docker部署教程暂时不提供，大家可以参考该[cors-anywhere](https://github.com/Rob--W/cors-anywhere)项目说明，如何在没有服务器的情况下搭建cors代理，然后替换掉index.html中的代理CORS_PROXY，后面就可以在各种免费的Pages服务中使用了。

# 依赖项目说明及感谢

- [alpine](https://github.com/alpinejs/alpine)类似于一个极简的vue，可以非常方便的将标签和参数绑定

- [bulma](https://github.com/jgthms/bulma)非常好看的ui组件

- [cors-anywhere](https://github.com/Rob--W/cors-anywhere)解决了跨域的问题

- [rss-parser](https://github.com/rbren/rss-parser)获取RSS订阅的内容，并将其转换为对象

- [jquery](https://github.com/jquery/jquery)获取本地的JSON文件，并转换为对象
