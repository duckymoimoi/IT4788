package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response la cau truc JSON chuan tra ve cho tat ca API.
// Moi API deu tra ve dung dinh dang nay, khong ngoai le.
//
// Vi du thanh cong:
//
//	{ "code": 1000, "message": "OK", "data": { "token": "..." } }
//
// Vi du loi:
//
//	{ "code": 3008, "message": "Password incorrect", "data": null }
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// =====================================
// MA RESPONSE - SUCCESS
// =====================================

const CodeOK = 1000

// =====================================
// MA RESPONSE - LOI CHUNG (1xxx / 4000 / 5000)
// =====================================

const (
	CodeInvalidToken  = 1004 // Token khong hop le hoac khong du quyen
	CodeNotAccess     = 1009 // Truy cap bi tu choi (vd: user goi API admin)
	CodeLimitExceeded = 1010 // Vuot qua gioi han nghiep vu
	CodeBadRequest    = 4000 // Yeu cau khong hop le
	CodeInternalError = 5000 // Loi may chu noi bo
)

// =====================================
// MA RESPONSE - VALIDATION (2xxx)
// Loi du lieu dau vao tu phia client
// =====================================

const (
	CodeMissingParam   = 2001 // Thieu parameter bat buoc
	CodeInvalidType    = 2002 // Sai kieu du lieu
	CodeInvalidValue   = 2003 // Gia tri khong hop le
	CodeMethodNotAllow = 2004 // Sai HTTP method
	CodeBodyInvalid    = 2005 // Body sai format JSON
)

// =====================================
// MA RESPONSE - AUTH (3xxx)
// Loi xac thuc nguoi dung
// =====================================

const (
	CodeTokenInvalid      = 3001 // Token sai chu ky hoac bi gia mao
	CodeTokenExpired      = 3002 // Token het han
	CodeNotAuthenticated  = 3003 // Chua dang nhap, khong co token
	CodeOTPIncorrect      = 3004 // Sai ma OTP
	CodeOTPExpired        = 3005 // OTP het han
	CodeUserAlreadyExists = 3006 // So dien thoai da dang ky
	CodeUserNotFound      = 3007 // Khong tim thay nguoi dung
	CodePasswordIncorrect = 3008 // Sai mat khau
)

// =====================================
// MA RESPONSE - AUTHORIZATION (31xx)
// Loi phan quyen
// =====================================

const (
	CodePermissionDenied = 3101 // Khong co quyen truy cap
	CodeAdminRequired    = 3102 // Chi admin moi duoc phep
)

// =====================================
// MA RESPONSE - MAP (4xxx)
// Loi lien quan ban do
// =====================================

const (
	CodeFloorNotFound     = 4001 // Tang khong ton tai
	CodeNodeNotFound      = 4002 // Phong khong ton tai
	CodeEdgeNotFound      = 4003 // Hanh lang khong ton tai
	CodeInvalidCoordinate = 4004 // Toa do sai
)

// =====================================
// MA RESPONSE - ROUTE (5xxx)
// Loi lien quan tim duong
// =====================================

const (
	CodeInvalidStartLocation = 5001 // Diem bat dau sai
	CodeInvalidDestination   = 5002 // Diem dich sai
	CodePathNotFound         = 5003 // Khong tim thay duong di
)

// =====================================
// MA RESPONSE - FLOW (6xxx)
// Loi lien quan luong nguoi / mat do
// =====================================

const (
	CodeInvalidLocationData = 6001 // Toa do ping sai
	CodeDensityUnavailable  = 6002 // Khong co du lieu mat do
)

// =====================================
// MA RESPONSE - MEDICAL (7xxx)
// Loi lien quan he thong HIS
// =====================================

const (
	CodeHISUnavailable = 7001 // HIS service down
	CodeTaskNotFound   = 7002 // Khong co chi dinh
)

// =====================================
// MA RESPONSE - ASSET (8xxx)
// Loi lien quan xe lan / tai san
// =====================================

const (
	CodeAssetNotFound    = 8001 // Xe khong ton tai
	CodeAssetUnavailable = 8002 // Khong co xe trong
)

// =====================================
// MA RESPONSE - ENGINE (9xxx)
// Loi lien quan thuat toan
// =====================================

const (
	CodeEngineUnavailable = 9001 // Server thuat toan down
	CodeEngineTimeout     = 9002 // Qua thoi gian xu ly
)

// =====================================
// MA RESPONSE - SYSTEM (99xx)
// Loi he thong
// =====================================

const (
	CodeDBConnectionFailed = 9901 // Loi ket noi database
	CodeDBQueryFailed      = 9902 // Loi SQL query
	CodeUnexpected         = 9999 // Loi he thong khong xac dinh
)

// =====================================
// HAM HELPER
// =====================================

// Success tra ve HTTP 200 voi code 1000 va data dinh kem.
// Dung khi xu ly thanh cong.
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeOK,
		Message: "OK",
		Data:    data,
	})
}

// Error tra ve HTTP 200 voi ma loi tuong ung va data = null.
// Tat ca loi deu tra HTTP 200, phan biet qua truong code trong body.
// Day la quy uoc cua nhom, giu nhat quan tren toan bo API.
func Error(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// =====================================
// HAM SHORTCUT CHO CAC LOI PHAT SINH NHIEU
// Giup handler viet ngan gon hon
// =====================================

func ErrMissingParam(c *gin.Context) {
	Error(c, CodeMissingParam, "Missing required parameter")
}

func ErrBodyInvalid(c *gin.Context) {
	Error(c, CodeBodyInvalid, "Request body invalid")
}

func ErrTokenInvalid(c *gin.Context) {
	Error(c, CodeTokenInvalid, "Token is invalid")
}

func ErrTokenExpired(c *gin.Context) {
	Error(c, CodeTokenExpired, "Token expired")
}

func ErrNotAuthenticated(c *gin.Context) {
	Error(c, CodeNotAuthenticated, "User not authenticated")
}

func ErrOTPIncorrect(c *gin.Context) {
	Error(c, CodeOTPIncorrect, "OTP incorrect")
}

func ErrOTPExpired(c *gin.Context) {
	Error(c, CodeOTPExpired, "OTP expired")
}

func ErrUserAlreadyExists(c *gin.Context) {
	Error(c, CodeUserAlreadyExists, "User already exists")
}

func ErrUserNotFound(c *gin.Context) {
	Error(c, CodeUserNotFound, "User not found")
}

func ErrPasswordIncorrect(c *gin.Context) {
	Error(c, CodePasswordIncorrect, "Password incorrect")
}

func ErrPermissionDenied(c *gin.Context) {
	Error(c, CodePermissionDenied, "Permission denied")
}

func ErrAdminRequired(c *gin.Context) {
	Error(c, CodeAdminRequired, "Admin role required")
}

func ErrUnexpected(c *gin.Context) {
	Error(c, CodeUnexpected, "Unexpected exception")
}

func ErrBadRequest(c *gin.Context, msg string) {
	if msg == "" {
		msg = "Bad request"
	}
	Error(c, CodeBadRequest, msg)
}

func ErrInternalError(c *gin.Context) {
	Error(c, CodeInternalError, "Internal server error")
}

func ErrInvalidToken(c *gin.Context) {
	Error(c, CodeInvalidToken, "Invalid token")
}

func ErrNotAccess(c *gin.Context) {
	Error(c, CodeNotAccess, "Not access")
}

func ErrLimitExceeded(c *gin.Context) {
	Error(c, CodeLimitExceeded, "Limit exceeded")
}

func ErrInvalidType(c *gin.Context) {
	Error(c, CodeInvalidType, "Invalid parameter type")
}

func ErrInvalidValue(c *gin.Context) {
	Error(c, CodeInvalidValue, "Invalid parameter value")
}

func ErrMethodNotAllowed(c *gin.Context) {
	Error(c, CodeMethodNotAllow, "Method not allowed")
}
