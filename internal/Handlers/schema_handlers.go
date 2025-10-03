package handlers

import (
	"io"
	"net/http"

	"github.com/24tylerdurden/levo-api/internal/services"
	"github.com/gin-gonic/gin"
)

type SchemaHandler struct {
	schemaService *services.SchemaService
}

func NewSchemaHandler(service *services.SchemaService) *SchemaHandler {
	return &SchemaHandler{
		schemaService: service,
	}
}

// Upload Schema Service for Application
func (s *SchemaHandler) UploadApplicationSchema(c *gin.Context) {
	appName := c.Param("application")

	file, err := c.FormFile("file")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	// Open the file

	src, err := file.Open()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open the file"})
		return
	}

	defer src.Close()

	// Read the file content
	content, err := io.ReadAll(src)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file contents!"})
		return
	}

	response, err := s.schemaService.UploadSchema(appName, "", content, file.Filename)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusCreated, response)
}

// Upload Schema Service for Services

func (s *SchemaHandler) UploadServiceSchema(c *gin.Context) {
	appName := c.Param("application")

	serviceName := c.Param("service")

	file, err := c.FormFile("file")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required."})
	}

	// Open the file

	src, err := file.Open()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file."})
		return
	}

	defer src.Close()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file."})
		return
	}

	content, err := io.ReadAll(src)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file content"})
		return
	}

	response, err := s.schemaService.UploadSchema(appName, serviceName, content, file.Filename)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusCreated, response)
}

// Get Latest application schema

func (s *SchemaHandler) GetLatestApplicationSchema(c *gin.Context) {
	appName := c.Param("application")

	schema, err := s.schemaService.GetSchema(appName, "", "latest")

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, schema)
}

// Get Application schema version

func (s *SchemaHandler) GetApplicationSchemVersion(c *gin.Context) {
	appName := c.Param("application")
	version := c.Param("version")

	schema, err := s.schemaService.GetSchema(appName, "", version)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, schema)

}

func (s *SchemaHandler) GetLatestServiceSchema(c *gin.Context) {
	appName := c.Param("application")
	serviceName := c.Param("service")

	schema, err := s.schemaService.GetSchema(appName, serviceName, "latest")

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, schema)

}

func (s *SchemaHandler) GetServiceSchemaVersion(c *gin.Context) {
	appName := c.Param("application")
	serviceName := c.Param("service")
	version := c.Param("version")

	schema, err := s.schemaService.GetSchema(appName, serviceName, version)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, schema)
}
