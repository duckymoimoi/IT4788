# POI Naming Standard

Tai lieu nay quy dinh cach dat `poi_code`, `poi_name` va `poi_type` khi tao map gia lap. Pham vi hien tai: 1 tang, 1 khu, khong tao POI cho hanh lang, cau thang hoac thang may. Muc tieu la de du lieu nhat quan giua Admin Panel, app benh nhan, routing, medical queue va simulation.

## 1. Nguyen tac chung

- `poi_id` la ID tu dong sinh boi database. Khong tu dat, khong dung `poi_id` trong tai lieu, UI, task mau hoac test case.
- `poi_code` la ma dinh danh nghiep vu cua POI. Truong nay phai duy nhat trong toan bo he thong, khong chi rieng mot map.
- `poi_name` la ten hien thi cho nguoi dung. Ten nay co dau tieng Viet, ngan gon, de tim kiem.
- `poi_type` uu tien dung trong map hien tai: `room`, `entrance`, `wc`, `pharmacy`, `canteen`, `info`, `wifi`, `parking`, `other`.
- Khong tao POI `corridor`, `elevator`, `stairs` trong map gia lap hien tai.
- Khong dat POI trung toa do tren cung mot map.
- Khong dat POI vao o wall/blocked. POI phai nam tren o walkable.
- Anh map chi la nen tinh. Ten phong, marker, tooltip va trang thai nen duoc FE overlay tu API POI.

## 2. Dinh dang `poi_code`

Dung chu in hoa, so va dau gach ngang. Khong dung dau cach, tieng Viet co dau, ky tu dac biet hoac ten qua dai.

Format khuyen nghi cho map hien tai:

```text
<TYPE>-<NUMBER>
```

Trong do:

- `TYPE`: ma loai POI, xem bang ben duoi.
- `NUMBER`: so thu tu 2 chu so hoac so phong thuc te.

Vi du: `ENT-01`, `RM-101`, `WC-01`.

## 3. Bang ma loai POI

| poi_type | Prefix | Khi nao dung | Vi du poi_code |
| --- | --- | --- | --- |
| `room` | `RM` | Phong kham, phong xet nghiem, phong chuc nang | `RM-101`, `RM-102` |
| `entrance` | `ENT` | Cong vao, cua vao sanh, cua phu | `ENT-01`, `ENT-02` |
| `wc` | `WC` | Nha ve sinh | `WC-01`, `WC-02` |
| `pharmacy` | `PH` | Nha thuoc, quay phat thuoc | `PH-01` |
| `canteen` | `CAN` | Canteen, khu an uong | `CAN-01` |
| `info` | `INFO` | Ban huong dan, le tan, quay thong tin | `INFO-01` |
| `parking` | `PKG` | Bai do xe, diem gui xe | `PKG-01` |
| `wifi` | `WIFI` | Diem/khu vuc wifi | `WIFI-01` |
| `other` | `OTH` | Diem khac, chi dung khi khong phu hop cac loai tren | `OTH-01` |

Khong dung trong map hien tai:

| poi_type | Ly do |
| --- | --- |
| `elevator` | Map chi co 1 tang, chua can dieu huong lien tang |
| `stairs` | Map chi co 1 tang, chua can dieu huong lien tang |
| `corridor` | Hanh lang la o walkable tren grid, khong can tao POI rieng |

## 4. Quy uoc `poi_name`

`poi_name` la ten hien thi cho nguoi dung, nen viet tieng Viet co dau va ro nghia.

Format khuyen nghi theo loai:

- `room`: `Phong <chuc nang>` hoac `Phong <so phong> - <chuc nang>`.
- `entrance`: `Cong chinh`, `Cong phu`, `Cua vao sanh chinh`.
- `wc`: `Nha ve sinh`, `WC gan phong 101`.
- `pharmacy`: `Nha thuoc`, `Quay phat thuoc`.
- `canteen`: `Canteen`, `Khu an uong`.
- `info`: `Quay thong tin`, `Le tan sanh chinh`.

Khong nen dat:

- `Phong 1`, `Phong 2` neu khong gan voi so phong/chuc nang ro rang.
- `Test`, `abc`, `poi1`, `node1`.
- Ten qua dai, vi se kho hien thi tren mobile.

## 5. Quy uoc cho map 1 tang, 1 khu

Vi hien tai chi lam 1 tang va 1 khu, khong dua tang/khu vao `poi_code`. Dung ma ngan:

```text
RM-101
RM-102
ENT-01
WC-01
```

Neu sau nay mo rong nhieu tang/khu, tao tai lieu version moi truoc khi doi format. Khong tu y tron format cu va moi trong cung mot bo du lieu.

## 6. Quy uoc cho phong kham va medical task

Phong co lien quan medical queue nen dung `poi_type = room`.

Khuyen nghi:

| poi_code | poi_name | poi_type |
| --- | --- | --- |
| `RM-101` | `Phong kham Noi khoa` | `room` |
| `RM-102` | `Phong kham Ngoai khoa` | `room` |
| `RM-103` | `Phong Xet nghiem` | `room` |
| `RM-104` | `Phong X-Quang` | `room` |
| `RM-105` | `Phong Sieu am` | `room` |
| `RM-106` | `Phong Dien tim` | `room` |
| `RM-107` | `Phong Tai Mui Hong` | `room` |
| `RM-108` | `Phong Nhi khoa` | `room` |

Medical task nen tham chieu den `poi_id` lay tu backend, nhung khi trao doi trong nhom nen ghi kem `poi_code` de de doc:

```text
Task kham noi khoa -> RM-101
Task xet nghiem mau -> RM-103
Task chup X-Quang -> RM-104
```

## 7. Vi du bo POI toi thieu cho map gia lap

Moi map gia lap nen co it nhat:

- 1 entrance.
- 3 den 5 room.
- 1 info landmark de nguoi dung de dinh huong.
- 1 wc neu map co khong gian dich vu.

Vi du:

| poi_code | poi_name | poi_type | is_landmark | Ghi chu |
| --- | --- | --- | --- | --- |
| `ENT-01` | `Cong chinh` | `entrance` | `true` | Diem bat dau pho bien |
| `INFO-01` | `Quay thong tin` | `info` | `true` | Landmark gan sanh |
| `RM-101` | `Phong kham Noi khoa` | `room` | `false` | Phong medical queue |
| `RM-102` | `Phong kham Ngoai khoa` | `room` | `false` | Phong medical queue |
| `RM-103` | `Phong Xet nghiem` | `room` | `false` | Phong medical queue |
| `RM-104` | `Phong X-Quang` | `room` | `false` | Phong medical queue |
| `PH-01` | `Nha thuoc` | `pharmacy` | `true` | Diem tien ich |
| `WC-01` | `Nha ve sinh` | `wc` | `false` | Tien ich |
| `CAN-01` | `Canteen` | `canteen` | `false` | Tien ich |

## 8. Checklist truoc khi luu map

- `poi_code` dung prefix va khong trung voi POI da co.
- `poi_name` de doc, khong phai ten test.
- `poi_type` dung enum backend.
- POI nam tren o walkable.
- Cac phong medical dung `poi_type = room`.
- Cac diem bat dau/diem moc quan trong co `is_landmark = true`.
- Khong tao `corridor`, `elevator`, `stairs` cho map hien tai.
- Neu map nay se dung cho app benh nhan, map active phai co anh nen `map_image_url`.
