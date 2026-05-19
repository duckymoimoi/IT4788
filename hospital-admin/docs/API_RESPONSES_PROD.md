# API Responses from Production Server

Generated from `https://group3.it4788.sukkaito.id.vn/api` to provide sample data for the Admin Panel.

> **Auth:** Admin account `0900000001` / `password123`

---

## AUTH

### POST /auth/login

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "user_id": 1,
    "full_name": "Nguyen Van Admin",
    "phone_number": "0900000001",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "avatar": null,
    "active": 1,
    "role": "admin"
  }
}
```

---

## MAP

### GET /map/get_floors

```json
{
  "code": 1000,
  "message": "OK",
  "data": [
    {
      "map_id": 1,
      "map_name": "Hospital Main Floor",
      "rows": 33,
      "cols": 57,
      "map_image_url": null
    }
  ]
}
```

### GET /map/get_meta?map_id=1

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "map_id": 1,
    "map_name": "Hospital Main Floor",
    "rows": 33,
    "cols": 57,
    "map_image_url": null
  }
}
```

### GET /map/get_nodes?map_id=1

```json
{
  "code": 1000,
  "message": "OK",
  "data": [
    {
      "poi_id": 1,
      "map_id": 1,
      "ward_id": null,
      "poi_code": "ENT-01",
      "poi_name": "Cổng chính",
      "poi_type": "entrance",
      "grid_row": 4,
      "grid_col": 4,
      "grid_location": 232,
      "is_landmark": true,
      "is_accessible": true,
      "wheelchair_accessible": false,
      "custom_weight": 1,
      "capacity": null,
      "details": null,
      "open_hours": null
    },
    {
      "poi_id": 2,
      "map_id": 1,
      "ward_id": null,
      "poi_code": "ENT-02",
      "poi_name": "Cổng phụ",
      "poi_type": "entrance",
      "grid_row": 32,
      "grid_col": 52,
      "grid_location": 1876,
      "is_landmark": true,
      "is_accessible": true,
      "wheelchair_accessible": false,
      "custom_weight": 1,
      "capacity": null,
      "details": null,
      "open_hours": null
    },
    {
      "poi_id": 3,
      "map_id": 1,
      "ward_id": null,
      "poi_code": "RM-101",
      "poi_name": "Phòng khám Nội khoa",
      "poi_type": "room",
      "grid_row": 4,
      "grid_col": 8,
      "grid_location": 236,
      "is_landmark": true,
      "is_accessible": true,
      "wheelchair_accessible": false,
      "custom_weight": 1,
      "capacity": null,
      "details": null,
      "open_hours": null
    },
    "... (truncated — 20+ POIs total)"
  ]
}
```

### GET /map/get_edges?map_id=1

> **Note:** Edges tự động tính từ grid adjacency (4 hướng N/S/E/W). Response chỉ xác nhận.

```json
{
  "code": 2003,
  "message": "OK",
  "data": {
    "map_id": 1,
    "message": "edges are auto-computed from grid adjacency"
  }
}
```

### GET /map/get_depts

```json
{
  "code": 1000,
  "message": "OK",
  "data": [
    { "ward_id": 2, "ward_name": "Khoa Chan Doan Hinh Anh", "poi_count": 0 },
    { "ward_id": 4, "ward_name": "Khoa Ngoai", "poi_count": 0 },
    { "ward_id": 3, "ward_name": "Khoa Noi", "poi_count": 0 },
    { "ward_id": 1, "ward_name": "Khoa Xet Nghiem", "poi_count": 0 },
    { "ward_id": 5, "ward_name": "Tien Ich Benh Vien", "poi_count": 0 }
  ]
}
```

### GET /map/search_location?keyword=phong&map_id=1

```json
{
  "code": 1000,
  "message": "OK",
  "data": []
}
```

---

## FLOW

### GET /flow/get_heatmap

```json
{
  "code": 1000,
  "message": "OK",
  "data": [
    { "grid_location": 100, "density": 3 },
    { "grid_location": 300, "density": 2 },
    { "grid_location": 205, "density": 1 },
    "... (truncated)"
  ]
}
```

### GET /flow/get_density?grid_location=232

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "grid_location": 232,
    "count": 0,
    "window_minutes": 30
  }
}
```

### GET /flow/get_bottlenecks?limit=5

```json
{
  "code": 1000,
  "message": "OK",
  "data": []
}
```

### GET /flow/get_alerts

```json
{
  "code": 1000,
  "message": "OK",
  "data": [
    {
      "priority_id": 1,
      "emergency_id": null,
      "set_by": 2,
      "from_location": 100,
      "to_location": 300,
      "reason": "Don duong cap cuu tu tang 1 den phong mo",
      "status": "active",
      "activated_at": "2026-04-29T06:51:53.048935Z",
      "ExpiredAt": null,
      "Staff": null
    }
  ]
}
```

### GET /flow/get_forecast?hours=24

```json
{
  "code": 1000,
  "message": "OK",
  "data": [
    { "hour": 7, "count": 10 }
  ]
}
```

### GET /flow/get_obstacles

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "limit": 20,
    "page": 1,
    "reports": [
      {
        "report_id": 3,
        "user_id": 4,
        "route_id": null,
        "grid_location": 300,
        "report_type": "elevator_broken",
        "description": "Thang may T2 bi hong",
        "ResolvedBy": null,
        "status": "pending",
        "created_at": "2026-04-29T07:06:53.048935Z",
        "ResolvedAt": null,
        "User": null,
        "Resolver": null
      },
      {
        "report_id": 1,
        "user_id": 4,
        "route_id": null,
        "grid_location": 150,
        "report_type": "wet_floor",
        "description": "San uot gan phong kham 101",
        "ResolvedBy": null,
        "status": "pending",
        "created_at": "2026-04-29T07:01:53.048935Z",
        "ResolvedAt": null,
        "User": null,
        "Resolver": null
      },
      {
        "report_id": 2,
        "user_id": 5,
        "route_id": null,
        "grid_location": 200,
        "report_type": "construction",
        "description": "Hanh lang dang sua chua",
        "ResolvedBy": null,
        "status": "resolved",
        "created_at": "2026-04-29T06:11:53.048935Z",
        "ResolvedAt": null,
        "User": null,
        "Resolver": null
      }
    ],
    "total": 3
  }
}
```

---

## ADMIN

### GET /admin/stats_flow?hours=24 🔑

```json
{
  "code": 1000,
  "message": "OK",
  "data": [
    { "hour": 7, "count": 10 }
  ]
}
```

---

## ENGINE

### GET /engine/health 🔑

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "status": "ok",
    "db_connected": true,
    "grid_loaded": false,
    "mapf_loaded": false,
    "agent_count": 0
  }
}
```

---

## SIMULATE

### GET /simulate/status 🔑

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "running": false,
    "team_size": 0,
    "makespan": 0,
    "current_timestep": 0,
    "tick_rate_ms": 0,
    "output_file": ""
  }
}
```

---

## SOS

### GET /sos/get_list ✅

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "limit": 20,
    "page": 1,
    "sos_list": [],
    "total": 0
  }
}
```

---

## MEDICAL

### GET /medical/get_queue?poi_id=3 ✅

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "queue_id": 1,
    "poi_id": 3,
    "current_number": 15,
    "waiting_count": 8,
    "avg_wait_minutes": 20,
    "updated_at": "2026-04-29T07:11:53.017477Z",
    "POI": null
  }
}
```

### GET /medical/get_tasks ✅

```json
{
  "code": 1000,
  "message": "OK",
  "data": []
}
```

### GET /medical/room_open?poi_id=3 ✅

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "close": "17:00",
    "open": "07:00",
    "poi_id": 3,
    "poi_name": "Phòng khám Nội khoa"
  }
}
```

---

## DEVICE

### GET /device/stations ✅

```json
{
  "code": 1000,
  "message": "OK",
  "data": [
    {
      "station_id": 1,
      "poi_id": 1,
      "station_name": "Trạm Sảnh Chính - Tầng 1",
      "capacity": 15,
      "is_active": true,
      "POI": {
        "poi_id": 1,
        "map_id": 1,
        "poi_code": "ENT-01",
        "poi_name": "Cổng chính",
        "poi_type": "entrance",
        "grid_row": 4,
        "grid_col": 4,
        "grid_location": 232,
        "is_landmark": true,
        "is_accessible": true,
        "wheelchair_accessible": false,
        "custom_weight": 1,
        "is_active": true
      },
      "Devices": null
    },
    "... (truncated — 5 stations total)"
  ]
}
```

### GET /device/wheelchairs ✅

```json
{
  "code": 1000,
  "message": "OK",
  "data": [
    {
      "device_id": 1,
      "device_code": "WL-001",
      "device_type": "wheelchair",
      "StationID": 1,
      "CurrentPoiID": null,
      "status": "available",
      "BatteryLevel": 100,
      "is_active": true,
      "Station": {
        "station_id": 1,
        "poi_id": 1,
        "station_name": "Trạm Sảnh Chính - Tầng 1",
        "capacity": 15,
        "is_active": true,
        "POI": null,
        "Devices": null
      },
      "CurrentPOI": null
    },
    "... (truncated — 10 wheelchairs total)"
  ]
}
```

---

## CHAT

### GET /chat/get_rooms ✅

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "rooms": []
  }
}
```

### GET /chat/get_unread_count ✅

> **Note:** Requires `room_id` query parameter.

```json
{
  "code": 2001,
  "message": "Missing required parameter",
  "data": null
}
```

---

## NOTIFICATION

### GET /notification/get_list ✅

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "limit": 20,
    "notifications": [],
    "page": 1,
    "total": 0
  }
}
```

---

## UTIL

### GET /util/faq

```json
{
  "code": 1000,
  "message": "OK",
  "data": [
    {
      "faq_id": 1,
      "category": "Chung",
      "question": "Làm sao để mượn xe lăn?",
      "answer": "Bạn vào mục 'Thiết bị', chọn xe còn trống và nhấn 'Mượn'. Sau đó quét mã QR trên xe để xác nhận.",
      "sort_order": 1,
      "is_active": true
    },
    {
      "faq_id": 2,
      "category": "Khám bệnh",
      "question": "Tôi có thể xem số thứ tự khám ở đâu?",
      "answer": "Vào mục 'Y tế' -> 'Hàng đợi' để theo dõi vị trí hiện tại của mình trong danh sách chờ.",
      "sort_order": 2,
      "is_active": true
    },
    {
      "faq_id": 3,
      "category": "Chỉ đường",
      "question": "Làm thế nào để tìm đường đến phòng khám?",
      "answer": "Nhấn nút 'Tìm đường' trên màn hình chính, nhập tên phòng hoặc mã phòng, hệ thống sẽ vẽ lộ trình đi bộ chi tiết cho bạn.",
      "sort_order": 3,
      "is_active": true
    },
    "... (truncated)"
  ]
}
```

### GET /util/about

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "description": "Hệ thống điều hướng và quản lý bệnh viện thông minh.",
    "hospital_name": "Bệnh viện Đa khoa Trung Tâm",
    "version": "1.0.0"
  }
}
```

### GET /util/contact

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "address": "123 Đường Y Tế, Quận 1, TP.HCM",
    "email": "support@hospital.vn",
    "hotline": "1900-1234"
  }
}
```

### GET /util/feedback_summary

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "average_rating": 0,
    "total_feedbacks": 0
  }
}
```

---

## ERROR RESPONSES (Mẫu)

### 401 Unauthorized (Token hết hạn hoặc thiếu)

```json
{
  "code": 1002,
  "message": "Unauthorized",
  "data": null
}
```

### 2001 Missing Parameter

```json
{
  "code": 2001,
  "message": "Missing required parameter",
  "data": null
}
```

### 2003 Not Supported

```json
{
  "code": 2003,
  "message": "Not supported",
  "data": null
}
```
