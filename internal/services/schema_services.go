package services

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/24tylerdurden/levo-api/internal/models"
	"gorm.io/gorm"
)

type SchemaService struct {
	db          *gorm.DB
	storagePath string
}

func NewSchemaService(db *gorm.DB, storagePath string) *SchemaService {
	return &SchemaService{
		db:          db,
		storagePath: storagePath,
	}
}

func (s *SchemaService) CreateOrGetApplication(name string) (*models.Application, error) {
	var app models.Application

	result := s.db.Where("name = ?", name).First(&app)

	if result.Error == gorm.ErrRecordNotFound {
		app := models.Application{Name: name}
		if err := s.db.Create(&app).Error; err != nil {
			return nil, err
		}
	} else if result.Error != nil {
		return nil, result.Error
	}

	return &app, nil
}

func (s *SchemaService) CreateOrGetService(appName, serviceName string) (*models.Service, error) {
	var app models.Application

	if err := s.db.Where("name = ?", appName).First(&app).Error; err != nil {
		return nil, fmt.Errorf("application not found : %s", appName)
	}

	var service models.Service

	result := s.db.Where("application_id = ? AND name = ?", app.ID, serviceName).First(&service)

	if result.Error == gorm.ErrRecordNotFound {
		service := models.Service{
			Name:          serviceName,
			ApplicationID: app.ID,
		}

		if err := s.db.Create(&service).Error; err != nil {
			return nil, err
		}
	} else if result.Error != nil {
		return nil, result.Error
	}

	return &service, nil
}

// Calculate Next Version Number
func (s *SchemaService) CalculateNextVersion(applicationID uint, serviceID *uint) (string, error) {
	var count int64

	if err := s.db.Model(&models.SchemaVersion{}).
		Where("application_id = ? AND service_id IS NOT DISTINCT FROM ?", applicationID, serviceID).
		Count(&count).Error; err != nil {
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
		if err := os.Mkdir(appDir, 0755); err != nil {
			return "", err
		}

		ext := filepath.Ext(fileName)

		filepath.Join(appDir, fmt.Sprintf("%s%s", version, ext))

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

func (s *SchemaService) UploadSchema(appName, serviceName string, fileContent []byte, filename string) (*models.UploadResponse, error) {

	// Validate the OpenApi spec

	if err := s.ValidateOpenAPISpec(fileContent, filename); err != nil {
		return nil, err
	}

	// Get Or Create Application

	app, err := s.CreateOrGetApplication(appName)

	if err != nil {
		return nil, err
	}

	var serviceId *uint

	if serviceName != "" {
		service, err := s.CreateOrGetService(appName, serviceName)

		if err != nil {
			return nil, err
		}

		serviceId = &service.ID
	}

	// calculate the next Version

	version, err := s.CalculateNextVersion(app.ID, serviceId)

	if err != nil {
		return nil, err
	}

	// Calcualte the fileHash
	fileHash := s.CalculateFileHash(fileContent)

	// save file to storage
	filePath, err := s.SaveSchemaFile(fileContent, appName, serviceName, version, filename)

	if err != nil {
		return nil, err
	}

	schemaVersion := models.SchemaVersion{
		ApplicationID: app.ID,
		ServiceID:     serviceId,
		Version:       version,
		FilePath:      filePath,
		FileHash:      fileHash,
	}

	if err := s.db.Create(&schemaVersion).Error; err != nil {
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

	query := s.db.Joins("JOIN applications ON applications.id = schema_versions.application_id").
		Where("applications.name = ?", appName)

	if serviceName != "" {
		query = query.Joins("JOIN services ON services.id = schema_versions.service_id").
			Where("services.name = ?", serviceName)
	} else {
		query = query.Where("schema_versions.service_id IS NULL")
	}

	if version == "latest" {
		query = query.Order("schema_versions.created_at DESC")
	} else {
		query = query.Where("schema_versions.version = ?", version)
	}

	if err := query.First(&schema).Error; err != nil {
		return nil, fmt.Errorf("Schema Not found %v", err)
	}

	// Read File Content

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
