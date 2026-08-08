# 普通用户域名控制台与注册策略实现计划

## 1. 目标与范围

普通用户直接输入根域或子域创建 CDN 域名，完成 TXT 验证后进入域名级控制台，并在套餐额度内使用 Pages、TLS、WAF 和部署能力。管理员创建的全局 WAF 规则、系统 IP 组和默认限流策略对普通用户只读可见。

本阶段同时修复注册开关语义，增加注册邮箱域名白名单、邮箱别名归一化、根域/子域 DNS TXT 所有权验证以及固定 CNAME 接入提示。用户资源必须按 `owner_id` 隔离，旧数据继续归管理员所有（`owner_id=0`）。

## 2. 关键决策

- `registration_enabled` 是注册总开关；`password_register_enabled` 只控制密码注册方式，关闭时页面说明原因而不是伪装成系统未开放注册。
- 邮箱原文保留用于展示，新增 `email_normalized` 用于登录、注册唯一性和别名去重。所有邮箱去除 `+alias`；仅 `gmail.com` 与 `googlemail.com` 去除 local-part 中的点号。
- 根域声明拥有全部权利时验证 `_openflare-verification.<root>`，子域继承根域状态；否则每个子域单独验证。
- 普通用户 CNAME 必须指向 `cname.edge.infvar.com`。TXT 是所有权验证唯一方式，CNAME 仅用于接入检查和用户提示。
- 普通用户 WAF 规则带 `owner_id`，并绑定一个子域对应的代理路由；全局规则 `is_global=true` 优先且不可编辑。系统 IP 组可被规则引用但普通用户不可创建或修改。

## 3. API 与数据

普通用户接口继续挂在 `/api/v1/custom/resources`，新增域名验证、TLS、WAF 只读/自定义规则和 Pages owner-scoped 路由。所有资源不存在或不属于当前用户均返回 404。

## 4. 验证计划

- Go 单测覆盖邮箱归一化、注册域名匹配、TXT 记录解析和 owner 查询。
- 运行 `make swagger`、`make format`、`make code-check` 与相关 `go test`。
- 前端构建使用 `pnpm build:embed`，确认普通用户导航包含 Pages，资源页面按三层流程工作。
