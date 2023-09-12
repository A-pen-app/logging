package logging

type Level uint

const (
	LevelFirst Level = iota
	LevelCritical
	LevelError
	LevelWarn
	LevelInfo
	LevelDebug
	LevelLast
)

type Config struct {
	ProjectID    string `json:"project_id" yaml:"project_id"`
	Level        Level  `json:"level" yaml:"level"`
	Development  bool   `json:"development" yaml:"development"`
	KeyRequestID string `json:"key_request_id" yaml:"key_request_id"`
	KeyUserID    string `json:"key_user_id" yaml:"key_user_id"`
	KeyError     string `json:"key_error" yaml:"key_error"`
	KeyScope     string `json:"key_scope" yaml:"key_scope"`
}
