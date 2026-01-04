package middleware

import (
	"log"
	"net/http"
	"time"
)

// JSONLogger 记录请求日志的中间件
func JSONLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// 包装ResponseWriter以捕获状态码
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		// 调用下一个处理器
		next.ServeHTTP(rw, r)
		
		// 记录请求信息
		duration := time.Since(start)
		
		log.Printf(`{"timestamp": "%s", "method": "%s", "path": "%s", "status": %d, "duration_ms": %.2f, "user_agent": "%s", "client_ip": "%s"}`,
			time.Now().Format(time.RFC3339),
			r.Method,
			r.URL.Path,
			rw.statusCode,
			float64(duration.Milliseconds()),
			r.UserAgent(),
			r.RemoteAddr,
		)
	})
}

// responseWriter 包装http.ResponseWriter以捕获状态码
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
