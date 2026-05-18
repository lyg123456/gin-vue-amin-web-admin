# portal-loadtest — 门户 HTTP 并发压测（独立工具）

与 Gin-Vue-Admin 主服务**解耦**：单独目录、单独 `go.mod`，不参与业务 API 编译，避免误部署成「对外攻击接口」。

## 使用前提（必读）

- **仅**对您**拥有**或已取得**书面压测授权**的目标执行。
- 未经授权对他人站点高 QPS 访问可能构成违法；使用者自行承担法律责任。
- 本工具用于容量规划、压测验收、自有环境演练。

## 功能说明

- **高峰时段**（`windows` 内）：**每秒**一轮 `[qps_min, qps_max]` 随机 QPS 压测（这是日志里 `peak 1s` / `peak summary` 的来源，**不是**「每分钟一次」）。
- **非高峰**：发 **1 次** HTTP 心跳，再睡 **`idle_interval`**（默认 **5 分钟**）；若下一高峰更早到来会提前醒来。
- 日志：`peak_log_interval`（默认 **30s**）内汇总一行，避免每秒刷屏；设为 **`0s`** 则恢复每秒详细日志。
- 若你在 **09:00–12:00** 等窗口内运行，会一直走**高峰模式**；若只想**每 5 分钟发 1 次**请求、不要高峰，请在配置里设 **`peak_disabled: true`**（与 `windows` 无关，始终心跳）。
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

- **`idle_interval`**：非高峰 / 心跳模式下两次请求间隔（默认 **`5m`**）；下一高峰若更早到达且未设 `peak_disabled`，会缩短本次 sleep。
- **`peak_disabled: true`**：永不跑高峰，只按 `idle_interval` 发心跳。
- **`peak_log_interval`**：高峰时汇总日志间隔（默认 `30s`）；`0s` 为每秒详细日志。
- 浏览器地址里的 `#/...` **不会**发给服务器；压静态入口请写 **`http://host/web/`** 这类真实 HTTP 路径。
- `concurrency` 过低时实际 QPS 会低于目标；过高可能触发本机 `ulimit` 限制，需自行调优。

## 与「单独模块」的对应关系

| 层次 | 说明 |
|------|------|
| 仓库位置 | `tools/portal-loadtest/` |
| 依赖 | 仅标准库 + `golang.org/x/time/rate` + `gopkg.in/yaml.v3` |
| 不修改 | `server/` 路由、菜单、业务表 |

若希望并入 `server` 单模块，可将本目录迁到 `server/cmd/portal_loadtest/` 并删除独立 `go.mod`，改从仓库根 `go run ./server/cmd/portal_loadtest` 执行（需自行处理 module 路径）。
