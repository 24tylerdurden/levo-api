package services

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/24tylerdurden/levo-api/internal/models"
	"gopkg.in/yaml.v3"
)

type SchemaService struct {
	db          *sql.DB
	storagePath string
}

func NewSchemaService(db *sql.DB, storagePath string) *SchemaService {
	return &SchemaService{
		db:          db,
		storagePath: storagePath,
	}
}

func (s *SchemaService) CreateOrGetApplication(name string) (*models.Application, error) {
	var app models.Application
	
	// Try to find existing application
	query := "SELECT id, name, created_at, updated_at FROM applications WHERE name = ?"
	err := s.db.QueryRow(query, name).Scan(&app.ID, &app.Name, &app.CreatedAt, &app.UpdatedAt)
	
	if err == sql.ErrNoRows {
		// Application doesn't exist, create it
		insertQuery := "INSERT INTO applications (name) VALUES (?)"
		result, err := s.db.Exec(insertQuery, name)
		if err != nil {
			return nil, err
		}
		
		id, err := result.LastInsertId()
		if err != nil {
			return nil, err
		}
		
		app.ID = uint(id)
		app.Name = name
		// CreatedAt and UpdatedAt will be set by database defaults
		
		// Fetch the created record to get timestamps
		err = s.db.QueryRow("SELECT created_at, updated_at FROM applications WHERE id = ?", app.ID).Scan(&app.CreatedAt, &app.UpdatedAt)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	
	return &app, nil
}

func (s *SchemaService) CreateOrGetService(appName, serviceName string) (*models.Service, error) {
	// First get the application
	var app models.Application
	appQuery := "SELECT id, name, created_at, updated_at FROM applications WHERE name = ?"
	err := s.db.QueryRow(appQuery, appName).Scan(&app.ID, &app.Name, &app.CreatedAt, &app.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("application not found: %s", appName)
	}
	
	var service models.Service
	
	// Try to find existing service
	query := "SELECT id, name, application_id, created_at FROM services WHERE application_id = ? AND name = ?"
	err = s.db.QueryRow(query, app.ID, serviceName).Scan(&service.ID, &service.Name, &service.ApplicationID, &service.CreatedAt)
	
	if err == sql.ErrNoRows {
		// Service doesn't exist, create it
		insertQuery := "INSERT INTO services (name, application_id) VALUES (?, ?)"
		result, err := s.db.Exec(insertQuery, serviceName, app.ID)
		if err != nil {
			return nil, err
		}
		
		id, err := result.LastInsertId()
		if err != nil {
			return nil, err
		}
		
		service.ID = uint(id)
		service.Name = serviceName
		service.ApplicationID = app.ID
		
		// Fetch the created record to get timestamp
		err = s.db.QueryRow("SELECT created_at FROM services WHERE id = ?", service.ID).Scan(&service.CreatedAt)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	
	return &service, nil
}

// Calculate Next Version Number
func (s *SchemaService) CalculateNextVersion(applicationID uint, serviceID *uint) (string, error) {
	var count int
	
	query := "SELECT COUNT(*) FROM schema_versions WHERE application_id = ?"
	args := []interface{}{applicationID}
	
	if serviceID != nil {
		query += " AND service_id = ?"
		args = append(args, *serviceID)
	} else {
		query += " AND service_id IS NULL"
	}
	
	err := s.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return "", err
	}
	
	return fmt.Sprintf("v%d", count+1), nil
}

func (s *SchemaService) CalculateFileHash(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}

func (s *SchemaService) SaveSchemaFile(content []byte, appName, serviceName, version, fileName string) (string, error) {
	var filePath string
	
	if serviceName == "" {
		// Application-level schema
		appDir := filepath.Join(s.storagePath, "applications", appName)
		if err := os.MkdirAll(appDir, 0755); err != nil {
			return "", err
		}
		
		ext := filepath.Ext(fileName)
		filePath = filepath.Join(appDir, fmt.Sprintf("%s%s", version, ext))
	} else {
		// Service-level Schema
		serviceDir := filepath.Join(s.storagePath, "services", appName, serviceName)
		if err := os.MkdirAll(serviceDir, 0755); err != nil {
			return "", err
		}
		
		ext := filepath.Ext(fileName)
		filePath = filepath.Join(serviceDir, fmt.Sprintf("%s%s", version, ext))
	}
	
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return "", err
	}
	
	return filePath, nil
}

func (s *SchemaService) ValidateOpenAPISpec(content []byte, filename string) error {
	// Check file extension
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".json" && ext != ".yaml" && ext != ".yml" {
		return fmt.Errorf("unsupported file format: %s. Only JSON and YAML are supported", ext)
	}
	
	// Try to parse as JSON first
	var jsonData map[string]interface{}
	if err := json.Unmarshal(content, &jsonData); err == nil {
		// Validate required OpenAPI fields
		if _, hasOpenAPI := jsonData["openapi"]; !hasOpenAPI {
			if _, hasSwagger := jsonData["swagger"]; !hasSwagger {
				return fmt.Errorf("invalid OpenAPI spec: missing 'openapi' or 'swagger' field")
			}
		}
		if _, hasInfo := jsonData["info"]; !hasInfo {
			return fmt.Errorf("invalid OpenAPI spec: missing 'info' field")
		}
		if _, hasPaths := jsonData["paths"]; !hasPaths {
			return fmt.Errorf("invalid OpenAPI spec: missing 'paths' field")
		}
		return nil
	}
	
	// Try to parse as YAML
	var yamlData map[string]interface{}
	if err := yaml.Unmarshal(content, &yamlData); err != nil {
		return fmt.Errorf("file is neither valid JSON nor YAML: %v", err)
	}
	
	if _, hasOpenAPI := yamlData["openapi"]; !hasOpenAPI {
		if _, hasSwagger := yamlData["swagger"]; !hasSwagger {
			return fmt.Errorf("invalid OpenAPI spec: missing 'openapi' or 'swagger' field")
		}
	}
	if _, hasInfo := yamlData["info"]; !hasInfo {
		return fmt.Errorf("invalid OpenAPI spec: missing 'info' field")
	}
	if _, hasPaths := yamlData["paths"]; !hasPaths {
		return fmt.Errorf("invalid OpenAPI spec: missing 'paths' field")
	}
	
	return nil
}

func (s *SchemaService) UploadSchema(appName, serviceName string, fileContent []byte, filename string) (*models.UploadResponse, error) {
	// Validate the OpenAPI spec
	if err := s.ValidateOpenAPISpec(fileContent, filename); err != nil {
		return nil, err
	}
	
	// Get Or Create Application
	app, err := s.CreateOrGetApplication(appName)
	if err != nil {
		return nil, err
	}
	
	var serviceID *uint
	if serviceName != "" {
		service, err := s.CreateOrGetService(appName, serviceName)
		if err != nil {
			return nil, err
		}
		serviceID = &service.ID
	}
	
	// Calculate the next version
	version, err := s.CalculateNextVersion(app.ID, serviceID)
	if err != nil {
		return nil, err
	}
	
	// Calculate the file hash
	fileHash := s.CalculateFileHash(fileContent)
	
	// Save file to storage
	filePath, err := s.SaveSchemaFile(fileContent, appName, serviceName, version, filename)
	if err != nil {
		return nil, err
	}
	
	// Insert schema version record
	insertQuery := "INSERT INTO schema_versions (application_id, service_id, version, file_path, file_hash) VALUES (?, ?, ?, ?, ?)"
	_, err = s.db.Exec(insertQuery, app.ID, serviceID, version, filePath, fileHash)
	if err != nil {
		os.Remove(filePath)
		return nil, err
	}
	
	response := &models.UploadResponse{
		Message:     "Schema Upload Successful",
		Version:     version,
		Application: appName,
		FileHash:    fileHash,
	}
	
	if serviceName != "" {
		response.Service = &serviceName
	}
	
	return response, nil
}

func (s *SchemaService) GetSchema(appName, serviceName string, version string) (*models.SchemaResponse, error) {
	var schema models.SchemaVersion
	
	// Build the query based on parameters
	query := `
		SELECT sv.id, sv.application_id, sv.service_id, sv.version, sv.file_path, sv.file_hash, sv.created_at
		FROM schema_versions sv
		JOIN applications a ON a.id = sv.application_id
		WHERE a.name = ?
	`
	args := []interface{}{appName}
	
	if serviceName != "" {
		query += " AND EXISTS (SELECT 1 FROM services s WHERE s.id = sv.service_id AND s.name = ?)"
		args = append(args, serviceName)
	} else {
		query += " AND sv.service_id IS NULL"
	}
	
	if version == "latest" {
		query += " ORDER BY sv.created_at DESC LIMIT 1"
	} else {
		query += " AND sv.version = ?"
		args = append(args, version)
	}
	
	err := s.db.QueryRow(query, args...).Scan(
		&schema.ID, &schema.ApplicationID, &schema.ServiceID, 
		&schema.Version, &schema.FilePath, &schema.FileHash, &schema.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("schema not found: %v", err)
	}
	
	// Read file content
	content, err := os.ReadFile(schema.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file: %v", err)
	}
	
	// Determine content type
	contentType := "application/json"
	if strings.HasSuffix(schema.FilePath, ".yaml") || strings.HasSuffix(schema.FilePath, ".yml") {
		contentType = "application/x-yml"
	}
	
	response := &models.SchemaResponse{
		Version:     schema.Version,
		Application: appName,
		Content:     string(content),
		ContentType: contentType,
		CreatedAt:   schema.CreatedAt,
	}
	
	if serviceName != "" {
		response.Service = &serviceName
	}
	
	return response, nil
}