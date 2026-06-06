package sse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	DefaultUpstreamErrorStatus = http.StatusBadGateway
	DefaultUpstreamErrorCode   = "upstream_error"
)

type UpstreamError struct {
	Status  int
	Message string
	Code    string
}

func ParseDeepSeekJSONLine(raw []byte) (map[string]any, bool) {
	line := strings.TrimSpace(string(raw))
	if line == "" || !strings.HasPrefix(line, "{") {
		return nil, false
	}
	chunk := map[string]any{}
	if err := json.Unmarshal([]byte(line), &chunk); err != nil {
		return nil, false
	}
	return chunk, true
}

func ParseDeepSeekBusinessError(chunk map[string]any) (UpstreamError, bool) {
	if chunk == nil {
		return UpstreamError{}, false
	}
	if err, ok := businessErrorFromMap(chunk); ok {
		return err, true
	}
	if data, _ := chunk["data"].(map[string]any); data != nil {
		return businessErrorFromMap(data)
	}
	return UpstreamError{}, false
}

func businessErrorFromMap(m map[string]any) (UpstreamError, bool) {
	bizCode := intValue(m["biz_code"])
	if bizCode == 0 {
		return UpstreamError{}, false
	}
	msg := firstNonEmptyString(m["biz_msg"], m["msg"], m["message"])
	if msg == "" {
		msg = fmt.Sprintf("DeepSeek returned biz_code=%d", bizCode)
	}
	status, code := classifyBusinessError(bizCode, msg)
	return UpstreamError{
		Status:  status,
		Message: fmt.Sprintf("DeepSeek upstream error: biz_code=%d, biz_msg=%s", bizCode, msg),
		Code:    code,
	}, true
}

func classifyBusinessError(_ int, msg string) (int, string) {
	lower := strings.ToLower(strings.TrimSpace(msg))
	switch {
	case strings.Contains(lower, "muted"):
		return http.StatusTooManyRequests, "upstream_account_muted"
	case strings.Contains(lower, "rate") || strings.Contains(lower, "limit"):
		return http.StatusTooManyRequests, "upstream_rate_limited"
	default:
		return DefaultUpstreamErrorStatus, "upstream_business_error"
	}
}

func firstNonEmptyString(values ...any) string {
	for _, v := range values {
		s, _ := v.(string)
		if strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func intValue(v any) int {
	switch x := v.(type) {
	case int:
		return x
	case int64:
		return int(x)
	case float64:
		return int(x)
	case json.Number:
		i, _ := x.Int64()
		return int(i)
	default:
		return 0
	}
}
