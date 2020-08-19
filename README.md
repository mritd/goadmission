## goadmission

> 一个用于快速开发 Kubernetes 动态准入控制的脚手架。

### 一、简介

goadmission 是一个开发 Kubernetes 动态准入控制的脚手架，goadmission 集成了以下组件:

- [zap](https://github.com/uber-go/zap) 日志组件提供高性能的日志处理以及定制化
- [cobra](https://github.com/spf13/cobra) 提供友好的终端命令行解析处理
- [json-iterator](https://github.com/json-iterator) 提供高性能的 json 序列化与反序列化
- [gorilla/mux](https://github.com/gorilla/mux) 提供 HTTP 路由处理

### 二、如何使用

克隆本项目到本地，在 [adfunc](https://github.com/mritd/goadmission/tree/master/pkg/adfunc) 添加新的准入控制 WebHook 即可，文件命名请尽量保持一致(`func_*.go`)；
原有的准入控制函数如果不需要可以直接删除，本脚手架会自动加载通过 [init](https://github.com/mritd/goadmission/blob/master/pkg/adfunc/func_print_request.go#L12) 方法注册的准入控制到全局 HTTP 路由。

### 三、补充说明

如果想要增加非准入控制 WebHook 的 HTTP 路由，请在 [route](https://github.com/mritd/goadmission/tree/master/pkg/route) 下新建文件，使用方式与 adfunc 类似。
**不要改动 [main.go](https://github.com/mritd/goadmission/blob/master/main.go#L43) 中的初始化方法顺序，否则可能导致准入控制路由无法正常加载。**
