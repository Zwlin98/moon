# Moon —— a sidecar for Skynet

主要特性
1. 与 Skynet Cluster 互相调用，支持 Skynet Cluster 的 RPC 格式，无缝接入 Skynet 的 Cluster 模式 
2. 由 Go 语言编写，能方便的集成 Go 语言所带来的各种生态优势
3. 不侵入 Skynet ，避免了使用 C 对 Skynet 进行扩展时的编程风险
4. API 设计参考 Skynet  Cluster，没有集成难度
5. 简单直观的 Lua 对象封装，可以方便的对 Skynet 集群调用中使用的 Lua 对象进行序列化和反序列化
