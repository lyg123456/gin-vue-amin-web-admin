# portal-loadtest — 门户 HTTP 并发压测（独立工具）

与 Gin-Vue-Admin 主服务**解耦**：单独目录、单独 `go.mod`，不参与业务 API 编译，避免误部署成「对外攻击接口」。

## 使用前提（必读）

- **仅**对您**拥有**或已取得**书面压测授权**的目标执行。
- 未经授权对他人站点高 QPS 访问可能构成违法；使用者自行承担法律责任。
- 本工具用于容量规划、压测验收、自有环境演练。

## 功能说明

- 在**非活跃时段**：先发 **1 次** HTTP（心跳），再睡眠 **`idle_interval`**（默认 1 小时）；若下一活跃时段早于该间隔结束，会**提前醒来**进入高峰压测，避免睡过头。
- 在活跃时段内，**每秒**在 `[qps_min, qps_max]` 随机一个目标 QPS，使用令牌桶 + 工作协程池尽量逼近该速率发 HTTP GET。
- 默认请求 `target` 与 `extra_paths`（均会 `Join` 到 `target` 上）。

## 运行方式

```bash
cd tools/portal-loadtest
cp config.example.yaml config.yaml
# 编辑 config.yaml：确认 target、时段、QPS、concurrency

go mod tidy
go run . -config config.yaml
```

常用参数：

| 参数 | 说明 |
|------|------|
| `-config` | 配置文件路径，默认 `./config.yaml` |
| `-once` | 仅跑满「当前若处于时段内则 1 分钟」或「立即 30s」用于联调（见代码） |

## 配置项

见 `config.example.yaml`。注意：

- **`idle_interval`**：非活跃时段两次「单请求心跳」之间的间隔（默认 `1h`）；下一活跃时段若更早到达，会缩短本次 sleep。
- 浏览器地址里的 `#/...` **不会**发给服务器；压静态入口请写 **`http://host/web/`** 这类真实 HTTP 路径。
- `concurrency` 过低时实际 QPS 会低于目标；过高可能触发本机 `ulimit` 限制，需自行调优。

## 与「单独模块」的对应关系

| 层次 | 说明 |
|------|------|
| 仓库位置 | `tools/portal-loadtest/` |
| 依赖 | 仅标准库 + `golang.org/x/time/rate` + `gopkg.in/yaml.v3` |
| 不修改 | `server/` 路由、菜单、业务表 |

若希望并入 `server` 单模块，可将本目录迁到 `server/cmd/portal_loadtest/` 并删除独立 `go.mod`，改从仓库根 `go run ./server/cmd/portal_loadtest` 执行（需自行处理 module 路径）。
