# 埋点数据采集服务

基于神策SDK的埋点数据采集后端，支持自定义字段，包含字段类型长度。

TODO:
1. [x] sdk接入验证，验证各端SDK（js、Android、ios）接入的数据格式可解析
   1. [x] js sdk
   2. [x] Android sdk
   3. [x] IOS sdk 
2. [x] field配置表新增jsonpath字段，定义字段路径。
3. [x] 新增元数据缓存，event定义表、属性定义表、枚举字典表。（进程内缓存 OR Redis分布式缓存）
4. [x] 服务分层。因为埋点数据接收服务没有ui展示，只提供接口服务，所有没有View层
   1. [x] Controller层：控制层，接收请求并调用相应Service层处理请求（router 路由层）
   2. [x] Service层：业务层，业务逻辑处理（handler）
   3. [x] Dao层：数据访问层，包括数据库访问、缓存访问等
   4. [x] Model层：数据模型层，定义数据模型（项目工程不大时，该层可与Dao层写到一起）
   5. [x] 中间件层：一些通用组件，比如日志组件等
5. [x] event valid，验证event字段是否存在，event值是否已经定义
6. [x] field valid，验证非空字段是否已传，验证枚举值是否已定义，验证字段类型是否正确，验证字段长度是否超长
7. [x] 行为埋点日志数据发送kafka（验证通过/event未定义/属性验证失败）
8. [x] 集成日志组件zap、lumberjack，支持日志写入文件，并可拆分
9. [x] 集成consul配置中心


## 技术栈

* language: [golang](https://golang.org/doc/)
* web：[gin](https://github.com/gin-gonic/gin)
* orm：[gorm](https://github.com/go-gorm/gorm)
* local cache: [go-cache](https://github.com/patrickmn/go-cache)
* redis：[go-redis](https://github.com/go-redis/redis)
* logger：[zap](https://github.com/uber-go/zap)
* logger file rolling：[Lumberjack](https://github.com/natefinch/lumberjack)
* json tool：[gabs](https://github.com/Jeffail/gabs)
* configuration: [viper](https://github.com/spf13/viper)