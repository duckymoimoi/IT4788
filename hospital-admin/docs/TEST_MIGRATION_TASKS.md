# 📋 Tài liệu Phân công — Chỉnh sửa Backend Go theo Test Suite

> **Mục tiêu:** Dựa vào bộ test `Hospital-Navigation-App` (58 files) làm đặc tả, chỉnh sửa Go backend để API contract khớp. Đồng thời bổ sung các tính năng còn thiếu.
>
> **Nguyên tắc:**
> - Go backend là source of truth — KHÔNG thay đổi kiến trúc grid-based
> - Test suite là reference cho validation rules và response format
> - Schema/field name khác không quan trọng, chỉ cần **ý nghĩa đầu ra đúng**


---

###  — Device + Utilities + User

#### Phần 1: Device — Sửa error codes + Thêm CRUD

**Files liên quan:** `handler/device_handler.go`, `service/device_service.go`

**Sửa error codes:**

| API | Code hiện tại | Code test expect | Sửa |
|-----|-------------|-----------------|-----|
| `book` — xe không tồn tại | `8001` (ASSET_NOT_FOUND) | `4004` (INVALID_COORDINATE) | Giữ `8001` — `8001` đúng ý nghĩa hơn. Nhóm test sẽ update |

> Hoặc nếu muốn thống nhất: thêm alias `DEVICE_NOT_FOUND = 8001` vào `response.go` cho rõ nghĩa

**Thêm Device CRUD APIs:**

| API mới | Method | Mô tả |
|---------|--------|-------|
| `POST /api/admin/add_device` | `POST` | Thêm thiết bị mới (xe lăn, beacon) — gắn vào node ID |
| `PATCH /api/admin/edit_device` | `PATCH` | Sửa thông tin thiết bị (status, vị trí) |
| `DELETE /api/admin/del_device` | `DELETE` | Xóa thiết bị |

> ⚠️ Device CRUD KHÔNG ảnh hưởng grid/simulation — device chỉ là overlay trên map

**Luồng `add_device`:**
```
Body: { node_id: 123, type: "wheelchair", status: "available", name: "WL-01" }
1. Validate node_id tồn tại trong POI
2. INSERT vào bảng devices
3. Trả device_id mới
```


## V. API mới cần thêm vào Swagger

| # | API | Method | Person | Mô tả |
|---|-----|--------|--------|-------|
| 1 | `/api/admin/upload_map` | POST | A | Upload file `.map` mới |
| 2 | `/api/admin/upload_output` | POST | A | Upload file `output.json` (MAPF paths) |
| 3 | `/api/admin/get_maps` | GET | A | Danh sách map files |
| 4 | `/api/admin/set_active_map` | POST | A | Chọn map active, reload grid |
| 5 | `/api/route/order_multi` | POST | A | Multi-stop ordered routing |
| 6 | `/api/route/order_unordered` | POST | A | Multi-stop optimized routing |
| 7 | `/api/admin/add_device` | POST | D | Thêm thiết bị |
| 8 | `/api/admin/edit_device` | PATCH | D | Sửa thiết bị |
| 9 | `/api/admin/del_device` | DELETE | D | Xóa thiết bị |

**Tổng API sau khi thêm:** 134 + 9 = **143 API**

---

## VI. Tham chiếu

| Tài liệu | Đường dẫn |
|-----------|-----------|
| Swagger API | `hospital/docs/swagger.yaml` |
| Response codes | `hospital/pkg/response.go` |
| Test response codes | `Hospital-Navigation-App/src/constants/response-codes.ts` |
| Team workflow gốc | `hospital/docs/team_workflow.txt` |
| Test suite repo | `Hospital-Navigation-App/tests/` |
