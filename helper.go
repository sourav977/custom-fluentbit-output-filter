package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"unsafe"

	"github.com/IBM/cloudant-go-sdk/cloudantv1"
	"github.com/IBM/go-sdk-core/core"
	"github.com/fluent/fluent-bit-go/output"
	"github.com/google/uuid"
)

// InitializeConfig validates and initializes required configurations
func InitializeConfig(plugin unsafe.Pointer) error {
	// Read config from Fluent Bit
	fmt.Println("[cloudant_output] In InitializeConfig")
	endpoint = output.FLBPluginConfigKey(plugin, "Endpoint")
	authMode = output.FLBPluginConfigKey(plugin, "Authentication_Mode")
	tokenPath = output.FLBPluginConfigKey(plugin, "CR_Token_Mount_Path")
	cloudantDatabase = output.FLBPluginConfigKey(plugin, "Database")

	// Validate mandatory configurations
	if endpoint == "" {
		return errors.New("missing mandatory config: Endpoint")
	}
	if authMode == "" {
		return errors.New("missing mandatory config: Authentication_Mode")
	} else if authMode != "IAMAPIKEY" {
		return errors.New("invalid Authentication_Mode: Must be IAMAPIKEY")
	}
	if tokenPath == "" {
		return errors.New("missing mandatory config: CR_Token_Mount_Path")
	}
	if cloudantDatabase == "" {
		return errors.New("missing mandatory config: Database name")
	}

	fmt.Println("[cloudant_output] Configurations initialized successfully.")
	return nil
}

// ReadCloudantAPIKey Read API Key based on authentication mode
func ReadCloudantAPIKey() error {
	fmt.Println("[cloudant_output] In ReadCloudantAPIKey")
	switch strings.ToUpper(authMode) {
	case "IAMAPIKEY":
		// Read API key from Kubernetes Secret (mounted file)
		apiKeyBytes, err := os.ReadFile(tokenPath)
		if err != nil {
			return fmt.Errorf("[http_output] ERROR: Failed to read IAMAPIKEY from: %s, error: %v", tokenPath, err)
		}
		apiKey = strings.TrimSpace(string(apiKeyBytes))

	case "ENV":
		// Read API key from environment variable, helpful when running locally
		apiKey = os.Getenv("API_KEY")
		if apiKey == "" {
			return fmt.Errorf("[http_output] ERROR: API_KEY environment variable not set")
		}

	default:
		fmt.Println("[http_output] WARNING: Authentication_Mode not set, proceeding without authentication.")
	}

	return nil
}

// initCloudantClient Initialize Cloudant client
func initCloudantClient() (*cloudantv1.CloudantV1, error) {
	fmt.Println("[cloudant_output] In initCloudantClient")
	var err error
	authenticator := &core.IamAuthenticator{
		ApiKey: apiKey,
	}
	cloudantService, err := cloudantv1.NewCloudantV1(
		&cloudantv1.CloudantV1Options{
			URL: func(endpoint string) string {
				if !strings.HasPrefix(endpoint, "https://") {
					return "https://" + endpoint
				}
				return endpoint
			}(endpoint),
			Authenticator: authenticator,
		},
	)
	if err != nil {
		panic(err)
	}
	return cloudantService, nil
}

// convertToStringKeyMap converts map[interface{}]interface{} to map[string]interface{}
// while preserving the original values.
func convertToStringKeyMap(input interface{}) (interface{}, error) {
	switch v := input.(type) {
	case map[interface{}]interface{}:
		// Convert map[interface{}]interface{} to map[string]interface{}
		output := make(map[string]interface{})
		for key, value := range v {

			strKey, ok := key.(string)
			if !ok {
				return nil, fmt.Errorf("non-string key found: %v", key)
			}
			// Recursively process values
			convertedValue, err := convertToStringKeyMap(value)
			if err != nil {
				return nil, err
			}
			output[strKey] = convertedValue
		}
		return output, nil

	case []interface{}:
		// If it's a slice, recursively convert its elements
		for i, item := range v {
			convertedItem, err := convertToStringKeyMap(item)
			if err != nil {
				return nil, err
			}
			v[i] = convertedItem
		}
		return v, nil

	case []byte:
		// Explicitly handle []byte by converting it to string
		return string(v), nil

	default:
		// Preserve the original type without unnecessary conversion
		return v, nil
	}
}

// sendToCloudant send record to destination Cloudant database
func sendToCloudant(service *cloudantv1.CloudantV1, records []interface{}) int {
	fmt.Println("[cloudant_output] In sendToCloudant")
	for _, record := range records {
		convertedMap, ok := record.(map[string]interface{})
		if !ok {
			fmt.Println("[cloudant_output] ERROR: Failed to convert to map[string]interface{}")
		}
		doc := &cloudantv1.Document{
			ID: core.StringPtr(generateUniqueID()),
		}
		for key, value := range convertedMap {
			doc.SetProperty(key, value)
		}

		// Create PostDocument options
		postDocumentOptions := service.NewPostDocumentOptions(cloudantDatabase)
		postDocumentOptions.SetDocument(doc)

		// Send document to Cloudant
		documentResult, _, err := service.PostDocument(postDocumentOptions)
		if err != nil {
			fmt.Printf("[cloudant_output] ERROR: Failed to send document to Cloudant: %v\nerr: %v\n", documentResult.Error, err)
			return output.FLB_ERROR
		}

		// Debugging: Print Cloudant response
		//getCloudantDocForDebug(service, documentResult)

	}
	fmt.Println("[cloudant_output] Successfully sent all records to Cloudant.")
	return output.FLB_OK
}

// Generate a unique ID using UUID v4
func generateUniqueID() string {
	fmt.Println("[cloudant_output] In generateUniqueID for cloudant doc")
	return uuid.New().String()
}

// getCloudantDocForDebug can be called to verify what has been sent to cloudant recently
func getCloudantDocForDebug(service *cloudantv1.CloudantV1, documentResult *cloudantv1.DocumentResult) {
	fmt.Println("[cloudant_output] In getCloudantDocForDebug to verify what has been sent to cloudant recently")
	// Debugging: Print Cloudant response
	b, _ := json.MarshalIndent(documentResult, "", "  ")
	fmt.Println("Cloudant Response:", string(b))

	getDocumentOptions := service.NewGetDocumentOptions(
		cloudantDatabase,
		*documentResult.ID,
	)
	document, _, err := service.GetDocument(getDocumentOptions)
	if err != nil {
		fmt.Printf("[cloudant_output] ERROR: Failed to retrieve document: %v\n", err)
		panic(err)
	}
	d, _ := json.MarshalIndent(document, "", "  ")
	fmt.Println("retrived doc:", string(d))
}
