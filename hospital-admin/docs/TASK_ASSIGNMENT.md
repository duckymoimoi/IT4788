# Phân công nhóm — Admin Panel

## Tổng quan

| Người | Module | Trang | Sprint 1 | Sprint 2 |
|-------|--------|-------|----------|----------|
| **A** | Nền tảng + Dashboard | Login, Dashboard | Setup, Layout, Auth | Dashboard hoàn chỉnh |
| **B** | Map Editor + Engine | Map Editor, Engine Panel | Canvas render (read) | CRUD node + Engine |
| **C** | Flow & Simulation | Flow Monitor, Sim Control | Heatmap + tables | Simulation + Obstacle |
| **D** | Medical, Device & Settings | Medical Dash, Device, Settings | Queue + Room | Device + Settings |
| **E** | SOS & Chat | SOS Dash, Chat | SOS CRUD | Chat WebSocket |

---

## Chi tiết từng người

### 👤 A — Nền tảng + Dashboard (2 trang + shared code)

**Phạm vi:** Mọi thứ "chung" + trang Dashboard

| # | Task | File chính | API |
|---|------|-----------|-----|
| A1 | Init project (Vite + deps) | `vite.config.js`, `package.json` | — |
| A2 | Axios client + JWT interceptor | `src/api/client.js` | — |
| A3 | Auth store (Zustand) | `src/stores/authStore.js` | — |
| A4 | Layout (Sidebar + Header + AuthGuard) | `src/components/Layout/` | — |
| A5 | Router config | `src/App.jsx` | — |
| A6 | Login page | `src/pages/Login.jsx` | `POST /auth/login` |
| A7 | Dashboard — KPI cards | `src/pages/Dashboard.jsx` | `GET /sys/check_version`, `GET /engine/health` |
| A8 | Dashboard — Biểu đồ mật độ 24h | `src/pages/Dashboard.jsx` | `GET /admin/stats_flow` |
| A9 | Dashboard — Alerts panel | `src/pages/Dashboard.jsx` | `GET /flow/get_alerts`, `GET /flow/get_bottlenecks` |
| A10 | Dashboard — Mini heatmap | `src/pages/Dashboard.jsx` | `GET /flow/get_heatmap` |

---

### 👤 B — Map Editor + Engine (2 trang)

**Phạm vi:** Hiển thị bản đồ grid (read) + sửa metadata POI + Engine tính toán

> **Lưu ý:** Map load từ file JSON cố định. Không có add/del node, không có set_weight. Chỉ sửa metadata POI (tên, giờ mở cửa, capacity).

| # | Task | File chính | API |
|---|------|-----------|-----|
| B1 | GridCanvas component (Konva.js) | `src/components/GridCanvas/` | — |
| B2 | Map Editor — render nodes/edges | `src/pages/MapEditor.jsx` | `GET /map/get_floors`, `GET /map/get_nodes`, `GET /map/get_edges` |
| B3 | Map Editor — Edit node metadata (modal) | `src/pages/MapEditor.jsx` | `POST /admin/edit_node` |
| B4 | Map Editor — Set capacity | `src/pages/MapEditor.jsx` | `PATCH /admin/set_capacity` |
| B5 | Map Editor — Search POI | `src/pages/MapEditor.jsx` | `GET /map/search_location`, `GET /map/get_depts` |
| B6 | Map Editor — Floor selector + Landmarks | `src/pages/MapEditor.jsx` | `GET /map/get_floors`, `GET /map/get_landmarks` |
| B7 | Engine — Test pathfinding | `src/pages/EnginePanel.jsx` | `POST /engine/solve` |
| B8 | Engine — MAPF viewer (slider timestep) | `src/pages/EnginePanel.jsx` | `GET /engine/mapf_positions`, `GET /engine/mapf_info` |
| B9 | Engine — Params + cache | `src/pages/EnginePanel.jsx` | `POST /engine/set_params`, `POST /engine/clear_cache`, `GET /engine/convergence` |

---

### 👤 C — Flow & Simulation (2 trang)

**Phạm vi:** Giám sát luồng người + điều khiển mô phỏng

| # | Task | File chính | API |
|---|------|-----------|-----|
| C1 | HeatmapCanvas component | `src/components/HeatmapCanvas/` | — |
| C2 | Flow Monitor — Heatmap live | `src/pages/FlowMonitor.jsx` | `GET /flow/get_heatmap` (auto-refresh 5s) |
| C3 | Flow Monitor — Density lookup | `src/pages/FlowMonitor.jsx` | `GET /flow/get_density` |
| C4 | Flow Monitor — Bottleneck table | `src/pages/FlowMonitor.jsx` | `GET /flow/get_bottlenecks` |
| C5 | Flow Monitor — Forecast chart | `src/pages/FlowMonitor.jsx` | `GET /flow/get_forecast` |
| C6 | Flow Monitor — Obstacle reports (table + resolve) | `src/pages/FlowMonitor.jsx` | `GET /flow/get_obstacles`, `POST /flow/resolve_obstacle` |
| C7 | Flow Monitor — Priority routes (set/expire) | `src/pages/FlowMonitor.jsx` | `POST /flow/set_priority`, `POST /flow/expire_priority`, `GET /flow/get_alerts` |
| C8 | Sim Control — Start/Stop | `src/pages/SimControl.jsx` | `POST /simulate/start`, `POST /simulate/stop` |
| C9 | Sim Control — Status + agent positions | `src/pages/SimControl.jsx` | `GET /simulate/status` |
| C10 | Sim Control — Reset flow data | `src/pages/SimControl.jsx` | `POST /admin/reset_flow` |

---

### 👤 D — Medical, Device & Settings (3 trang)

**Phạm vi:** Quản lý y tế, thiết bị + Cài đặt hệ thống

| # | Task | File chính | API |
|---|------|-----------|-----|
| D1 | Medical Dash — Queue per room | `src/pages/MedicalDash.jsx` | `GET /medical/get_queue` |
| D2 | Medical Dash — Room open hours | `src/pages/MedicalDash.jsx` | `GET /medical/room_open` |
| D3 | Medical Dash — Task table + results | `src/pages/MedicalDash.jsx` | `GET /medical/get_tasks`, `GET /medical/result_status` |
| D4 | Medical Dash — HIS sync button | `src/pages/MedicalDash.jsx` | `POST /medical/sync_now` |
| D5 | Medical Dash — Prescription + History | `src/pages/MedicalDash.jsx` | `GET /medical/get_prescription`, `GET /medical/get_history` |
| D6 | Device Monitor — Stations map | `src/pages/DeviceMonitor.jsx` | `GET /device/stations` |
| D7 | Device Monitor — Wheelchair availability | `src/pages/DeviceMonitor.jsx` | `GET /device/wheelchairs` |
| D8 | Device Monitor — Status tracking | `src/pages/DeviceMonitor.jsx` | `GET /device/status/:id`, `GET /device/track/:id` |
| D9 | Device Monitor — Broken reports | `src/pages/DeviceMonitor.jsx` | `POST /device/report_broken` |
| D10 | Device Monitor — Staff requests queue | `src/pages/DeviceMonitor.jsx` | `POST /device/request_staff` |
| D11 | Settings — Notification management | `src/pages/SystemSettings.jsx` | `GET /notification/get_list`, `DELETE /notification/delete` |
| D12 | Settings — Feedback, FAQ, About | `src/pages/SystemSettings.jsx` | `GET /util/feedback_summary`, `GET /util/faq`, `GET /util/about` |

---

### 👤 E — SOS & Chat (2 trang)

**Phạm vi:** Hỗ trợ khẩn cấp + Chat

| # | Task | File chính | API |
|---|------|-----------|-----|
| E1 | SOS Dashboard — List + filter | `src/pages/SOSDashboard.jsx` | `GET /sos/get_list` |
| E2 | SOS Dashboard — Detail modal | `src/pages/SOSDashboard.jsx` | `GET /sos/get_detail` |
| E3 | SOS Dashboard — Respond + Resolve | `src/pages/SOSDashboard.jsx` | `POST /sos/respond`, `POST /sos/resolve` |
| E4 | Chat Support — Room list | `src/pages/ChatSupport.jsx` | `GET /chat/get_rooms`, `GET /chat/get_unread_count` |
| E5 | Chat Support — Chat window | `src/components/ChatWindow/` | `GET /chat/get_messages`, `POST /chat/send_message` |
| E6 | Chat Support — WebSocket realtime | `src/pages/ChatSupport.jsx` | `WS /ws/chat` |
| E7 | Chat Support — Close room | `src/pages/ChatSupport.jsx` | `POST /chat/close_room` |
