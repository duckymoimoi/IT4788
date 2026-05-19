# Hospital Navigator — Admin Panel

## Tổng quan

Web admin panel cho hệ thống dẫn đường bệnh viện. Kết nối với backend API tại `https://group3.it4788.sukkaito.id.vn`.

## Tech Stack

| Thành phần | Công nghệ | Lý do |
|-----------|-----------|-------|
| Framework | React 18 + Vite | SPA nhanh, HMR tốt |
| UI Library | Ant Design 5 | Admin-focused, Table/Form/Modal sẵn |
| Routing | React Router v6 | Standard |
| Data Fetching | TanStack Query v5 | Cache, auto-refresh, loading state |
| HTTP Client | Axios | Interceptor JWT |
| Charts | Recharts | Line/Bar/Pie |
| Canvas | Konva.js + react-konva | Vẽ grid, heatmap, map editor |
| State | Zustand | Lightweight global state |
| Icons | @ant-design/icons | Match Ant Design |

## Kiến trúc

```
Browser ──► React SPA ──► Axios ──► Backend API (Go/Gin)
                │                      │
           TanStack Query          PostgreSQL
           (cache layer)
```

## Backend API

- **Base URL:** `https://group3.it4788.sukkaito.id.vn/api`
- **Auth:** JWT Bearer token (header `Authorization: Bearer <token>`)
- **Docs:** Swagger tại `/swagger/index.html`
- **Response format:** `{ "code": 1000, "message": "OK", "data": {...} }`

## Quy ước chung

- **Branch:** `main` (production), `develop` (dev), `feature/<tên>` (từng feature)
- **Commit message:** `feat:`, `fix:`, `docs:`, `style:`, `refactor:`
- **Code style:** ESLint + Prettier (đã config sẵn)
- **Component:** 1 file = 1 component, PascalCase
- **API file:** 1 file = 1 module, camelCase
