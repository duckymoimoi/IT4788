# Hospital Navigation API — Full Test Report

> **Server:** `https://group3.it4788.sukkaito.id.vn/api`  
> **Time:** 2026-04-29 15:32:06  
> **Accounts:** Admin `0900000001` / Staff `0900000003` / Patient `0900000004` / Pwd `password123`

## Summary: 94 ✅ / 34 ⚠️ / 6 ❌ / 134 total

| Status | Count |
|---|---|
| ✅ OK | 94 |
| ⚠️ Expected Error | 34 |
| ❌ Fail | 6 |
| Total | 134 |

## Quick Reference

| # | Tag | Method | Path | St | Code | Desc |
|---|-----|--------|------|----|------|------|
| 1 | Auth | `POST` | `/auth/login` | ✅ | 1000 | Login admin |
| 2 | Auth | `POST` | `/auth/login` | ✅ | 1000 | Login patient |
| 3 | Auth | `POST` | `/auth/login` | ✅ | 1000 | Login staff |
| 4 | Auth | `POST` | `/auth/login` | ❌ | 3007 | Login sai |
| 5 | Auth | `POST` | `/auth/signup` | ✅ | 1000 | Signup |
| 6 | Auth | `POST` | `/auth/verify_otp` | ❌ | 3004 | Verify OTP (sai) |
| 7 | Auth | `POST` | `/auth/forgot_password` | ✅ | 1000 | Forgot password |
| 8 | Auth | `POST` | `/auth/change_password` | ✅ | 1000 | Change password |
| 9 | Auth | `POST` | `/auth/logout` | ✅ | 1000 | Logout |
| 10 | Auth | `POST` | `/auth/login` | ✅ | 1000 | Re-login patient |
| 11 | User | `GET` | `/user/get_profile` | ✅ | 1000 | Get profile admin |
| 12 | User | `GET` | `/user/get_profile` | ✅ | 1000 | Get profile patient |
| 13 | User | `POST` | `/user/set_profile` | ✅ | 1000 | Set profile |
| 14 | User | `GET` | `/user/get_settings` | ✅ | 1000 | Get settings |
| 15 | User | `POST` | `/user/set_settings` | ✅ | 1000 | Set settings |
| 16 | User | `POST` | `/user/set_devtoken` | ✅ | 1000 | Set device token |
| 17 | System | `GET` | `/sys/check_version?platform=android&app_version=1.0.0` | ✅ | 1000 | Check version |
| 18 | System | `GET` | `/sys/get_voice_key` | ✅ | 1000 | Voice API key |
| 19 | System | `GET` | `/sys/get_voice_files` | ✅ | 1000 | Voice files |
| 20 | Map | `GET` | `/map/get_floors` | ✅ | 1000 | Get floors |
| 21 | Map | `GET` | `/map/get_nodes?map_id=1` | ✅ | 1000 | Get nodes map_id=1 |
| 22 | Map | `GET` | `/map/get_edges?map_id=1` | ⚠️ | 2003 | Get edges map_id=1 |
| 23 | Map | `GET` | `/map/get_edges` | ⚠️ | 2003 | Get edges (missing param) |
| 24 | Map | `GET` | `/map/get_edges?map_id=99999` | ⚠️ | 2003 | Get edges (not found) |
| 25 | Map | `GET` | `/map/get_meta?map_id=1` | ✅ | 1000 | Get meta |
| 26 | Map | `GET` | `/map/get_depts` | ✅ | 1000 | Get depts |
| 27 | Map | `GET` | `/map/get_depts?node_type=room` | ✅ | 1000 | Get depts filter type |
| 28 | Map | `GET` | `/map/search_location?keyword=phong&map_id=1` | ✅ | 1000 | Search location |
| 29 | Map | `GET` | `/map/search_location?keyword=xet_nghiem` | ✅ | 1000 | Search xet nghiem |
| 30 | Map | `GET` | `/map/get_landmarks?map_id=1` | ✅ | 1000 | Get landmarks |
| 31 | Map | `GET` | `/map/sync_full?map_id=1` | ✅ | 1000 | Sync full |
| 32 | Admin-Map | `POST` | `/admin/add_node` | ✅ | 1000 | Add node |
| 33 | Admin-Map | `POST` | `/admin/edit_node` | ✅ | 1000 | Edit node |
| 34 | Admin-Map | `PATCH` | `/admin/set_weight` | ✅ | 1000 | Set weight |
| 35 | Admin-Map | `PATCH` | `/admin/set_capacity` | ✅ | 1000 | Set capacity |
| 36 | Admin-Map | `DELETE` | `/admin/del_node` | ✅ | 1000 | Delete node |
| 37 | Admin-Map | `POST` | `/admin/add_node` | ⚠️ | 2005 | Add node (empty body) |
| 38 | Admin-Map | `POST` | `/admin/add_node` | ⚠️ | 3003 | Add node (no auth) |
| 39 | Route | `GET` | `/route/get_modes` | ✅ | 1000 | Get modes |
| 40 | Route | `POST` | `/route/preview` | ✅ | 1000 | Preview route |
| 41 | Route | `POST` | `/route/order` | ✅ | 1000 | Order route |
| 42 | Route | `GET` | `/route/get_active` | ✅ | 1000 | Get active (no route) |
| 43 | Route | `GET` | `/route/get_history?page=1&limit=5` | ✅ | 1000 | Get history |
| 44 | Route | `POST` | `/route/order` | ⚠️ | 2005 | Order (empty body) |
| 45 | Route | `POST` | `/route/order` | ⚠️ | 5003 | Order start==dest |
| 46 | Flow | `GET` | `/flow/get_density?grid_location=100` | ✅ | 1000 | Density loc=100 |
| 47 | Flow | `GET` | `/flow/get_density` | ⚠️ | 2001 | Density (missing param) |
| 48 | Flow | `GET` | `/flow/get_heatmap` | ✅ | 1000 | Heatmap |
| 49 | Flow | `GET` | `/flow/get_bottlenecks?limit=5` | ✅ | 1000 | Bottlenecks |
| 50 | Flow | `GET` | `/flow/get_forecast?hours=24` | ✅ | 1000 | Forecast |
| 51 | Flow | `GET` | `/flow/get_alerts` | ✅ | 1000 | Alerts |
| 52 | Flow | `GET` | `/flow/edge_status?grid_location=100` | ✅ | 1000 | Edge status |
| 53 | Flow | `GET` | `/flow/edge_status` | ⚠️ | 2001 | Edge status (missing) |
| 54 | Flow | `POST` | `/flow/ping_location` | ✅ | 1000 | Ping location |
| 55 | Flow | `POST` | `/flow/ping_location` | ⚠️ | 2005 | Ping (empty) |
| 56 | Flow | `POST` | `/flow/ping_location` | ⚠️ | 3003 | Ping (no auth) |
| 57 | Flow | `POST` | `/flow/report_obstacle` | ✅ | 1000 | Report obstacle |
| 58 | Flow | `GET` | `/flow/get_obstacles?status=pending&page=1&limit=5` | ✅ | 1000 | Get obstacles |
| 59 | Flow | `POST` | `/flow/set_priority` | ✅ | 1000 | Set priority |
| 60 | Flow | `POST` | `/flow/resolve_obstacle` | ✅ | 1000 | Resolve obstacle |
| 61 | Flow-Admin | `GET` | `/admin/stats_flow?hours=24` | ✅ | 1000 | Stats flow |
| 62 | Flow-Admin | `GET` | `/admin/stats_flow` | ⚠️ | 3003 | Stats flow (no auth) |
| 63 | Simulate | `GET` | `/simulate/status` | ✅ | 1000 | Status |
| 64 | Simulate | `GET` | `/simulate/status` | ❌ | 3102 | Status (patient→rejected) |
| 65 | Simulate | `GET` | `/simulate/status` | ⚠️ | 3003 | Status (no auth) |
| 66 | Medical | `POST` | `/medical/sync_now` | ✅ | 1000 | Sync HIS |
| 67 | Medical | `GET` | `/medical/get_tasks` | ✅ | 1000 | Get tasks (patient) |
| 68 | Medical | `GET` | `/medical/get_tasks` | ✅ | 1000 | Get tasks (admin) |
| 69 | Medical | `GET` | `/medical/get_tasks` | ⚠️ | 3003 | Get tasks (no auth) |
| 70 | Medical | `GET` | `/medical/get_queue?poi_id=3` | ✅ | 1000 | Get queue poi=3 |
| 71 | Medical | `GET` | `/medical/room_open?poi_id=3` | ✅ | 1000 | Room open poi=3 |
| 72 | Medical | `GET` | `/medical/get_queue` | ⚠️ | 3003 | Get queue (missing param) |
| 73 | Medical | `GET` | `/medical/room_open` | ⚠️ | 3003 | Room open (missing param) |
| 74 | Medical | `GET` | `/medical/get_prescription` | ✅ | 1000 | Get prescription |
| 75 | Medical | `GET` | `/medical/get_history` | ✅ | 1000 | Get history |
| 76 | Medical | `GET` | `/medical/get_history` | ⚠️ | 3003 | Get history (no auth) |
| 77 | Medical | `GET` | `/medical/result_status?treatment_id=1` | ✅ | 1000 | Result status tid=1 |
| 78 | Medical | `GET` | `/medical/result_status?treatment_id=99999` | ⚠️ | 4002 | Result status (not found) |
| 79 | Medical | `POST` | `/medical/checkin_room` | ✅ | 1000 | Checkin (invalid tid) |
| 80 | Medical | `POST` | `/medical/checkout_room` | ✅ | 1000 | Checkout (invalid tid) |
| 81 | Medical | `POST` | `/medical/cancel_task` | ✅ | 1000 | Cancel task (invalid) |
| 82 | Device | `GET` | `/device/stations` | ✅ | 1000 | Stations |
| 83 | Device | `GET` | `/device/stations` | ⚠️ | 3003 | Stations (no auth) |
| 84 | Device | `GET` | `/device/wheelchairs` | ✅ | 1000 | Wheelchairs |
| 85 | Device | `GET` | `/device/wheelchairs` | ⚠️ | 3003 | Wheelchairs (no auth) |
| 86 | Device | `GET` | `/device/status/1` | ✅ | 1000 | Status id=1 |
| 87 | Device | `GET` | `/device/status/99999` | ⚠️ | 8001 | Status id=99999 |
| 88 | Device | `GET` | `/device/track/1` | ⚠️ | 8001 | Track id=1 |
| 89 | Device | `POST` | `/device/book` | ❌ | 1010 | Book (invalid) |
| 90 | Device | `POST` | `/device/book` | ⚠️ | 2005 | Book (empty) |
| 91 | Device | `POST` | `/device/release` | ❌ | 4000 | Release (no booking) |
| 92 | Device | `POST` | `/device/report_broken` | ✅ | 1000 | Report broken |
| 93 | Device | `POST` | `/device/request_staff` | ✅ | 1000 | Request staff |
| 94 | Notification | `GET` | `/notification/get_list` | ✅ | 1000 | Get list |
| 95 | Notification | `GET` | `/notification/get_list` | ⚠️ | 3003 | Get list (no auth) |
| 96 | Notification | `POST` | `/notification/set_read` | ⚠️ | 2005 | Set read id=99999 |
| 97 | Notification | `POST` | `/notification/set_read` | ⚠️ | 2005 | Set read (empty) |
| 98 | Notification | `DELETE` | `/notification/delete` | ⚠️ | 2005 | Delete id=99999 |
| 99 | Notification | `DELETE` | `/notification/delete` | ⚠️ | 2005 | Delete (empty) |
| 100 | SOS | `GET` | `/sos/get_list` | ✅ | 1000 | Get list (admin) |
| 101 | SOS | `GET` | `/sos/get_list` | ⚠️ | 3003 | Get list (no auth) |
| 102 | SOS | `POST` | `/sos/create` | ✅ | 1000 | Create SOS |
| 103 | SOS | `GET` | `/sos/get_detail?sos_id=1` | ✅ | 1000 | Get detail |
| 104 | SOS | `POST` | `/sos/respond` | ✅ | 1000 | Respond SOS |
| 105 | SOS | `POST` | `/sos/resolve` | ✅ | 1000 | Resolve SOS |
| 106 | Chat | `GET` | `/chat/get_rooms` | ✅ | 1000 | Get rooms (admin) |
| 107 | Chat | `GET` | `/chat/get_rooms` | ✅ | 1000 | Get rooms (patient) |
| 108 | Chat | `POST` | `/chat/create_room` | ⚠️ | 2005 | Create room |
| 109 | Chat | `GET` | `/chat/get_unread_count` | ⚠️ | 2001 | Unread count |
| 110 | Util | `GET` | `/util/faq` | ✅ | 1000 | FAQ |
| 111 | Util | `GET` | `/util/about` | ✅ | 1000 | About |
| 112 | Util | `GET` | `/util/contact` | ✅ | 1000 | Contact |
| 113 | Util | `GET` | `/util/feedback_summary` | ✅ | 1000 | Feedback summary |
| 114 | Util | `GET` | `/util/languages` | ✅ | 1000 | Languages |
| 115 | Util | `GET` | `/util/pharmacy` | ✅ | 1000 | Pharmacy |
| 116 | Util | `GET` | `/util/canteen` | ✅ | 1000 | Canteen |
| 117 | Util | `GET` | `/util/parking` | ✅ | 1000 | Parking |
| 118 | Util | `GET` | `/util/wifi` | ✅ | 1000 | WiFi |
| 119 | Util | `GET` | `/util/weather` | ✅ | 1000 | Weather |
| 120 | Util | `POST` | `/util/feedback` | ✅ | 1000 | Send feedback |
| 121 | Util | `POST` | `/util/feedback` | ⚠️ | 2005 | Feedback rating=0 |
| 122 | Util | `POST` | `/util/feedback` | ⚠️ | 2005 | Feedback (empty) |
| 123 | Util | `POST` | `/util/feedback` | ⚠️ | 3003 | Feedback (no auth) |
| 124 | Engine | `GET` | `/engine/health` | ✅ | 1000 | Health |
| 125 | Engine | `GET` | `/engine/health` | ❌ | 3102 | Health (patient) |
| 126 | Engine | `GET` | `/engine/convergence` | ✅ | 1000 | Convergence |
| 127 | Engine | `POST` | `/engine/solve` | ✅ | 1000 | Solve Dijkstra |
| 128 | Engine | `POST` | `/engine/update_cost` | ✅ | 1000 | Update cost |
| 129 | Engine | `POST` | `/engine/set_params` | ✅ | 1000 | Set params |
| 130 | Engine | `POST` | `/engine/clear_cache` | ✅ | 1000 | Clear cache |
| 131 | Engine | `POST` | `/engine/load_mapf` | ✅ | 1000 | Load MAPF |
| 132 | Engine | `GET` | `/engine/mapf_positions?timestep=0` | ✅ | 1000 | MAPF positions t=0 |
| 133 | Engine | `GET` | `/engine/mapf_positions?timestep=5` | ✅ | 1000 | MAPF positions t=5 |
| 134 | Engine | `GET` | `/engine/mapf_info` | ✅ | 1000 | MAPF info |


---

## Auth

### ✅ `POST /auth/login`
**Login admin**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "user_id": 1,
    "full_name": "Nguyen Van Admin",
    "phone_number": "0900000001",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJyb2xlIjoiYWRtaW4iLCJleHAiOjE3NzgwNTYyOTksImlhdCI6MTc3NzQ1MTQ5OX0.txuLBKzddvj04QvYgcIQOhro3jHkq-ESgcHpcq7obes",
    "avatar": null,
    "active": 1,
    "role": "admin"
  }
}
```

### ✅ `POST /auth/login`
**Login patient**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "user_id": 4,
    "full_name": "Pham Thi Benh Nhan",
    "phone_number": "0900000004",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo0LCJyb2xlIjoicGF0aWVudCIsImV4cCI6MTc3ODA1NjI5OSwiaWF0IjoxNzc3NDUxNDk5fQ.jo_Acg6eAZM5exWuFKPF0SMuoubppjs0C-T16QTKe0o",
    "avatar": null,
    "active": 1,
    "role": "patient"
  }
}
```

### ✅ `POST /auth/login`
**Login staff**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "user_id": 3,
    "full_name": "Le Van Staff",
    "phone_number": "0900000003",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjozLCJyb2xlIjoic3RhZmYiLCJleHAiOjE3NzgwNTYzMDAsImlhdCI6MTc3NzQ1MTUwMH0.-eRCpNBV2d4OlZz-Npti2qWiIt4trj92DwCOYdCEh64",
    "avatar": null,
    "active": 1,
    "role": "staff"
  }
}
```

### ❌ `POST /auth/login`
**Login sai**

```json
{
  "code": 3007,
  "message": "User not found",
  "data": null
}
```

### ✅ `POST /auth/signup`
**Signup**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "user_id": 21,
    "otp_code": "544851"
  }
}
```

### ❌ `POST /auth/verify_otp`
**Verify OTP (sai)**

```json
{
  "code": 3004,
  "message": "OTP incorrect",
  "data": null
}
```

### ✅ `POST /auth/forgot_password`
**Forgot password**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "otp_code": "758585"
  }
}
```

### ✅ `POST /auth/change_password`
**Change password**

```json
{
  "code": 1000,
  "message": "OK",
  "data": null
}
```

### ✅ `POST /auth/logout`
**Logout**

```json
{
  "code": 1000,
  "message": "OK",
  "data": null
}
```

### ✅ `POST /auth/login`
**Re-login patient**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "user_id": 4,
    "full_name": "Pham Thi Benh Nhan",
    "phone_number": "0900000004",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo0LCJyb2xlIjoicGF0aWVudCIsImV4cCI6MTc3ODA1NjMwMSwiaWF0IjoxNzc3NDUxNTAxfQ.tdB0S2Txi63mfLHeeuKzpuBmQLH8q50it8fSV9w7EfU",
    "avatar": null,
    "active": 1,
    "role": "patient"
  }
}
```


---

## User

### ✅ `GET /user/get_profile`
**Get profile admin**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "user_id": 1,
    "full_name": "Nguyen Van Admin",
    "phone_number": "0900000001",
    "dob": null,
    "gender": 1,
    "avatar": null
  }
}
```

### ✅ `GET /user/get_profile`
**Get profile patient**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "user_id": 4,
    "full_name": "Pham Thi Benh Nhan",
    "phone_number": "0900000004",
    "dob": "1990-01-15",
    "gender": 0,
    "avatar": null
  }
}
```

### ✅ `POST /user/set_profile`
**Set profile**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "user_id": 4,
    "full_name": "Pham Thi Benh Nhan",
    "phone_number": "0900000004",
    "dob": "1990-01-15",
    "gender": 0,
    "avatar": null
  }
}
```

### ✅ `GET /user/get_settings`
**Get settings**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "language": "vi",
    "theme": "light",
    "notification": true
  }
}
```

### ✅ `POST /user/set_settings`
**Set settings**

```json
{
  "code": 1000,
  "message": "OK",
  "data": null
}
```

### ✅ `POST /user/set_devtoken`
**Set device token**

```json
{
  "code": 1000,
  "message": "OK",
  "data": null
}
```


---

## System

### ✅ `GET /sys/check_version?platform=android&app_version=1.0.0`
**Check version**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "latest_version": "1.0.0",
    "force_update": false,
    "download_url": "https://play.google.com/store/apps/details?id=com.hospital"
  }
}
```

### ✅ `GET /sys/get_voice_key`
**Voice API key**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "api_key": "DEMO_KEY_FOR_DEVELOPMENT",
    "enabled": true,
    "language": "vi-VN",
    "provider": "google"
  }
}
```

### ✅ `GET /sys/get_voice_files`
**Voice files**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "base_url": "https://group3.it4788.sukkaito.id.vn",
    "files": [
      {
        "key": "turn_left",
        "text": "Rẽ trái",
        "url": "https://group3.it4788.sukkaito.id.vn/audio/turn_left.mp3"
      },
      {
        "key": "turn_right",
        "text": "Rẽ phải",
        "url": "https://group3.it4788.sukkaito.id.vn/audio/turn_right.mp3"
      },
      {
        "key": "go_straight",
        "text": "Đi thẳng",
        "url": "https://group3.it4788.sukkaito.id.vn/audio/go_straight.mp3"
      },
      "...(8 total)"
    ],
    "language": "vi"
  }
}
```


---

## Map

### ✅ `GET /map/get_floors`
**Get floors**

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

### ✅ `GET /map/get_nodes?map_id=1`
**Get nodes map_id=1**

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
      "poi_name": "Phong Kham Da Khoa",
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
    "...(13 total)"
  ]
}
```

### ⚠️ `GET /map/get_edges?map_id=1`
**Get edges map_id=1**

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

### ⚠️ `GET /map/get_edges`
**Get edges (missing param)**

```json
{
  "code": 2003,
  "message": "OK",
  "data": {
    "map_id": 0,
    "message": "edges are auto-computed from grid adjacency"
  }
}
```

### ⚠️ `GET /map/get_edges?map_id=99999`
**Get edges (not found)**

```json
{
  "code": 2003,
  "message": "OK",
  "data": {
    "map_id": 99999,
    "message": "edges are auto-computed from grid adjacency"
  }
}
```

### ✅ `GET /map/get_meta?map_id=1`
**Get meta**

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

### ✅ `GET /map/get_depts`
**Get depts**

```json
{
  "code": 1000,
  "message": "OK",
  "data": [
    {
      "ward_id": 2,
      "ward_name": "Khoa Chan Doan Hinh Anh",
      "poi_count": 0
    },
    {
      "ward_id": 4,
      "ward_name": "Khoa Ngoai",
      "poi_count": 0
    },
    {
      "ward_id": 3,
      "ward_name": "Khoa Noi",
      "poi_count": 0
    },
    "...(5 total)"
  ]
}
```

### ✅ `GET /map/get_depts?node_type=room`
**Get depts filter type**

```json
{
  "code": 1000,
  "message": "OK",
  "data": [
    {
      "poi_id": 7,
      "map_id": 1,
      "ward_id": null,
      "poi_code": "RM-105",
      "poi_name": "Phòng Siêu âm",
      "poi_type": "room",
      "grid_row": 4,
      "grid_col": 24,
      "grid_location": 252,
      "is_landmark": false,
      "is_accessible": true,
      "wheelchair_accessible": false,
      "custom_weight": 1,
      "capacity": null,
      "details": null,
      "open_hours": null
    },
    {
      "poi_id": 6,
      "map_id": 1,
      "ward_id": null,
      "poi_code": "RM-104",
      "poi_name": "Phòng X-Quang",
      "poi_type": "room",
      "grid_row": 4,
      "grid_col": 20,
      "grid_location": 248,
      "is_landmark": false,
      "is_accessible": true,
      "wheelchair_accessible": false,
      "custom_weight": 1,
      "capacity": null,
      "details": null,
      "open_hours": null
    },
    {
      "poi_id": 5,
      "map_id": 1,
      "ward_id": null,
      "poi_code": "RM-103",
      "poi_name": "Phòng Xét nghiệm",
      "poi_type": "room",
      "grid_row": 4,
      "grid_col": 16,
      "grid_location": 244,
      "is_landmark": false,
      "is_accessible": true,
      "wheelchair_accessible": false,
      "custom_weight": 1,
      "capacity": null,
      "details": null,
      "open_hours": null
    },
    "...(5 total)"
  ]
}
```

### ✅ `GET /map/search_location?keyword=phong&map_id=1`
**Search location**

```json
{
  "code": 1000,
  "message": "OK",
  "data": []
}
```

### ✅ `GET /map/search_location?keyword=xet_nghiem`
**Search xet nghiem**

```json
{
  "code": 1000,
  "message": "OK",
  "data": []
}
```

### ✅ `GET /map/get_landmarks?map_id=1`
**Get landmarks**

```json
{
  "code": 1000,
  "message": "OK",
  "data": [
    {
      "poi_id": 11,
      "map_id": 1,
      "ward_id": null,
      "poi_code": "INFO-01",
      "poi_name": "Bàn thông tin",
      "poi_type": "info",
      "grid_row": 4,
      "grid_col": 40,
      "grid_location": 268,
      "is_landmark": true,
      "is_accessible": true,
      "wheelchair_accessible": false,
      "custom_weight": 1,
      "capacity": null,
      "details": null,
      "open_hours": null
    },
    {
      "poi_id": 10,
      "map_id": 1,
      "ward_id": null,
      "poi_code": "CAN-01",
      "poi_name": "Canteen Bệnh viện",
      "poi_type": "canteen",
      "grid_row": 4,
      "grid_col": 36,
      "grid_location": 264,
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
    "...(6 total)"
  ]
}
```

### ✅ `GET /map/sync_full?map_id=1`
**Sync full**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "maps": [
      {
        "map_id": 1,
        "map_name": "Hospital Main Floor",
        "rows": 33,
        "cols": 57,
        "map_image_url": null
      }
    ],
    "pois": [
      {
        "poi_id": 1,
        "map_id": 1,
        "ward_id": null,
        "poi_code": "ENT-01",
        "poi_name": "Phong Kham Da Khoa",
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
      "...(13 total)"
    ]
  }
}
```


---

## Admin-Map

### ✅ `POST /admin/add_node`
**Add node**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "poi_id": 15,
    "map_id": 1,
    "ward_id": null,
    "poi_code": "APIDOC_TEST",
    "poi_name": "API Doc Test",
    "poi_type": "room",
    "grid_row": 5,
    "grid_col": 10,
    "grid_location": 295,
    "is_landmark": false,
    "is_accessible": true,
    "wheelchair_accessible": true,
    "custom_weight": 1,
    "capacity": null,
    "details": null,
    "open_hours": null
  }
}
```

### ✅ `POST /admin/edit_node`
**Edit node**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "poi_id": 15,
    "map_id": 1,
    "ward_id": null,
    "poi_code": "APIDOC_TEST",
    "poi_name": "Edited",
    "poi_type": "room",
    "grid_row": 5,
    "grid_col": 10,
    "grid_location": 295,
    "is_landmark": false,
    "is_accessible": true,
    "wheelchair_accessible": true,
    "custom_weight": 1,
    "capacity": null,
    "details": null,
    "open_hours": null
  }
}
```

### ✅ `PATCH /admin/set_weight`
**Set weight**

```json
{
  "code": 1000,
  "message": "OK",
  "data": null
}
```

### ✅ `PATCH /admin/set_capacity`
**Set capacity**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "updated": true
  }
}
```

### ✅ `DELETE /admin/del_node`
**Delete node**

```json
{
  "code": 1000,
  "message": "OK",
  "data": null
}
```

### ⚠️ `POST /admin/add_node`
**Add node (empty body)**

```json
{
  "code": 2005,
  "message": "Request body invalid",
  "data": null
}
```

### ⚠️ `POST /admin/add_node`
**Add node (no auth)**

```json
{
  "code": 3003,
  "message": "User not authenticated",
  "data": null
}
```


---

## Route

### ✅ `GET /route/get_modes`
**Get modes**

```json
{
  "code": 1000,
  "message": "OK",
  "data": [
    {
      "mode_id": "walking",
      "mode_name": "Đi bộ",
      "speed_factor": 1
    },
    {
      "mode_id": "wheelchair",
      "mode_name": "Xe lăn",
      "speed_factor": 0.7
    },
    {
      "mode_id": "stretcher",
      "mode_name": "Cáng",
      "speed_factor": 0.5
    },
    "...(4 total)"
  ]
}
```

### ✅ `POST /route/preview`
**Preview route**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "distance": 76,
    "estimated_time": 76,
    "steps": [
      {
        "step_order": 0,
        "grid_row": 4,
        "grid_col": 4,
        "grid_location": 232
      },
      {
        "step_order": 1,
        "grid_row": 4,
        "grid_col": 5,
        "grid_location": 233
      },
      {
        "step_order": 2,
        "grid_row": 4,
        "grid_col": 6,
        "grid_location": 234
      },
      "...(77 total)"
    ],
    "mode_id": "walking",
    "speed_factor": 1
  }
}
```

### ✅ `POST /route/order`
**Order route**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "paths": [
      {
        "path_id": 1,
        "route_id": "fced225d-e05c-4589-b93a-351e90a97b32",
        "step_order": 0,
        "grid_row": 4,
        "grid_col": 4,
        "grid_location": 232,
        "instruction": "Bắt đầu tại vị trí hiện tại",
        "voice_text": "go_straight"
      },
      {
        "path_id": 2,
        "route_id": "fced225d-e05c-4589-b93a-351e90a97b32",
        "step_order": 1,
        "grid_row": 4,
        "grid_col": 5,
        "grid_location": 233,
        "instruction": "Rẽ phải (Đông)",
        "voice_text": "turn_right"
      },
      {
        "path_id": 3,
        "route_id": "fced225d-e05c-4589-b93a-351e90a97b32",
        "step_order": 2,
        "grid_row": 4,
        "grid_col": 6,
        "grid_location": 234,
        "instruction": "Tiếp tục đi thẳng",
        "voice_text": "go_straight"
      },
      "...(77 total)"
    ],
    "route": {
      "route_id": "fced225d-e05c-4589-b93a-351e90a97b32",
      "user_id": 4,
      "mode_id": "walking",
      "start_location": 232,
      "dest_location": 1876,
      "route_mode": "dijkstra",
      "total_distance": 76,
      "estimated_time": 76,
      "status": "active",
      "created_at": "2026-04-29T08:31:44.234906385Z"
    }
  }
}
```

### ✅ `GET /route/get_active`
**Get active (no route)**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "paths": [
      {
        "path_id": 1,
        "route_id": "fced225d-e05c-4589-b93a-351e90a97b32",
        "step_order": 0,
        "grid_row": 4,
        "grid_col": 4,
        "grid_location": 232,
        "instruction": "Bắt đầu tại vị trí hiện tại",
        "voice_text": "go_straight"
      },
      {
        "path_id": 2,
        "route_id": "fced225d-e05c-4589-b93a-351e90a97b32",
        "step_order": 1,
        "grid_row": 4,
        "grid_col": 5,
        "grid_location": 233,
        "instruction": "Rẽ phải (Đông)",
        "voice_text": "turn_right"
      },
      {
        "path_id": 3,
        "route_id": "fced225d-e05c-4589-b93a-351e90a97b32",
        "step_order": 2,
        "grid_row": 4,
        "grid_col": 6,
        "grid_location": 234,
        "instruction": "Tiếp tục đi thẳng",
        "voice_text": "go_straight"
      },
      "...(77 total)"
    ],
    "route": {
      "route_id": "fced225d-e05c-4589-b93a-351e90a97b32",
      "user_id": 4,
      "mode_id": "walking",
      "start_location": 232,
      "dest_location": 1876,
      "route_mode": "dijkstra",
      "total_distance": 76,
      "estimated_time": 76,
      "status": "active",
      "created_at": "2026-04-29T08:31:44.234906Z"
    }
  }
}
```

### ✅ `GET /route/get_history?page=1&limit=5`
**Get history**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "limit": 5,
    "page": 1,
    "routes": [
      {
        "route_id": "fced225d-e05c-4589-b93a-351e90a97b32",
        "user_id": 4,
        "mode_id": "walking",
        "start_location": 232,
        "dest_location": 1876,
        "route_mode": "dijkstra",
        "total_distance": 76,
        "estimated_time": 76,
        "status": "active",
        "created_at": "2026-04-29T08:31:44.234906Z"
      }
    ],
    "total": 1
  }
}
```

### ⚠️ `POST /route/order`
**Order (empty body)**

```json
{
  "code": 2005,
  "message": "Request body invalid",
  "data": null
}
```

### ⚠️ `POST /route/order`
**Order start==dest**

```json
{
  "code": 5003,
  "message": "start and destination cannot be the same",
  "data": null
}
```


---

## Flow

### ✅ `GET /flow/get_density?grid_location=100`
**Density loc=100**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "grid_location": 100,
    "count": 0,
    "window_minutes": 30
  }
}
```

### ⚠️ `GET /flow/get_density`
**Density (missing param)**

```json
{
  "code": 2001,
  "message": "Missing required parameter",
  "data": null
}
```

### ✅ `GET /flow/get_heatmap`
**Heatmap**

```json
{
  "code": 1000,
  "message": "OK",
  "data": [
    {
      "grid_location": 200,
      "density": 2
    }
  ]
}
```

### ✅ `GET /flow/get_bottlenecks?limit=5`
**Bottlenecks**

```json
{
  "code": 1000,
  "message": "OK",
  "data": [
    {
      "grid_location": 200,
      "count": 2
    }
  ]
}
```

### ✅ `GET /flow/get_forecast?hours=24`
**Forecast**

```json
{
  "code": 1000,
  "message": "OK",
  "data": [
    {
      "hour": 7,
      "count": 10
    },
    {
      "hour": 8,
      "count": 2
    }
  ]
}
```

### ✅ `GET /flow/get_alerts`
**Alerts**

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

### ✅ `GET /flow/edge_status?grid_location=100`
**Edge status**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "grid_location": 100,
    "count": 0,
    "window_minutes": 30
  }
}
```

### ⚠️ `GET /flow/edge_status`
**Edge status (missing)**

```json
{
  "code": 2001,
  "message": "Missing required parameter",
  "data": null
}
```

### ✅ `POST /flow/ping_location`
**Ping location**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "pinged": true
  }
}
```

### ⚠️ `POST /flow/ping_location`
**Ping (empty)**

```json
{
  "code": 2005,
  "message": "Request body invalid",
  "data": null
}
```

### ⚠️ `POST /flow/ping_location`
**Ping (no auth)**

```json
{
  "code": 3003,
  "message": "User not authenticated",
  "data": null
}
```

### ✅ `POST /flow/report_obstacle`
**Report obstacle**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "report_id": 4,
    "user_id": 4,
    "route_id": null,
    "grid_location": 300,
    "report_type": "equipment",
    "description": "test",
    "ResolvedBy": null,
    "status": "pending",
    "created_at": "2026-04-29T08:31:45.322545655Z",
    "ResolvedAt": null,
    "User": null,
    "Resolver": null
  }
}
```

### ✅ `GET /flow/get_obstacles?status=pending&page=1&limit=5`
**Get obstacles**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "limit": 5,
    "page": 1,
    "reports": [
      {
        "report_id": 4,
        "user_id": 4,
        "route_id": null,
        "grid_location": 300,
        "report_type": "equipment",
        "description": "test",
        "ResolvedBy": null,
        "status": "pending",
        "created_at": "2026-04-29T08:31:45.322545Z",
        "ResolvedAt": null,
        "User": null,
        "Resolver": null
      },
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
      }
    ],
    "total": 3
  }
}
```

### ✅ `POST /flow/set_priority`
**Set priority**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "priority_id": 3,
    "emergency_id": null,
    "set_by": 1,
    "from_location": 100,
    "to_location": 101,
    "reason": "emergency",
    "status": "active",
    "activated_at": "2026-04-29T08:31:45.452209558Z",
    "ExpiredAt": null,
    "Staff": null
  }
}
```

### ✅ `POST /flow/resolve_obstacle`
**Resolve obstacle**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "resolved": true
  }
}
```


---

## Flow-Admin

### ✅ `GET /admin/stats_flow?hours=24`
**Stats flow**

```json
{
  "code": 1000,
  "message": "OK",
  "data": [
    {
      "hour": 7,
      "count": 10
    },
    {
      "hour": 8,
      "count": 3
    }
  ]
}
```

### ⚠️ `GET /admin/stats_flow`
**Stats flow (no auth)**

```json
{
  "code": 3003,
  "message": "User not authenticated",
  "data": null
}
```


---

## Simulate

### ✅ `GET /simulate/status`
**Status**

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

### ❌ `GET /simulate/status`
**Status (patient→rejected)**

```json
{
  "code": 3102,
  "message": "Admin role required",
  "data": null
}
```

### ⚠️ `GET /simulate/status`
**Status (no auth)**

```json
{
  "code": 3003,
  "message": "User not authenticated",
  "data": null
}
```


---

## Medical

### ✅ `POST /medical/sync_now`
**Sync HIS**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "synced": true
  }
}
```

### ✅ `GET /medical/get_tasks`
**Get tasks (patient)**

```json
{
  "code": 1000,
  "message": "OK",
  "data": [
    {
      "treatment_id": 1,
      "user_id": 4,
      "poi_id": 3,
      "ward_id": 3,
      "task_type": "exam",
      "task_name": "Kham noi tong quat",
      "priority": 0,
      "sequence_number": 1,
      "status": "pending",
      "note": "",
      "has_result": false,
      "created_at": "2026-04-29T07:11:53.027953Z",
      "updated_at": "2026-04-29T07:11:53.027953Z",
      "CheckinAt": null,
      "CompletedAt": null,
      "User": null,
      "POI": null,
      "Ward": null
    },
    {
      "treatment_id": 2,
      "user_id": 4,
      "poi_id": 4,
      "ward_id": 4,
      "task_type": "lab",
      "task_name": "Xet nghiem mau",
      "priority": 0,
      "sequence_number": 1,
      "status": "pending",
      "note": "",
      "has_result": false,
      "created_at": "2026-04-29T07:11:53.027953Z",
      "updated_at": "2026-04-29T07:11:53.027953Z",
      "CheckinAt": null,
      "CompletedAt": null,
      "User": null,
      "POI": null,
      "Ward": null
    }
  ]
}
```

### ✅ `GET /medical/get_tasks`
**Get tasks (admin)**

```json
{
  "code": 1000,
  "message": "OK",
  "data": [
    {
      "treatment_id": 12,
      "user_id": 1,
      "poi_id": 5,
      "ward_id": 0,
      "task_type": "",
      "task_name": "Kham lam sang Phòng Xét nghiệm",
      "priority": 5,
      "sequence_number": 1,
      "status": "pending",
      "note": "",
      "has_result": false,
      "created_at": "2026-04-29T08:25:06.51422Z",
      "updated_at": "2026-04-29T08:25:06.51422Z",
      "CheckinAt": null,
      "CompletedAt": null,
      "User": null,
      "POI": null,
      "Ward": null
    },
    {
      "treatment_id": 20,
      "user_id": 1,
      "poi_id": 3,
      "ward_id": 0,
      "task_type": "",
      "task_name": "Kham lam sang Phòng khám Nội khoa",
      "priority": 5,
      "sequence_number": 1,
      "status": "pending",
      "note": "",
      "has_result": false,
      "created_at": "2026-04-29T08:31:48.412131Z",
      "updated_at": "2026-04-29T08:31:48.412131Z",
      "CheckinAt": null,
      "CompletedAt": null,
      "User": null,
      "POI": null,
      "Ward": null
    },
    {
      "treatment_id": 15,
      "user_id": 1,
      "poi_id": 4,
      "ward_id": 0,
      "task_type": "",
      "task_name": "Kham lam sang Phòng khám Ngoại khoa",
      "priority": 5,
      "sequence_number": 1,
      "status": "pending",
      "note": "",
      "has_result": false,
      "created_at": "2026-04-29T08:25:44.995868Z",
      "updated_at": "2026-04-29T08:25:44.995868Z",
      "CheckinAt": null,
      "CompletedAt": null,
      "User": null,
      "POI": null,
      "Ward": null
    },
    "...(11 total)"
  ]
}
```

### ⚠️ `GET /medical/get_tasks`
**Get tasks (no auth)**

```json
{
  "code": 3003,
  "message": "User not authenticated",
  "data": null
}
```

### ✅ `GET /medical/get_queue?poi_id=3`
**Get queue poi=3**

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

### ✅ `GET /medical/room_open?poi_id=3`
**Room open poi=3**

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

### ⚠️ `GET /medical/get_queue`
**Get queue (missing param)**

```json
{
  "code": 3003,
  "message": "User not authenticated",
  "data": null
}
```

### ⚠️ `GET /medical/room_open`
**Room open (missing param)**

```json
{
  "code": 3003,
  "message": "User not authenticated",
  "data": null
}
```

### ✅ `GET /medical/get_prescription`
**Get prescription**

```json
{
  "code": 1000,
  "message": "OK",
  "data": [
    {
      "prescription_id": 1,
      "user_id": 4,
      "issued_by": 2,
      "PharmacyPoiID": null,
      "items_json": "[{\"name\":\"Paracetamol 500mg\",\"dosage\":\"2 vien/lan, 3 lan/ngay\",\"qty\":30,\"note\":\"Uong sau an\"},{\"name\":\"Vitamin C 1000mg\",\"dosage\":\"1 vien/ngay\",\"qty\":30,\"note\":\"Uong buoi sang\"}]",
      "status": "pending",
      "issued_at": "2026-04-29T07:11:53.031674Z",
      "DispensedAt": null,
      "User": null,
      "Issuer": null,
      "Pharmacy": null
    }
  ]
}
```

### ✅ `GET /medical/get_history`
**Get history**

```json
{
  "code": 1000,
  "message": "OK",
  "data": [
    {
      "treatment_id": 3,
      "user_id": 4,
      "poi_id": 3,
      "ward_id": 3,
      "task_type": "exam",
      "task_name": "Do huyet ap",
      "priority": 0,
      "sequence_number": 1,
      "status": "completed",
      "note": "",
      "has_result": false,
      "created_at": "2026-04-29T07:11:53.027953Z",
      "updated_at": "2026-04-29T07:11:53.027953Z",
      "CheckinAt": null,
      "CompletedAt": null,
      "User": null,
      "POI": null,
      "Ward": null
    }
  ]
}
```

### ⚠️ `GET /medical/get_history`
**Get history (no auth)**

```json
{
  "code": 3003,
  "message": "User not authenticated",
  "data": null
}
```

### ✅ `GET /medical/result_status?treatment_id=1`
**Result status tid=1**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "has_result": false,
    "status": "pending",
    "task_name": "Kham noi tong quat",
    "treatment_id": 1
  }
}
```

### ⚠️ `GET /medical/result_status?treatment_id=99999`
**Result status (not found)**

```json
{
  "code": 4002,
  "message": "Not found",
  "data": null
}
```

### ✅ `POST /medical/checkin_room`
**Checkin (invalid tid)**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "checkin": true
  }
}
```

### ✅ `POST /medical/checkout_room`
**Checkout (invalid tid)**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "checkout": true
  }
}
```

### ✅ `POST /medical/cancel_task`
**Cancel task (invalid)**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "cancelled": true
  }
}
```


---

## Device

### ✅ `GET /device/stations`
**Stations**

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
        "poi_name": "Phong Kham Da Khoa",
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
    {
      "station_id": 2,
      "poi_id": 2,
      "station_name": "Trạm Cấp Cứu - TNGT",
      "capacity": 8,
      "is_active": true,
      "POI": {
        "poi_id": 2,
        "map_id": 1,
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
        "is_active": true
      },
      "Devices": null
    },
    {
      "station_id": 3,
      "poi_id": 3,
      "station_name": "Trạm Khoa Nội - Tầng 2",
      "capacity": 10,
      "is_active": true,
      "POI": {
        "poi_id": 3,
        "map_id": 1,
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
        "is_active": true
      },
      "Devices": null
    },
    "...(4 total)"
  ]
}
```

### ⚠️ `GET /device/stations`
**Stations (no auth)**

```json
{
  "code": 3003,
  "message": "User not authenticated",
  "data": null
}
```

### ✅ `GET /device/wheelchairs`
**Wheelchairs**

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
    {
      "device_id": 2,
      "device_code": "WL-002",
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
    {
      "device_id": 3,
      "device_code": "WL-003",
      "device_type": "wheelchair",
      "StationID": 1,
      "CurrentPoiID": null,
      "status": "available",
      "BatteryLevel": 85,
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
    "...(10 total)"
  ]
}
```

### ⚠️ `GET /device/wheelchairs`
**Wheelchairs (no auth)**

```json
{
  "code": 3003,
  "message": "User not authenticated",
  "data": null
}
```

### ✅ `GET /device/status/1`
**Status id=1**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "device_id": 1,
    "device_code": "WL-001",
    "device_type": "wheelchair",
    "StationID": 1,
    "CurrentPoiID": null,
    "status": "available",
    "BatteryLevel": 100,
    "is_active": true,
    "Station": null,
    "CurrentPOI": null
  }
}
```

### ⚠️ `GET /device/status/99999`
**Status id=99999**

```json
{
  "code": 8001,
  "message": "Không tìm thấy thiết bị",
  "data": null
}
```

### ⚠️ `GET /device/track/1`
**Track id=1**

```json
{
  "code": 8001,
  "message": "khong xac dinh duoc vi tri thiet bi luc nay",
  "data": null
}
```

### ❌ `POST /device/book`
**Book (invalid)**

```json
{
  "code": 1010,
  "message": "thiet bi khong ton tai",
  "data": null
}
```

### ⚠️ `POST /device/book`
**Book (empty)**

```json
{
  "code": 2005,
  "message": "Request body invalid",
  "data": null
}
```

### ❌ `POST /device/release`
**Release (no booking)**

```json
{
  "code": 4000,
  "message": "khong tim thay thiet bi dang muon",
  "data": null
}
```

### ✅ `POST /device/report_broken`
**Report broken**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "message": "Đã ghi nhận thiết bị hỏng"
  }
}
```

### ✅ `POST /device/request_staff`
**Request staff**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "message": "Đã gửi yêu cầu nhân viên hỗ trợ"
  }
}
```


---

## Notification

### ✅ `GET /notification/get_list`
**Get list**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "limit": 20,
    "notifications": [
      {
        "notif_id": 3,
        "user_id": 4,
        "title": "Nhắc nhở uống thuốc",
        "content": "Đã đến giờ uống Paracetamol 500mg (2 viên). Uống sau ăn.",
        "notif_type": "medicine",
        "is_read": true,
        "ExpiresAt": null,
        "created_at": "2026-04-29T06:41:53.035806Z",
        "ReadAt": null,
        "User": null
      },
      {
        "notif_id": 2,
        "user_id": 4,
        "title": "Kết quả xét nghiệm",
        "content": "Kết quả xét nghiệm máu của bạn đã có. Vui lòng liên hệ bác sĩ để nhận kết quả.",
        "notif_type": "result",
        "is_read": false,
        "ExpiresAt": null,
        "created_at": "2026-04-29T06:11:53.035806Z",
        "ReadAt": null,
        "User": null
      },
      {
        "notif_id": 1,
        "user_id": 4,
        "title": "Lịch khám hôm nay",
        "content": "Bạn có lịch khám Nội tổng quát lúc 9:00 sáng tại Phòng 101",
        "notif_type": "reminder",
        "is_read": false,
        "ExpiresAt": null,
        "created_at": "2026-04-29T05:11:53.035806Z",
        "ReadAt": null,
        "User": null
      },
      "...(4 total)"
    ],
    "page": 1,
    "total": 4
  }
}
```

### ⚠️ `GET /notification/get_list`
**Get list (no auth)**

```json
{
  "code": 3003,
  "message": "User not authenticated",
  "data": null
}
```

### ⚠️ `POST /notification/set_read`
**Set read id=99999**

```json
{
  "code": 2005,
  "message": "Request body invalid",
  "data": null
}
```

### ⚠️ `POST /notification/set_read`
**Set read (empty)**

```json
{
  "code": 2005,
  "message": "Request body invalid",
  "data": null
}
```

### ⚠️ `DELETE /notification/delete`
**Delete id=99999**

```json
{
  "code": 2005,
  "message": "Request body invalid",
  "data": null
}
```

### ⚠️ `DELETE /notification/delete`
**Delete (empty)**

```json
{
  "code": 2005,
  "message": "Request body invalid",
  "data": null
}
```


---

## SOS

### ✅ `GET /sos/get_list`
**Get list (admin)**

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

### ⚠️ `GET /sos/get_list`
**Get list (no auth)**

```json
{
  "code": 3003,
  "message": "User not authenticated",
  "data": null
}
```

### ✅ `POST /sos/create`
**Create SOS**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "sos_id": 1,
    "user_id": 4,
    "grid_location": 100,
    "PosX": 0,
    "PosY": 0,
    "note": "",
    "status": "received",
    "AssignedStaff": null,
    "created_at": "2026-04-29T08:31:53.954578542Z",
    "ResolvedAt": null,
    "User": null,
    "Staff": null
  }
}
```

### ✅ `GET /sos/get_detail?sos_id=1`
**Get detail**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "sos_id": 1,
    "user_id": 4,
    "grid_location": 100,
    "PosX": 0,
    "PosY": 0,
    "note": "",
    "status": "received",
    "AssignedStaff": null,
    "created_at": "2026-04-29T08:31:53.954578Z",
    "ResolvedAt": null,
    "User": null,
    "Staff": null
  }
}
```

### ✅ `POST /sos/respond`
**Respond SOS**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "responded": true
  }
}
```

### ✅ `POST /sos/resolve`
**Resolve SOS**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "resolved": true
  }
}
```


---

## Chat

### ✅ `GET /chat/get_rooms`
**Get rooms (admin)**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "rooms": []
  }
}
```

### ✅ `GET /chat/get_rooms`
**Get rooms (patient)**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "rooms": []
  }
}
```

### ⚠️ `POST /chat/create_room`
**Create room**

```json
{
  "code": 2005,
  "message": "Request body invalid",
  "data": null
}
```

### ⚠️ `GET /chat/get_unread_count`
**Unread count**

```json
{
  "code": 2001,
  "message": "Missing required parameter",
  "data": null
}
```


---

## Util

### ✅ `GET /util/faq`
**FAQ**

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
    "...(10 total)"
  ]
}
```

### ✅ `GET /util/about`
**About**

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

### ✅ `GET /util/contact`
**Contact**

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

### ✅ `GET /util/feedback_summary`
**Feedback summary**

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

### ✅ `GET /util/languages`
**Languages**

```json
{
  "code": 1000,
  "message": "OK",
  "data": [
    {
      "code": "vi",
      "name": "Tiếng Việt"
    },
    {
      "code": "en",
      "name": "English"
    }
  ]
}
```

### ✅ `GET /util/pharmacy`
**Pharmacy**

```json
{
  "code": 1000,
  "message": "OK",
  "data": [
    {
      "poi_id": 8,
      "map_id": 1,
      "poi_code": "PH-01",
      "poi_name": "Nhà thuốc",
      "poi_type": "pharmacy",
      "grid_row": 4,
      "grid_col": 28,
      "grid_location": 256,
      "is_landmark": true,
      "is_accessible": true,
      "wheelchair_accessible": false,
      "custom_weight": 1,
      "is_active": true
    }
  ]
}
```

### ✅ `GET /util/canteen`
**Canteen**

```json
{
  "code": 1000,
  "message": "OK",
  "data": [
    {
      "poi_id": 10,
      "map_id": 1,
      "poi_code": "CAN-01",
      "poi_name": "Canteen Bệnh viện",
      "poi_type": "canteen",
      "grid_row": 4,
      "grid_col": 36,
      "grid_location": 264,
      "is_landmark": true,
      "is_accessible": true,
      "wheelchair_accessible": false,
      "custom_weight": 1,
      "is_active": true
    }
  ]
}
```

### ✅ `GET /util/parking`
**Parking**

```json
{
  "code": 1000,
  "message": "OK",
  "data": []
}
```

### ✅ `GET /util/wifi`
**WiFi**

```json
{
  "code": 1000,
  "message": "OK",
  "data": [
    {
      "poi_id": 12,
      "map_id": 1,
      "poi_code": "WIFI-01",
      "poi_name": "Wifi Lobby",
      "poi_type": "wifi",
      "grid_row": 4,
      "grid_col": 44,
      "grid_location": 272,
      "is_landmark": false,
      "is_accessible": true,
      "wheelchair_accessible": false,
      "custom_weight": 1,
      "is_active": true
    }
  ]
}
```

### ✅ `GET /util/weather`
**Weather**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "city": "Hanoi",
    "raw": null
  }
}
```

### ✅ `POST /util/feedback`
**Send feedback**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "message": "Cảm ơn bạn đã đánh giá!"
  }
}
```

### ⚠️ `POST /util/feedback`
**Feedback rating=0**

```json
{
  "code": 2005,
  "message": "Request body invalid",
  "data": null
}
```

### ⚠️ `POST /util/feedback`
**Feedback (empty)**

```json
{
  "code": 2005,
  "message": "Request body invalid",
  "data": null
}
```

### ⚠️ `POST /util/feedback`
**Feedback (no auth)**

```json
{
  "code": 3003,
  "message": "User not authenticated",
  "data": null
}
```


---

## Engine

### ✅ `GET /engine/health`
**Health**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "status": "ok",
    "db_connected": true,
    "grid_loaded": true,
    "mapf_loaded": false,
    "agent_count": 0
  }
}
```

### ❌ `GET /engine/health`
**Health (patient)**

```json
{
  "code": 3102,
  "message": "Admin role required",
  "data": null
}
```

### ✅ `GET /engine/convergence`
**Convergence**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "iteration": 0,
    "cost": 0,
    "converged": false
  }
}
```

### ✅ `POST /engine/solve`
**Solve Dijkstra**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "distance": 76,
    "estimated_time": 76,
    "steps": [
      {
        "step_order": 0,
        "grid_row": 4,
        "grid_col": 4,
        "grid_location": 232
      },
      {
        "step_order": 1,
        "grid_row": 4,
        "grid_col": 5,
        "grid_location": 233
      },
      {
        "step_order": 2,
        "grid_row": 4,
        "grid_col": 6,
        "grid_location": 234
      },
      "...(77 total)"
    ],
    "mode_id": "walking",
    "speed_factor": 1
  }
}
```

### ✅ `POST /engine/update_cost`
**Update cost**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "poi_id": 1,
    "updated": true,
    "weight": 1
  }
}
```

### ✅ `POST /engine/set_params`
**Set params**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "max_agents": 20,
    "time_step_ms": 500,
    "cost_multiplier": 1
  }
}
```

### ✅ `POST /engine/clear_cache`
**Clear cache**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "cache_cleared": true
  }
}
```

### ✅ `POST /engine/load_mapf`
**Load MAPF**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "loaded": true,
    "makespan": 38,
    "num_task_finished": 31,
    "team_size": 10,
    "total_tasks": 41
  }
}
```

### ✅ `GET /engine/mapf_positions?timestep=0`
**MAPF positions t=0**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "positions": [
      {
        "Row": 14,
        "Col": 55,
        "Location": 0,
        "Orientation": 0
      },
      {
        "Row": 5,
        "Col": 29,
        "Location": 0,
        "Orientation": 0
      },
      {
        "Row": 20,
        "Col": 39,
        "Location": 0,
        "Orientation": 0
      },
      "...(10 total)"
    ],
    "timestep": 0
  }
}
```

### ✅ `GET /engine/mapf_positions?timestep=5`
**MAPF positions t=5**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "positions": [
      {
        "Row": 13,
        "Col": 53,
        "Location": 0,
        "Orientation": 2
      },
      {
        "Row": 5,
        "Col": 33,
        "Location": 0,
        "Orientation": 1
      },
      {
        "Row": 23,
        "Col": 39,
        "Location": 0,
        "Orientation": 2
      },
      "...(10 total)"
    ],
    "timestep": 5
  }
}
```

### ✅ `GET /engine/mapf_info`
**MAPF info**

```json
{
  "code": 1000,
  "message": "OK",
  "data": {
    "loaded": true,
    "makespan": 38,
    "num_task_finished": 31,
    "team_size": 10,
    "total_tasks": 41
  }
}
```
