package main

import (
	"C"
	"fmt"
	"unsafe"

	"github.com/IBM/cloudant-go-sdk/cloudantv1"
	"github.com/fluent/fluent-bit-go/output"
)

// Global variables
var (
	endpoint         string
	apiKey           string
	authMode         string
	tokenPath        string
	cloudantDatabase string
	cloudantService  *cloudantv1.CloudantV1
)

// FLBPluginRegister Registers the plugin with Fluent Bit.
//
//export FLBPluginRegister
func FLBPluginRegister(def unsafe.Pointer) int {
	fmt.Println("[cloudant_output] In FLBPluginRegister")
	return output.FLBPluginRegister(def, "cloudant_output", "Custom HTTP Output Plugin which writes logs to IBM Cloudant.")
}

// FLBPluginInit Initializes the plugin. Called once when Fluent Bit starts.
//
//export FLBPluginInit
func FLBPluginInit(plugin unsafe.Pointer) int {
	fmt.Println("[cloudant_output] In FLBPluginInit")

	// validate and initialize the config
	if err := InitializeConfig(plugin); err != nil {
		fmt.Println("Configuration error:", err)
		return output.FLB_ERROR
	}

	// Read Cloudant API Key based on authentication mode
	if err := ReadCloudantAPIKey(); err != nil {
		fmt.Println("Failed to read Cloudant API key:", err)
		return output.FLB_ERROR
	}

	// Initialize Cloudant Service
	var err error
	cloudantService, err = initCloudantClient()
	if err != nil {
		fmt.Println("Failed to Initialize Cloudant Service:", err)
		return output.FLB_ERROR
	}
	fmt.Println("[cloudant_output] Cloudant service initialized successfully.")
	fmt.Println("[cloudant_output] Output Plugin initialized with Endpoint:", endpoint)
	return output.FLB_OK
}

// FLBPluginFlushCtx Processes and sends log records to Cloudant endpoint.
//
//export FLBPluginFlushCtx
func FLBPluginFlushCtx(ctx, data unsafe.Pointer, length C.int, tag *C.char) int {
	fmt.Println("[cloudant_output] In FLBPluginFlushCtx")

	// read data from fluentbit Input plugin
	dec := output.NewDecoder(data, int(length))
	var records []interface{}
	//var convertedRecord interface{}

	for {
		ret, _, record := output.GetRecord(dec)
		if ret != 0 {
			break
		}
		var err error
		// Convert the record to a map[string]interface{}
		convertedRecord, err := convertToStringKeyMap(record)
		if err != nil {
			fmt.Println("[cloudant_output] ERROR: Failed to convert record:", err)
			continue
		}
		records = append(records, convertedRecord)
	}
	// for _, rec := range records {
	// 	fmt.Println("Record:")
	// 	for k, v := range rec.(map[string]interface{}) {
	// 		fmt.Printf("got key = %v\n value=%v\n", k, v)
	// 	}
	// }

	return sendToCloudant(cloudantService, records)

}

// FLBPluginExit Cleans up resources when Fluent Bit stops.
//
//export FLBPluginExit
func FLBPluginExit() int {
	fmt.Println("[cloudant_output] In FLBPluginExit")
	fmt.Println("[cloudant_output] Plugin exiting")
	return output.FLB_OK
}

func main() {}
