# Setup Guide — Hướng dẫn cài đặt & chạy dự án

## Yêu cầu

- **Node.js** >= 18
- **npm** >= 9 (hoặc yarn/pnpm)
- **Git**

## Bước 1: Clone & Install

```bash
git clone https://github.com/duckymoimoi/IT4788-admin-panel.git
cd IT4788-admin-panel
npm install
```

## Bước 2: Config environment

Copy file `.env.example` thành `.env`:

```bash
cp .env.example .env
```

Điền tài khoản admin (xin từ leader):
```env
VITE_API_BASE_URL=https://group3.it4788.sukkaito.id.vn/api
VITE_WS_URL=ws://group3.it4788.sukkaito.id.vn/api/ws
VITE_ADMIN_PHONE=your_phone_here
VITE_ADMIN_PASSWORD=your_password_here
```

Khi develop local với backend local:
```env
VITE_API_BASE_URL=http://localhost:8080/api
VITE_WS_URL=ws://localhost:8080/api/ws
```

## Bước 3: Chạy dev server

```bash
npm run dev
```

Mở `http://localhost:5173` — Hot reload tự động.

## Bước 4: Build production

```bash
npm run build
npm run preview   # xem bản build local
```

## Scripts

| Command | Mô tả |
|---------|-------|
| `npm run dev` | Dev server (port 5173) |
| `npm run build` | Build production |
| `npm run preview` | Preview bản build |
| `npm run lint` | Kiểm tra linting |
| `npm run lint:fix` | Auto-fix lint errors |

---

## Quy trình Git

### 1. Tạo branch

Luôn tạo branch từ `develop`, đặt tên theo format:

```
feature/<tên-trang>
```

```bash
git checkout develop
git pull origin develop
git checkout -b feature/map-editor
```

Ví dụ tên branch theo từng người:

| Người | Branch |
|-------|--------|
| A | `feature/dashboard`, `feature/login` |
| B | `feature/map-editor`, `feature/engine-panel` |
| C | `feature/flow-monitor`, `feature/sim-control` |
| D | `feature/medical-dash`, `feature/device-monitor`, `feature/settings` |
| E | `feature/sos-dashboard`, `feature/chat-support` |

### 2. Quy tắc Commit Message

Dùng format **Conventional Commits** để nhất quán cả nhóm:

```
<type>(<scope>): <mô tả ngắn gọn>
```

#### Các type bắt buộc:

| Type | Khi nào dùng | Ví dụ |
|------|-------------|-------|
| `feat` | Thêm tính năng mới | `feat(map): render nodes trên canvas` |
| `fix` | Sửa bug | `fix(login): xử lý lỗi 401 khi token hết hạn` |
| `style` | Sửa CSS, format code (không ảnh hưởng logic) | `style(dashboard): căn chỉnh KPI cards` |
| `refactor` | Tái cấu trúc code (không thêm/sửa feature) | `refactor(flow): tách HeatmapCanvas thành component riêng` |
| `docs` | Cập nhật tài liệu | `docs: cập nhật API_REFERENCE cho flow module` |
| `chore` | Cấu hình, dependencies, CI/CD | `chore: thêm eslint-plugin-react-refresh` |

#### Scope (phạm vi) — dùng tên module của mình:

| Người | Scope |
|-------|-------|
| A | `login`, `dashboard`, `layout`, `auth` |
| B | `map`, `engine`, `canvas` |
| C | `flow`, `sim`, `heatmap` |
| D | `medical`, `device`, `settings` |
| E | `sos`, `chat` |

#### Ví dụ commit đúng:

```bash
# Thêm feature
git commit -m "feat(map): render grid nodes bằng Konva.js"
git commit -m "feat(sos): thêm bảng danh sách SOS với filter"
git commit -m "feat(dashboard): thêm biểu đồ mật độ 24h"

# Sửa bug
git commit -m "fix(chat): tin nhắn WebSocket bị duplicate"
git commit -m "fix(auth): redirect loop khi token hết hạn"

# Style
git commit -m "style(medical): đổi màu Tag trạng thái"

# Refactor
git commit -m "refactor(flow): dùng useQuery thay useEffect cho heatmap"

# Docs
git commit -m "docs: thêm response mẫu cho /sos/get_list"

# Chore
git commit -m "chore: update antd lên 5.25.0"
```

#### ❌ Commit SAI (tránh):

```bash
git commit -m "update"              # Quá chung chung
git commit -m "fix bug"             # Không rõ bug gì
git commit -m "abc"                 # Vô nghĩa
git commit -m "sửa lỗi đăng nhập"  # Thiếu type prefix
```

### 3. Push + Pull Request

```bash
git push origin feature/map-editor
# Tạo PR vào develop trên GitHub
```

**Tiêu đề PR:** dùng giống commit message, ví dụ: `feat(map): hoàn thành Map Editor page`

### 4. Review + Merge

- Ít nhất 1 người review
- Merge vào `develop`
- Khi xong sprint → merge `develop` vào `main`

### 5. Lưu ý quan trọng

- **KHÔNG push trực tiếp vào `main` hoặc `develop`** — luôn tạo PR
- **KHÔNG sửa file của người khác** — mỗi người chỉ sửa file trong phạm vi của mình (xem TASK_ASSIGNMENT.md)
- **Pull develop thường xuyên** để tránh conflict:
  ```bash
  git checkout develop
  git pull origin develop
  git checkout feature/my-feature
  git merge develop
  ```

---

## Troubleshooting

| Vấn đề | Giải pháp |
|--------|----------|
| CORS error | Backend đã config CORS. Nếu local, dùng proxy trong `vite.config.js` |
| 401 Unauthorized | Token hết hạn → re-login |
| API trả 502 | Backend chưa start (server chỉ online 4h/ngày từ 7PM) |
| `ERR_SSL_PROTOCOL_ERROR` | Dùng `http://` thay vì `https://` trong `.env` |
| Conflict khi merge | Pull develop trước khi push, chỉ sửa file của mình |
