# 📋 Tài liệu Phân công — Chỉnh sửa Backend Go theo Test Suite

> **Mục tiêu:** Dựa vào bộ test `Hospital-Navigation-App` (58 files) làm đặc tả, chỉnh sửa Go backend để API contract khớp. Đồng thời bổ sung các tính năng còn thiếu.
>
> **Nguyên tắc:**
> - Go backend là source of truth — KHÔNG thay đổi kiến trúc grid-based
> - Test suite là reference cho validation rules và response format
> - Schema/field name khác không quan trọng, chỉ cần **ý nghĩa đầu ra đúng**

---

## I. Thay đổi chung (tất cả modules)

### Response code type
- Test expect **string** (`'1000'`), Go trả **number** (`1000`)
- **Quyết định:** Giữ nguyên Go (number). Team test sẽ tự adapt khi integrate

### Auth mechanism
- Test dùng `token` query param hoặc mock header
- Go dùng JWT qua `Authorization: Bearer <token>` header
- **Quyết định:** Giữ nguyên Go (JWT). Không sửa

---

## II. Phân công theo Person

---

### 👤 Person A (Leader) — Auth + Map + Route + Admin Engine

#### Phần 1: Auth — Tăng cường password validation

**File liên quan:** `service/auth_service.go`

| Việc cần làm | Chi tiết |
|-------------|---------|
| Sửa signup validation | Password phải có: ít nhất 8 ký tự, 1 chữ hoa, 1 số (hiện tại chỉ yêu cầu min 6) |
| Thêm phone regex | Validate phone format: bắt đầu `0` hoặc `+84`, 10-12 chữ số |
| Thêm full_name validation | Không chứa số, max 100 ký tự |

Test cases tham chiếu: `tests/auth/signup.test.ts` — TC-14→TC-18, TC-20→TC-22

---

#### Phần 2: Route — Thêm Multi-stop routing

**Files liên quan:** `handler/route_handler.go`, `service/route_service.go`

| Việc cần làm | Chi tiết |
|-------------|---------|
| Thêm `POST /api/route/order_multi` | Nhận `start_location` + `target_locations[]` (mảng điểm đến), tính route theo thứ tự |
| Thêm `POST /api/route/order_unordered` | Nhận `start_location` + `target_locations[]`, tối ưu thứ tự bằng nearest-neighbor |
| Giới hạn max 10 điểm đến | Trả code `2003` nếu vượt quá |

Logic: Chạy Dijkstra nối tiếp cho từng cặp (A→B→C→D), tổng hợp path + total_distance + estimated_time.

Test cases tham chiếu: `tests/routing/route_planning.test.ts` — `route_ordered`, `route_unordered`

---

#### Phần 3: Admin — Upload Map + Output

**Files liên quan:** `handler/engine_handler.go` (mới), `service/engine_service.go`

| API mới | Method | Mô tả |
|---------|--------|-------|
| `/api/admin/upload_map` | `POST` (multipart) | Upload file `.map` mới (octile grid format) |
| `/api/admin/upload_output` | `POST` (multipart) | Upload file `output.json` (MAPF pre-computed paths) |
| `/api/admin/get_maps` | `GET` | Danh sách map files đã upload |
| `/api/admin/set_active_map` | `POST` | Chọn map đang dùng, reload Dijkstra grid |

**Luồng xử lý `upload_map`:**
```
1. Nhận file multipart → validate format octile (dòng 1: "type octile")
2. Lưu vào data/ với tên file gốc (hoặc timestamp)
3. Nếu admin chọn set active:
   a. Dừng simulation (nếu đang chạy)
   b. Reload grid vào memory
   c. Re-seed POI từ grid mới
   d. Dijkstra dùng grid mới ngay lập tức
4. Simulation cần admin bật thủ công sau khi upload output.json mới
```

**Luồng xử lý `upload_output`:**
```
1. Nhận file output.json → validate JSON format
2. Lưu vào data/
3. Admin có thể bật simulation với output mới qua API start_simulation có sẵn
```

---

#### Phần 4: Admin — Map CRUD

**Nguyên tắc:** Admin CÓ THỂ quản lý maps, nhưng **KHÔNG ĐƯỢC sửa map đang chạy simulation**.

| API | Hành động |
|-----|-----------|
| `upload_map` | ✅ Upload file `.map` mới — lưu vào `data/`, không ảnh hưởng map đang active |
| `upload_output` | ✅ Upload file `output.json` mới — MAPF paths cho map mới |
| `set_active_map` | ✅ Chuyển sang map mới → **tự động dừng simulation** → reload grid → Dijkstra dùng map mới |
| `get_maps` | ✅ Liệt kê tất cả map files |
| `edit_node` (metadata) | ✅ Sửa tên, mô tả, giờ mở cửa, capacity — **KHÔNG sửa grid_row/grid_col** |
| `add_node` / `del_node` | ❌ **BỎ** — phá vỡ grid topology, dùng upload_map thay thế |
| `add_edge` / `del_edge` | ❌ **BỎ** — edges auto-computed từ grid adjacency |
| `set_weight` | ❌ **BỎ** — weights = khoảng cách Euclidean trên grid |

> ⚠️ **Quy tắc vàng:** Muốn thay đổi topology lưới (tường, lối đi) → Dùng Frontend kéo `.map` về sửa (Export) → Upload file `.map` mới + `output.json` mới → Set active → Simulation tự restart.

**Effort Person A:** ~4-5 giờ

---

### 👤 Person B — Flow Module

#### Sửa Flow APIs sang route-based params

**Files liên quan:** `handler/flow_handler.go`, `service/flow_service.go`

Hiện tại Flow APIs dùng `grid_location` (int). Test suite dùng `route_id` (string). Cần sửa để hỗ trợ **cả hai** hoặc chuyển sang route-based.

| API | Param hiện tại | Param cần sửa | Chi tiết |
|-----|---------------|--------------|---------|
| `GET /api/flow/get_density` | `grid_location` (int) | Thêm `route_id` (string) | Nếu có `route_id` → aggregate density cho toàn route. Nếu có `grid_location` → trả density 1 cell |
| `GET /api/flow/get_heatmap` | Không param | Thêm optional `route_id` | Nếu có → filter heatmap cho route. Nếu không → trả toàn bộ |
| `GET /api/flow/get_forecast` | `hours` (int) | Thêm `time_offset` (phút) | Accept cả `hours` và `time_offset`. `time_offset=15` → `hours=0.25` |
| `GET /api/flow/edge_status` | `grid_location` (int) | Thêm `edge_id` (string) | Nếu có `edge_id` → lookup edge, trả status. Nếu có `grid_location` → giữ nguyên |

**Concept mapping Route → Grid:**
```
route_id "R123" → lookup route trong DB → lấy danh sách grid cells → aggregate density
```

Test cases tham chiếu: `tests/flow/get_density.test.ts`, `get_heatmap.test.ts`, `flow_forecast.test.ts`, `edge_status.test.ts`

**Effort Person B:** ~2-3 giờ

---

### 👤 Person C — Medical + Notification

#### Không cần sửa logic — chỉ review

| Module | Files test | Đánh giá | Việc cần làm |
|--------|-----------|----------|-------------|
| Medical (4 files) | tasks_and_queue, room_operations, medical_records, system_sync | ✅ Logic khớp 100% | Review test cases, đảm bảo error codes đúng |
| Notification (3 files) | get_notification, read_notification, del_notification | ✅ Logic khớp 100% | Review test cases |

**Việc duy nhất:** Kiểm tra các error code edge cases trong test có khớp với Go handler hay không. Ví dụ:
- Test expect `5000` cho DB error → Go trả `9999` (UNEXPECTED) → sửa Go trả `5000` cho consistency nếu cần
- Test dùng `token` query param → Go dùng JWT header → không cần sửa

**Effort Person C:** ~30 phút (review only)

---

### 👤 Person D — Device + Utilities + User

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

#### Phần 2: Utilities + User — Review only

| Module | Files test | Đánh giá |
|--------|-----------|----------|
| Util (8 files) | canteen, pharmacy, parking, wifi, weather, FAQ, feedback, upload | ✅ Khớp |
| User (1 file) | set_devtoken | ✅ Khớp |

**Effort Person D:** ~2 giờ (CRUD mới) + 30 phút (review)

---

### 👤 Person E — Chat + SOS

#### Không cần sửa logic — chỉ review

| Module | Files test | Đánh giá | Việc cần làm |
|--------|-----------|----------|-------------|
| Chat (5 files) | create_chat, list_conversations, get_messages, send_messages, mark_read | ✅ Khớp | Review |
| SOS (1 file) | sos_request | ✅ Khớp | Review |
| Chatbot (1 file) | chatbot_query | ❌ Bỏ | API không tồn tại trong Go |

**Effort Person E:** ~30 phút (review only)

---

## III. Tổng kết phân công

| Người | Công việc chính | Code mới | Review | Effort |
|-------|----------------|----------|--------|--------|
| **A (Leader)** | Password validation + Multi-stop route + Upload map/output APIs | ~400 dòng | Auth + Route tests | ~4-5h |
| **B** | Flow APIs route-based params | ~200 dòng | Flow tests | ~2-3h |
| **C** | — | 0 | Medical + Notif tests | ~30 phút |
| **D** | Device CRUD (3 APIs mới) + error codes | ~150 dòng | Device + Util tests | ~2.5h |
| **E** | — | 0 | Chat + SOS tests | ~30 phút |
| **TỔNG** | | ~750 dòng | 55 test files | **~10-12h** |

---

## IV. Files test bị bỏ (7 files)

| File | Lý do |
|------|-------|
| `admin/node_management.test.ts` | add/del node phá vỡ grid → dùng upload_map thay thế |
| `admin/edge_management.test.ts` | Edges auto-computed từ grid |
| `admin/weight_management.test.ts` | Weights = khoảng cách Euclidean, không cho sửa thủ công |
| `admin/device_management.test.ts` | Viết lại CRUD mới cho Go (Person D) |
| `chat/chatbot_query.test.ts` | Go không có chatbot |
| `routing/route_planning.test.ts` (phần unordered) | Thêm mới vào Go (Person A) |
| ~~routing/route_planning.test.ts (phần ordered)~~ | Thêm mới vào Go (Person A) |

---

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
