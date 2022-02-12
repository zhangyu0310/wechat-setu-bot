# 企业微信机器人-随机色图

通过企业微信机器人，定时向企业微信群中推送随机色图。一次会推送三种格式，`News` `Text` `Image` （因为企业微信有屏蔽功能，可能导致一些`News`格式推送的图片没办法打开，所以加上了后面两种。同时，`Text` 可以设置是否@所有人）

<div style="text-align: center;">
<img src="https://raw.githubusercontent.com/zhangyu0310/wechat-setu-bot/main/pic/Snow.jpg" alt="avatar"/>
</div>

## 参数说明

```shell
.\setuServer.exe -at-all -dl-dir="./pic" -pic-msg -wechat-url="xxx" -intervals=3600 -r18
```

|         参数         |       说明       |                  备注                  |
|:------------------:|:--------------:|:------------------------------------:|
|     `-at-all`      |  具有该选项会自带@所有人  |                bool类型                |
|     `-dl-dir`      |   指定下载图片的路径    |          仅在`-pic-msg`被指定时生效          |
|   `-dump-server`   |   图片转储服务器信息    |               ip:port                |
|    `-dump-url`     |     图片转储域名     |         推送`Text`格式消息的转储图片Url         |
|      `-help`       |       帮助       |                                      |
|    `-intervals`    | 色图推送间隔时间（单位：秒） |       默认60，最小10（别把人家色图服务搞挂了！）        |
|      `-keep`       |     保留本地原图     | 推送`Image`消息需下载图片，该参数表示是否保留图片。默认为true |
|    `-news-msg`     |  是否推送`News`消息  |           bool类型，默认为 true            |
|      `-once`       |  执行一次推送后立即退出   | 将定时逻辑分离，可使用 crontab 等外部定时机制实现更灵活的控制  |
|    `-pic-dump`     |    是否开启图片转储    |                bool类型                |
|     `-pic-msg`     | 是否推送`Image`消息  |         bool类型，有这个参数的具体原因见下          |
|    `-pic-size`     |  可以下载不同尺寸的图片   |            默认为`original`             |
|       `-r18`       |      懂得都懂      |                bool类型                |
|  `-setu-api-url`   |   色图API Url    | 默认为`https://api.lolicon.app/setu/v1` |
|  `-setu-transmit`  |      消息传递      |       将所有消息传递至转储服务，由其广播分发，具体见下       |
|      `-tags`       |      图片标签      |          可以指定推送某种带有某个标签的图片           |
| `-transmit-server` |   消息传递目标服务器    |               ip:port                |
|     `-version`     |      打印版本      |                                      |
|   `-wechat-url`    | 微信机器人Webhook地址 |                  必填                  |

## 使用方法

1. 创建一个新的企业微信机器人，获取到它的Webhook地址
2. 将地址填入参数`-wechat-url`  启动服务

## 关于`-pic-msg`这个参数

企业微信会屏蔽一些网站的`News`消息，为了能在点进链接前看到是否是喜欢的图片，这里做了`Image`消息的推送。

`Image`消息需要先将图片下载到本地（所以需要指定`-dl-dir`图片的下载路径）

由于企业微信要求图片大小不能超过2M，所以对下载好的图片进行了大小调整。群里看到的很有可能不是原图。

## 关于图片转储功能

由于pixiv及其代理服务经常被墙，输出的域名经常无法打开。所以提供了转储功能，可以将服务部署在墙外服务器上，然后转储到墙内服务器。这个功能需要墙内的转储接收模块配合使用。

[EasyPicServer](https://github.com/zhangyu0310/EasyPicServer)

## 关于消息传递

如果不希望只推送到一个webhook上，可以将消息转发至[EasyPicServer](https://github.com/zhangyu0310/EasyPicServer)，由其代为广播。它已经实现了简单的webhook注册页面，开箱即用。为什么不直接使用这个服务广播呢？为了节省跨境服务器的流量费用。。。

## 效果截图

<div style="text-align: center;">
<img src="https://raw.githubusercontent.com/zhangyu0310/wechat-setu-bot/main/pic/%E6%88%AA%E5%9B%BE.png" alt="avatar"/>
</div>

## 感谢

THX [Lolicon API](https://api.lolicon.app/)

愿我们的未来能有光芒照耀！
