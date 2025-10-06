package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	apiBaseURL  = "http://localhost:8080"
	appName     string
	serviceName string
	specPath    string
)

// Root command
var rootCmd = &cobra.Command{
	Use:   "levo",
	Short: "Levo CLI - Automated pen-testing tool for API-driven applications",
	Long:  `Levo CLI is an automated pen-testing tool that runs in CI/CD pipelines and uncovers sophisticated business logic vulnerabilities present in modern API-driven applications.`,
}

// Import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import OpenAPI specification to Levo platform",
	Long:  `Import an OpenAPI specification file (JSON or YAML) to the Levo platform for an application or service.`,
	RunE:  runImport,
}

// Test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test application or service schemas",
	Long:  `Fetch and display the latest schema for an application or service to verify it's available for testing.`,
	RunE:  runTest,
}

func init() {
	// Import command flags
	importCmd.Flags().StringVarP(&specPath, "spec", "s", "", "Path to the OpenAPI specification file (required)")
	importCmd.Flags().StringVarP(&appName, "application", "a", "", "Application name (required)")
	importCmd.Flags().StringVarP(&serviceName, "service", "S", "", "Service name (optional)")
	importCmd.MarkFlagRequired("spec")
	importCmd.MarkFlagRequired("application")

	// Test command flags
	testCmd.Flags().StringVarP(&appName, "application", "a", "", "Application name (required)")
	testCmd.Flags().StringVarP(&serviceName, "service", "S", "", "Service name (optional)")
	testCmd.MarkFlagRequired("application")

	// Add commands to root
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(testCmd)
}

func Execute() error {
	return rootCmd.Execute()
}

func runImport(cmd *cobra.Command, args []string) error {
	// Validate spec file exists
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		return fmt.Errorf("specification file not found: %s", specPath)
	}

	// Read the spec file
	fileContent, err := os.ReadFile(specPath)
	if err != nil {
		return fmt.Errorf("failed to read specification file: %v", err)
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(specPath))
	if ext != ".json" && ext != ".yaml" && ext != ".yml" {
		return fmt.Errorf("unsupported file format: %s. Only JSON and YAML are supported", ext)
	}

	fmt.Printf("Importing OpenAPI specification for application: %s", appName)
	if serviceName != "" {
		fmt.Printf(", service: %s", serviceName)
	}
	fmt.Printf("\n")

	// Upload the schema
	var uploadURL string
	if serviceName != "" {
		uploadURL = fmt.Sprintf("%s/api/v1/applications/%s/services/%s/schemas", apiBaseURL, appName, serviceName)
	} else {
		uploadURL = fmt.Sprintf("%s/api/v1/applications/%s/schemas", apiBaseURL, appName)
	}

	response, err := uploadFile(uploadURL, fileContent, filepath.Base(specPath))
	if err != nil {
		return fmt.Errorf("failed to upload schema: %v", err)
	}

	// Parse and display response
	var uploadResp struct {
		Message     string  `json:"message"`
		Version     string  `json:"version"`
		Application string  `json:"application"`
		Service     *string `json:"service,omitempty"`
		FileHash    string  `json:"file_hash"`
	}

	if err := json.Unmarshal(response, &uploadResp); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	if uploadResp.Service != nil {
		fmt.Printf("   Service: %s\n", *uploadResp.Service)
	}

	return nil
}

func runTest(cmd *cobra.Command, args []string) error {
	fmt.Printf("Testing schema for application: %s", appName)
	if serviceName != "" {
		fmt.Printf(", service: %s", serviceName)
	}
	fmt.Printf("\n")

	// Fetch the latest schema
	var fetchURL string
	if serviceName != "" {
		fetchURL = fmt.Sprintf("%s/api/v1/applications/%s/services/%s/schemas/latest", apiBaseURL, appName, serviceName)
	} else {
		fetchURL = fmt.Sprintf("%s/api/v1/applications/%s/schemas/latest", apiBaseURL, appName)
	}

	response, err := fetchSchema(fetchURL)
	if err != nil {
		return fmt.Errorf("failed to fetch schema: %v", err)
	}

	// Parse and display response
	var schemaResp struct {
		Version     string  `json:"version"`
		Application string  `json:"application"`
		Service     *string `json:"service,omitempty"`
		Content     string  `json:"content"`
		ContentType string  `json:"content_type"`
		CreatedAt   string  `json:"created_at"`
	}

	if err := json.Unmarshal(response, &schemaResp); err != nil {
		return fmt.Errorf("failed to parse schema response: %v", err)
	}

	if schemaResp.Service != nil {
		fmt.Printf("   Service: %s\n", *schemaResp.Service)
	}

	return nil
}

func uploadFile(url string, fileContent []byte, filename string) ([]byte, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Create form file field
	fileWriter, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}

	// Write file content
	if _, err := fileWriter.Write(fileContent); err != nil {
		return nil, err
	}

	// Close the writer
	writer.Close()

	// Create HTTP request
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func fetchSchema(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
