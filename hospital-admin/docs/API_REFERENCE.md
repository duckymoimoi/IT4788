# API Reference — Danh sách endpoints dùng trong Admin Panel

## Base URL

```
Production: https://group3.it4788.sukkaito.id.vn/api
Local:      http://localhost:8080/api
```

## Authentication

```
POST /auth/login
Body: { "phone_number": "0123456789", "password": "xxx" }
Response: { "code": 1000, "data": { "token": "jwt...", "user": {...} } }

Header cho mọi request private:
Authorization: Bearer <token>
```

## Response format chuẩn

```json
{
  "code": 1000,
  "message": "OK",
  "data": { ... }
}
```

Error codes: 1001 = invalid params, 1002 = unauthorized, 1003 = not found, 1004 = server error

---

## Endpoints theo module

### Auth
| Method | Path | Auth | Mô tả |
|--------|------|------|-------|
| POST | `/auth/login` | ❌ | Đăng nhập |
| POST | `/auth/logout` | ✅ | Đăng xuất |

### Map (Public)
| Method | Path | Auth | Mô tả |
|--------|------|------|-------|
| GET | `/map/get_floors` | ❌ | Danh sách tầng |
| GET | `/map/get_nodes?floor_id=` | ❌ | Nodes trên tầng |
| GET | `/map/get_edges?floor_id=` | ❌ | Edges trên tầng |
| GET | `/map/get_meta` | ❌ | Metadata (grid size) |
| GET | `/map/get_depts` | ❌ | Danh sách khoa |
| GET | `/map/search_location?q=` | ❌ | Tìm kiếm POI |
| GET | `/map/get_landmarks` | ❌ | Landmarks |

### Admin-Map (Admin only)
| Method | Path | Auth | Mô tả |
|--------|------|------|-------|
| POST | `/admin/add_node` | 🔑 | Thêm node |
| POST | `/admin/edit_node` | 🔑 | Sửa node |
| DELETE | `/admin/del_node` | 🔑 | Xóa node |
| POST | `/admin/add_edge` | 🔑 | Thêm cạnh |
| DELETE | `/admin/del_edge` | 🔑 | Xóa cạnh |
| PATCH | `/admin/set_weight` | 🔑 | Sửa trọng số |
| PATCH | `/admin/set_capacity` | 🔑 | Đặt capacity |

### Flow
| Method | Path | Auth | Mô tả |
|--------|------|------|-------|
| GET | `/flow/get_density?grid_location=` | ❌ | Mật độ tại 1 ô (window 5min khi sim chạy) |
| GET | `/flow/get_heatmap` | ❌ | Bản đồ nhiệt (throughput khi sim chạy) |
| GET | `/flow/get_bottlenecks?limit=10` | ❌ | Top N điểm tắc nghẽn |
| GET | `/flow/get_forecast?hours=24` | ❌ | Dự báo theo giờ |
| GET | `/flow/get_alerts` | ❌ | Tuyến ưu tiên active |
| GET | `/flow/edge_status?grid_location=` | ❌ | Trạng thái cạnh |
| POST | `/flow/ping_location` | ✅ | Gửi vị trí |
| POST | `/flow/report_obstacle` | ✅ | Báo cáo vật cản |
| GET | `/flow/get_obstacles?status=&page=&limit=` | ✅ | Danh sách reports |
| POST | `/flow/set_priority` | ✅ | Đặt tuyến ưu tiên |
| POST | `/flow/expire_priority` | ✅ | Hủy ưu tiên |
| POST | `/flow/resolve_obstacle` | ✅🔑 | Xử lý vật cản (Staff) |

### Admin-Flow
| Method | Path | Auth | Mô tả |
|--------|------|------|-------|
| GET | `/admin/stats_flow?hours=24` | 🔑 | Thống kê theo giờ |
| POST | `/admin/reset_flow` | 🔑 | Reset flow data |

### Simulate
| Method | Path | Auth | Mô tả |
|--------|------|------|-------|
| POST | `/simulate/start` | 🔑 | Bắt đầu simulation |
| POST | `/simulate/stop` | 🔑 | Dừng simulation |
| GET | `/simulate/status` | 🔑 | Trạng thái + positions |

### Engine
| Method | Path | Auth | Mô tả |
|--------|------|------|-------|
| POST | `/engine/solve` | 🔑 | Chạy Dijkstra |
| POST | `/engine/update_cost` | 🔑 | Cập nhật cost |
| GET | `/engine/convergence` | 🔑 | Trạng thái hội tụ |
| POST | `/engine/set_params` | 🔑 | Thiết lập tham số |
| GET | `/engine/health` | 🔑 | Health check |
| POST | `/engine/clear_cache` | 🔑 | Xóa cache |
| POST | `/engine/load_mapf` | 🔑 | Load MAPF output |
| GET | `/engine/mapf_positions?timestep=` | 🔑 | Vị trí agents |
| GET | `/engine/mapf_info` | 🔑 | MAPF metadata |

### Medical
| Method | Path | Auth | Mô tả |
|--------|------|------|-------|
| GET | `/medical/get_tasks` | ✅ | Danh sách chỉ định |
| GET | `/medical/get_queue?poi_id=` | ✅ | Hàng đợi phòng |
| GET | `/medical/result_status?treatment_id=` | ✅ | Kết quả xét nghiệm |
| GET | `/medical/get_prescription` | ✅ | Đơn thuốc |
| GET | `/medical/room_open?poi_id=` | ✅ | Giờ mở cửa |
| GET | `/medical/get_history` | ✅ | Lịch sử khám |
| POST | `/medical/sync_now` | ✅ | Đồng bộ HIS |

### Device
| Method | Path | Auth | Mô tả |
|--------|------|------|-------|
| GET | `/device/stations` | ✅ | Trạm thiết bị |
| GET | `/device/wheelchairs` | ✅ | Xe lăn trống |
| GET | `/device/status/:id` | ✅ | Trạng thái |
| GET | `/device/track/:id` | ✅ | Vị trí |

### SOS
| Method | Path | Auth | Mô tả |
|--------|------|------|-------|
| GET | `/sos/get_list` | ✅ | Danh sách |
| GET | `/sos/get_detail` | ✅ | Chi tiết |
| POST | `/sos/respond` | ✅ | Nhận xử lý |
| POST | `/sos/resolve` | ✅ | Đóng SOS |

### Chat
| Method | Path | Auth | Mô tả |
|--------|------|------|-------|
| GET | `/chat/get_rooms` | ✅ | Phòng chat |
| GET | `/chat/get_messages` | ✅ | Tin nhắn |
| POST | `/chat/send_message` | ✅ | Gửi tin |
| POST | `/chat/close_room` | ✅ | Đóng phòng |
| GET | `/chat/get_unread_count` | ✅ | Chưa đọc |
| WS | `/ws/chat?token=xxx` | WS | WebSocket |

### Notification
| Method | Path | Auth | Mô tả |
|--------|------|------|-------|
| GET | `/notification/get_list` | ✅ | Danh sách |
| POST | `/notification/set_read` | ✅ | Đánh dấu đọc |
| DELETE | `/notification/delete` | ✅ | Xóa |

### Util
| Method | Path | Auth | Mô tả |
|--------|------|------|-------|
| GET | `/util/faq` | ❌ | FAQ |
| GET | `/util/about` | ❌ | About |
| GET | `/util/contact` | ❌ | Contact |
| GET | `/util/feedback_summary` | ❌ | Feedback |

**Legend:** ❌ = Public, ✅ = JWT required, 🔑 = Admin role required
