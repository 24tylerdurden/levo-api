package models

import "time"

type Application struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Service struct {
	ID            uint      `json:"id"`
	Name          string    `json:"name"`
	ApplicationID uint      `json:"application_id"`
	CreatedAt     time.Time `json:"created_at"`
}

type SchemaVersion struct {
	ID            uint       `json:"id"`
	ApplicationID uint       `json:"application_id"`
	ServiceID     *uint      `json:"service_id,omitempty"`
	Version       string     `json:"version"`
	FilePath      string     `json:"file_path"`
	FileHash      string     `json:"file_hash"`
	CreatedAt     time.Time  `json:"created_at"`
}

type UploadResponse struct {
	Message     string  `json:"message"`
	Version     string  `json:"version"`
	Application string  `json:"application"`
	Service     *string `json:"service,omitempty"`
	FileHash    string  `json:"file_hash"`
}

type SchemaResponse struct {
	Version     string    `json:"version"`
	Application string    `json:"application"`
	Service     *string   `json:"service,omitempty"`
	Content     string    `json:"content"`
	ContentType string    `json:"content_type"`
	CreatedAt   time.Time `json:"created_at"`
}
