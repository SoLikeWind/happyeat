# HappyEat 管理后台（前端）

与 HappyEat 后端配套的 React 前端，技术栈：Vite + React 18 + TypeScript + Ant Design + React Router + Axios。

## 开发

```bash
# 安装依赖（若未安装）
npm install

# 启动开发服务，默认 http://localhost:5173
npm run dev
```

请先启动后端服务（如 `localhost:8888`），并在 `.env.development` 中配置 `VITE_API_BASE`。

## 构建

```bash
npm run build
```

产物在 `dist/`，可部署到任意静态托管。

## 目录说明

- `src/api/`：接口封装（client、types、auth）
- `src/contexts/`：全局状态（如 AuthContext）
- `src/layouts/`：布局（侧栏 + 顶栏）
- `src/pages/`：页面（登录、菜单、餐桌、订单、工作台）

接口约定以项目根目录的 `happyeat.json`（Swagger）为准，后端变更后请同步该文件。
