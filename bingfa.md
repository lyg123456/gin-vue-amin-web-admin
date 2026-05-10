客户端请求
    ↓
Nginx 反向代理（负载均衡）
    ↓
多实例 Web 服务（Golang/Java/Python）
    ↓
Redis 缓存（热点数据、防击穿）
    ↓
MySQL 数据库（连接池优化）