package database

import (
	"time"

	"gorm.io/gorm"
)

// User admin table
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Username  string         `gorm:"uniqueIndex;size:100" json:"username"`
	Password  string         `json:"-"` // bcrypt hashed
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// ConfigItem system settings (JWT secret, etc)
type ConfigItem struct {
	Key   string `gorm:"primaryKey" json:"key"`
	Value string `json:"value"`
}

// Provider represents a downstream LLM provider
type Provider struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Name          string         `gorm:"uniqueIndex;size:100" json:"name"`
	Type          string         `json:"type"` // openai, anthropic, gemini
	BaseURL       string         `json:"base_url"`
	RPM           int            `json:"rpm"`
	IsActive      bool           `gorm:"default:true" json:"is_active"`
	AllowedModels string         `json:"allowed_models"` // comma separated globs
	DeniedModels  string         `json:"denied_models"`  // comma separated globs
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
	APIKeys       []APIKey       `gorm:"foreignKey:ProviderID" json:"api_keys"`
}

// APIKey is a specific key for a provider
type APIKey struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	ProviderID uint           `gorm:"index" json:"provider_id"`
	KeyValue   string         `json:"key_value"`
	Status     string         `gorm:"default:'active'" json:"status"` // active, disabled
	TotalTokens  int64          `gorm:"default:0" json:"total_tokens"`
	RequestCount int64          `gorm:"default:0" json:"request_count"`
	LastUsedAt *time.Time     `json:"last_used_at"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (k *APIKey) BeforeSave(tx *gorm.DB) (err error) {
	if k.KeyValue != "" {
		encrypted, err := Encrypt(k.KeyValue)
		if err != nil {
			return err
		}
		k.KeyValue = encrypted
	}
	return nil
}

func (k *APIKey) AfterFind(tx *gorm.DB) (err error) {
	if k.KeyValue != "" {
		decrypted, err := Decrypt(k.KeyValue)
		if err != nil {
			// If decryption fails, maybe it's not encrypted (legacy)
			return nil
		}
		k.KeyValue = decrypted
	}
	return nil
}

// AccessToken is a client-facing 'sk-xxx' token
type AccessToken struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	TokenHash string         `gorm:"uniqueIndex" json:"-"` // Still keeping hash for compatibility/legacy
	Token     string         `json:"token"`              // Plain text token for retrieval
	Name      string         `json:"name"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	ExpiresAt       *time.Time     `json:"expires_at"`
	AllowedChannels string         `json:"allowed_channels"` // Comma separated IDs
	AllowedModels   string         `json:"allowed_models"`   // Comma separated models
	DeniedModels    string         `json:"denied_models"`    // Comma separated models
	AllowedIPs      string         `json:"allowed_ips"`      // Comma separated IPs/CIDRs
	DeniedIPs       string         `json:"denied_ips"`       // Comma separated IPs/CIDRs
	RPM             *int           `json:"rpm"`              // Requests per minute
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (a *AccessToken) BeforeSave(tx *gorm.DB) (err error) {
	if a.Token != "" {
		encrypted, err := Encrypt(a.Token)
		if err != nil {
			return err
		}
		a.Token = encrypted
	}
	return nil
}

func (a *AccessToken) AfterFind(tx *gorm.DB) (err error) {
	if a.Token != "" {
		decrypted, err := Decrypt(a.Token)
		if err != nil {
			return nil
		}
		a.Token = decrypted
	}
	return nil
}

// Channel represents a simplified LLM upstream (merged Provider + APIKey)
type Channel struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Name          string         `gorm:"size:100" json:"name"`
	Type          string         `json:"type"` // openai, anthropic, gemini
	BaseURL       string         `json:"base_url"`
	RPM           int            `json:"rpm"`
	IsActive      bool           `gorm:"default:true" json:"is_active"`
	AllowedModels string         `json:"allowed_models"` // comma separated
	DeniedModels  string         `json:"denied_models"`  // comma separated
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
	APIKeys       []APIKey       `gorm:"foreignKey:ProviderID" json:"api_keys"`
}

// RequestLog tracks every request
type RequestLog struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	RequestID        string    `json:"request_id"`
	AccessTokenID    uint      `gorm:"index" json:"access_token_id"`
	ChannelID        uint      `gorm:"index" json:"channel_id"`
	Provider         string    `json:"provider"`
	Model            string    `json:"model"`
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	TotalTokens      int       `json:"total_tokens"`
	KeyHint          string    `json:"key_hint"`
	IP               string    `json:"ip_address"` // Fixed to match frontend
	StatusCode       int       `json:"status_code"`
	LatencyMS        int64     `json:"latency_ms"`
	ErrorMessage     string    `json:"error_message"`
	CreatedAt        time.Time `json:"created_at"` // Fixed to match frontend
}
