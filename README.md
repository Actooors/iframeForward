## iframeForward
一个类似反向代理的请求转发程序，为了解决iframe跨域提示`Refused to display 'http://xxx.xxx.com/xxx/xxxx' in a frame because it set 'X-Frame-Options' to 'sameorigin'.`的问题。
### 如何使用
```shell
### 设置一个主要的反向代理API路径
修改常量FirstRequestPath为你想要的路径

### 前端使用
以FirstRequestPath="/getforward/get"为例
iframe的src设为http://localhost:8090/getforward/get?url=https%3A%2F%2Fjuejin.im%2Fwelcome%2Ffrontend%3Futm_source%3Dbootcss
即可欢畅在iframe中浏览https://juejin.im/welcome/frontend?utm_source=bootcss

!!上例仅用来展示使用方法，并非最佳例子
```