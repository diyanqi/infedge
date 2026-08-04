# 普通用户 CDN 控制台设计

## 目标

普通用户登录后的首页应当先回答三个问题：当前有多少站点、哪些站点正在运行、下一步应该做什么。控制台面向 CDN 使用者，不展示管理员节点、配置版本和底层 OpenResty 指标，详细资源编辑仍保留在“我的资源”页面。

## 页面分工

| 页面 | 责任 | 用户操作 |
| --- | --- | --- |
| `/` | 普通用户 CDN 控制台 | 查看资源概况、接入进度、站点状态，进入常用操作 |
| `/resources` | 资源配置工作区 | 管理根域、子域、源站、站点、HTTPS 与 WAF |
| `/pages` | 静态网站管理 | 创建项目、上传部署包、查看部署历史 |
| `/plans` | 套餐与额度 | 查看和购买套餐 |

管理员访问 `/` 仍使用全局节点与流量总览。前端根据用户身份选择看板，管理员总览接口不会被普通用户调用。

## 数据与权限

控制台只复用现有 owner-scoped API，不新增数据库表和聚合接口：

- `/api/v1/custom/resources/zones`
- `/api/v1/custom/resources/origins`
- `/api/v1/custom/resources/proxy-routes`
- `/api/v1/custom/resources/policies`
- `/api/v1/custom/resources/pages`

所有接口继续由当前登录用户和后端 owner 过滤保护。Pages 前端服务路径必须与后端实际注册的 `/api/v1/custom/resources/pages` 保持一致，管理员通过同一路径读取全局 Pages 项目。

```mermaid
flowchart LR
  User[普通用户] --> Console[CDN 控制台 /]
  Console --> Zones[Zones / 域名]
  Console --> Origins[源站]
  Console --> Routes[CDN 站点]
  Console --> Pages[Pages 项目]
  Console --> Actions[快速操作]
  Actions --> Resources[资源配置 /resources]
  Actions --> PagesPage[Pages /pages]
  Actions --> Plans[套餐 /plans]
```

## 交互决策

- 顶部只保留刷新和“添加站点”两个高频操作，减少用户在首次进入时面对的配置字段。
- 四个概览指标展示站点、域名、源站和 Pages 项目数量；这些数量直接由现有资源查询计算，避免引入新的统计口径。
- 接入进度按“根域、源站、CDN 站点”三个必要步骤计算，点击快速操作进入现有资源工作区完成配置。
- 站点列表只展示状态和域名摘要，详细编辑统一跳转 `/resources`，避免在首页复制复杂表单。
- `/pages` 页面使用 `/api/v1/custom/resources/pages`，修复普通用户和管理员打开 Pages 后列表请求 404 的问题。

## 非目标

本次不改变普通用户资源 API、套餐额度、发布流程和 Pages 部署模型，也不在首页加入实时访问日志或节点健康数据。需要流量统计时，后续应新增明确的 owner-scoped 观测接口并单独定义数据口径。
