package middleware

import (
	"bytes"
	"io"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLog chứa thông tin một lượt request/response.
type RequestLog struct {
	ID           int64             `json:"id"`
	Timestamp    string            `json:"timestamp"`
	Method       string            `json:"method"`
	Path         string            `json:"path"`
	Query        string            `json:"query,omitempty"`
	Status       int               `json:"status"`
	Duration     int64             `json:"duration_ms"`
	ClientIP     string            `json:"client_ip"`
	UserAgent    string            `json:"user_agent,omitempty"`
	RequestBody  string            `json:"request_body,omitempty"`
	ResponseBody string            `json:"response_body,omitempty"`
	Headers      map[string]string `json:"headers,omitempty"`
}

// bodyLogWriter wraps gin.ResponseWriter to capture response body.
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// LogStore là ring buffer lưu log in-memory (thread-safe).
var logStore = struct {
	sync.RWMutex
	logs  []RequestLog
	maxID int64
}{
	logs: make([]RequestLog, 0, 500),
}

const maxLogs = 500
const maxBodySize = 4096 // Giới hạn body log để không tốn RAM

// GetLogs trả về tất cả log hiện tại (mới nhất trước).
func GetLogs() []RequestLog {
	logStore.RLock()
	defer logStore.RUnlock()
	result := make([]RequestLog, len(logStore.logs))
	copy(result, logStore.logs)
	return result
}

// ClearLogs xóa toàn bộ log.
func ClearLogs() {
	logStore.Lock()
	defer logStore.Unlock()
	logStore.logs = logStore.logs[:0]
}

// truncate cắt string nếu quá dài.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "...(truncated)"
}

// APILogger là Gin middleware ghi nhận mọi request/response vào ring buffer.
func APILogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Đọc request body (nếu có) rồi tạo lại reader
		var reqBody string
		if c.Request.Body != nil && c.Request.ContentLength > 0 && c.Request.ContentLength < int64(maxBodySize*2) {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				reqBody = truncate(string(bodyBytes), maxBodySize)
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		// Wrap response writer để capture response body
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// Xử lý request
		c.Next()

		duration := time.Since(start).Milliseconds()

		// Thu thập headers quan trọng
		headers := map[string]string{
			"Content-Type": c.GetHeader("Content-Type"),
		}
		if auth := c.GetHeader("Authorization"); auth != "" {
			if len(auth) > 20 {
				headers["Authorization"] = auth[:20] + "..."
			} else {
				headers["Authorization"] = auth
			}
		}

		entry := RequestLog{
			Timestamp:    start.Format(time.RFC3339),
			Method:       c.Request.Method,
			Path:         c.Request.URL.Path,
			Query:        c.Request.URL.RawQuery,
			Status:       c.Writer.Status(),
			Duration:     duration,
			ClientIP:     c.ClientIP(),
			UserAgent:    truncate(c.Request.UserAgent(), 200),
			RequestBody:  reqBody,
			ResponseBody: truncate(blw.body.String(), maxBodySize),
			Headers:      headers,
		}

		logStore.Lock()
		logStore.maxID++
		entry.ID = logStore.maxID
		// Prepend (mới nhất trước)
		logStore.logs = append([]RequestLog{entry}, logStore.logs...)
		if len(logStore.logs) > maxLogs {
			logStore.logs = logStore.logs[:maxLogs]
		}
		logStore.Unlock()
	}
}
