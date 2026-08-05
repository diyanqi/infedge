# 普通用户 CDN 控制台实现计划

## 1. 目标与范围

将普通用户登录后的首页改为参考 EdgeOne 的服务总览，提供网站安全加速与 Pages 两类服务入口；把 `/resources` 调整为网站列表，并将站点级概览与详细配置分层。同时修复 `/pages` 页面调用错误 API 路径导致的 404。

本阶段不新增后端数据表和统计 API，不改变管理员总览、普通用户资源隔离和 Pages 部署逻辑。

## 2. 设计与决策

- 普通用户首页复用现有 owner-scoped Zones、Origins、Proxy Routes、Policies 和 Pages 查询，通过客户端计算概览指标。
- 管理员继续调用 `/api/v1/d/dashboard/overview`；普通用户不再请求管理员接口后再跳转。
- Pages 服务基路径统一为 `/api/v1/custom/resources/pages`，与 `internal/apps/openflare/useraccess` 的注册路径一致。
- `/resources` 负责网站列表，`/resources/detail` 负责单站点控制台，原完整编辑器迁移到 `/resources/configure`；资源变更继续由配置页与 `/pages` 页面完成。

## 3. 修改文件

### 前端 Web

- [NEW] `frontend/app/(main)/components/dashboard/ordinary-dashboard.tsx`：普通用户 CDN 控制台。
- [MODIFY] `frontend/app/(main)/page.tsx`：按用户身份渲染普通用户或管理员看板，并避免普通用户调用管理员接口。
- [MODIFY] `frontend/lib/navigation/openflare-nav.ts`：增加普通用户控制台入口。
- [NEW] `frontend/app/(main)/resources/detail/page.tsx`：普通用户站点控制台与站点级二级导航。
- [NEW] `frontend/app/(main)/resources/configure/page.tsx`：完整资源编辑器的路由入口。
- [MODIFY] `frontend/lib/services/openflare/pages.service.ts`：修正 Pages owner-scoped API 路径。

### 文档

- [NEW] `docs/design/ordinary-user-dashboard.md`：记录页面分工、数据来源和交互边界。
- [MODIFY] `docs/design/index.md`、`docs/design/architecture.md`、`docs/config.ts`：注册设计文档和前端职责说明。
- [MODIFY] `docs/changelog/index.md`：记录控制台改进和 Pages 404 修复。

## 4. 验证计划

- 前端运行 TypeScript 检查、ESLint 和生产构建。
- 确认普通用户进入 `/` 时只请求 `/api/v1/custom/*`，管理员进入 `/` 时仍请求 `/api/v1/d/dashboard/overview`。
- 确认普通用户和管理员访问 `/pages` 时列表请求为 `/api/v1/custom/resources/pages`，不再请求 `/api/v1/custom/pages`。
- 确认普通用户左侧导航保持在页面左侧，并显示“服务总览、网站安全加速、Pages、套餐与用量”。
- 执行仓库要求的 `make format` 和 `make code-check`。
