
[![License](https://img.shields.io/badge/license-Apache%202-4EB1BA.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)

# gojob
Go-Job是轻量级的任务调度解决方案，使用Go语言编写，协程并发机制可支持大规模任务并行调度。Go-Job为调度节点，你的业务系统就是执行节点，凡是暴露出HTTP服务的业务方法都可以被Go-Job调度，对业务系统完全无侵入。HTTP协议方便夸平台，契合当前主流的微服务架构。

Go-Job通过raft强数据一致性协议、数据库异步多写、执行节点路由等措施，消除调度节点和执行节点的单点隐患，使调度系统具有极高的可用性。无人值守也能高枕无忧。

通过WEB UI进行作业元数据库管理，将散落在各个业务系统的定时任务，收拢到Go-Job进行统一管理和调度。Go-Job使用方便，解压安装包，配置好MySQL连接(表结构会自动创建)，即可运行。
