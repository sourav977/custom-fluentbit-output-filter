# Custom Fluent Bit Output Plugin Design

### Context
Fluent Bit is a lightweight and high-performance log processor and forwarder that comes with various built-in output plugins to send logs to different destinations. However, Cloudant is not included as a default output plugin. Since Fluent Bit supports extending its capabilities through custom output plugins, I have developed a Fluent Bit custom output plugin for Cloudant using Go.

This plugin allows Fluent Bit to send logs directly to a IBM Cloudant database. The Fluent Bit Go SDK (fluent-bit-go) provides the necessary interface for writing custom output plugins in Go, enabling integration with Cloudant via its API.

This document explains how the plugin works, how to build and use it with Fluent Bit, and how to test its functionality. 🚀

# Custom Fluent Bit Output Plugin for Cloudant

## Overview
This project implements a **custom Fluent Bit output plugin** in Go that sends log data to **IBM Cloudant**. The plugin is built using the [fluent-bit-go](https://github.com/fluent/fluent-bit-go) package, allowing Fluent Bit to forward log records to a Cloudant database.

## Designing a Custom Fluent Bit Output Plugin
Fluent Bit provides a **Go interface** for developing custom plugins. The [fluent-bit-go](https://github.com/fluent/fluent-bit-go) package assists developers in writing output plugins using Go. A Fluent Bit output plugin should implement key lifecycle functions:

- **FLBPluginRegister**: Registers the plugin with Fluent Bit.
- **FLBPluginInit**: Initializes the plugin with user-provided configuration.
- **FLBPluginFlushCtx**: Processes incoming log records and forwards them to Cloudant.
- **FLBPluginExit**: Cleans up resources before Fluent Bit shuts down.

## Plugin Implementation
### 1. **FLBPluginRegister**
This function registers the plugin with Fluent Bit by specifying its name and version.
```go
func FLBPluginRegister(def unsafe.Pointer) int {
    return output.FLBPluginRegister(def, "cloudant_output", "Send logs to IBM Cloudant")
}
```

### 2. **FLBPluginInit**
This function initializes the plugin by reading user-provided configuration settings (Cloudant API details, authentication mode, database name, etc.) specified in `fluent-bit.conf` file, designed for Cloudant Output Plugin.
```go
func FLBPluginInit(ctx unsafe.Pointer) int {
    endpoint = output.FLBPluginConfigKey(ctx, "Endpoint")
    cloudantDatabase = output.FLBPluginConfigKey(ctx, "Database")
    authMode = output.FLBPluginConfigKey(ctx, "Authentication_Mode")
    // Initialize Cloudant client...
    return output.FLB_OK
}
```

### 3. **FLBPluginFlushCtx**
This function receives logs from Fluent Bit, processes them, and sends them to Cloudant.
```go
func FLBPluginFlushCtx(ctx, data unsafe.Pointer, length C.int, tag *C.char) int {
    dec := output.NewDecoder(data, int(length))
    var records []interface{}
    for {
        ret, _, record := output.GetRecord(dec)
        if ret != 0 {
            break
        }
        convertedRecord, err := convertToStringKeyMap(record)
        if err != nil {
            fmt.Println("Error converting record:", err)
            continue
        }
        records = append(records, convertedRecord)
    }
    return sendToCloudant(cloudantService, records)
}
```

### 4. **FLBPluginExit**
This function cleans up resources before Fluent Bit shuts down.
```go
func FLBPluginExit() int {
    return output.FLB_OK
}
```

## Fluent Bit Configuration
To use this plugin, configure Fluent Bit to load it. Below is an example Fluent Bit configuration file:
```ini
[SERVICE]
    Log_Level debug

[INPUT]
    name http
    listen 0.0.0.0
    port 9091

[PLUGINS]
    Path /path/to/out_cloudant.so

[OUTPUT]
    Name cloudant_output
    Endpoint https://your-cloudant-instance.cloudantnosqldb.appdomain.cloud
    Database fluentbit
    Authentication_Mode IAMAPIKEY
```

## Building the Plugin
After implementing the plugin logic, compile it as a **shared object (.so) file**:
```sh
go build -buildmode=c-shared -o out_cloudant.so *.go
```


## Running Fluent Bit with the Plugin
### **Method 1: Load the Plugin Directly**
You can run Fluent Bit and load the plugin dynamically using the `-e` flag:
```sh
fluent-bit -i http -p port=9091 -e ./out_cloudant.so -v
```

### **Method 2: Use a Configuration File**
Alternatively, you can specify the plugin in a Fluent Bit configuration file and start Fluent Bit using:
```sh
fluent-bit -c fluent-bit.conf -v
```

## Sending Sample Data
To test the plugin, send JSON logs to Fluent Bit's HTTP input plugin:
```sh
curl -X POST -H "Content-Type: application/json" -d '[
  { "color": "red", "item_code": 112 },
  { "color": "green", "item_code": 113 },
  { "color": "blue", "item_code": 114 }
]' http://localhost:9091
```

## Verifying Output
Check the Fluent Bit logs for Cloudant responses. If successful, the data should appear in the specified Cloudant database.

For above data, the output looks like below:

Press `Ctrl + c` to exit from fluent-bit.

<details>

<summary>fluent-bit -i http -p port=9091 -e ./out_cloudant.so  -c fluent-bit2.conf  -v</summary>

```
% fluent-bit -i http -p port=9091 -e ./out_cloudant.so  -c fluent-bit2.conf  -v                      [9/03/25 | 10:29:51]
[cloudant_output] In FLBPluginRegister
Fluent Bit v2.2.2
* Copyright (C) 2015-2024 The Fluent Bit Authors
* Fluent Bit is a CNCF sub-project under the umbrella of Fluentd
* https://fluentbit.io

____________________
< Fluent Bit v2.2.2 >
 -------------------
          \
           \
            \          __---__
                    _-       /--______
               __--( /     \ )XXXXXXXXXXX\v.
             .-XXX(   O   O  )XXXXXXXXXXXXXXX-
            /XXX(       U     )        XXXXXXX\
          /XXXXX(              )--_  XXXXXXXXXXX\
         /XXXXX/ (      O     )   XXXXXX   \XXXXX\
         XXXXX/   /            XXXXXX   \__ \XXXXX
         XXXXXX__/          XXXXXX         \__---->
 ---___  XXX__/          XXXXXX      \__         /
   \-  --__/   ___/\  XXXXXX            /  ___--/=
    \-\    ___/    XXXXXX              '--- XXXXXX
       \-\/XXX\ XXXXXX                      /XXXXX
         \XXXXXXXXX   \                    /XXXXX/
          \XXXXXX      >                 _/XXXXX/
            \XXXXX--__/              __-- XXXX/
             -XXXXXXXX---------------  XXXXXX-
                \XXXXXXXXXXXXXXXXXXXXXXXXXX/
                  ""VXXXXXXXXXXXXXXXXXXV""

[2025/03/09 10:29:53] [ info] Configuration:
[2025/03/09 10:29:53] [ info]  flush time     | 1.000000 seconds
[2025/03/09 10:29:53] [ info]  grace          | 5 seconds
[2025/03/09 10:29:53] [ info]  daemon         | 0
[2025/03/09 10:29:53] [ info] ___________
[2025/03/09 10:29:53] [ info]  inputs:
[2025/03/09 10:29:53] [ info]      http
[2025/03/09 10:29:53] [ info] ___________
[2025/03/09 10:29:53] [ info]  filters:
[2025/03/09 10:29:53] [ info] ___________
[2025/03/09 10:29:53] [ info]  outputs:
[2025/03/09 10:29:53] [ info]      cloudant_output.0
[2025/03/09 10:29:53] [ info] ___________
[2025/03/09 10:29:53] [ info]  collectors:
[2025/03/09 10:29:53] [ info] [fluent bit] version=2.2.2, commit=, pid=51707
[2025/03/09 10:29:53] [debug] [engine] coroutine stack size: 24576 bytes (24.0K)
[2025/03/09 10:29:53] [ info] [storage] ver=1.5.1, type=memory, sync=normal, checksum=off, max_chunks_up=128
[2025/03/09 10:29:53] [ info] [cmetrics] version=0.6.6
[2025/03/09 10:29:53] [ info] [ctraces ] version=0.4.0
[2025/03/09 10:29:53] [ info] [input:http:http.0] initializing
[2025/03/09 10:29:53] [ info] [input:http:http.0] storage_strategy='memory' (memory only)
[2025/03/09 10:29:53] [debug] [http:http.0] created event channels: read=22 write=23
[2025/03/09 10:29:53] [debug] [downstream] listening on 0.0.0.0:9091
[2025/03/09 10:29:53] [debug] [cloudant_output:cloudant_output.0] created event channels: read=25 write=26
[cloudant_output] In FLBPluginInit
[cloudant_output] In InitializeConfig
[cloudant_output] Configurations initialized successfully.
[cloudant_output] In ReadCloudantAPIKey
[cloudant_output] In initCloudantClient
[cloudant_output] Cloudant service initialized successfully.
[cloudant_output] Output Plugin initialized with Endpoint: 0f222168-5d29-4940-b0e9-25aaba0872a7-bluemix.cloudantnosqldb.appdomain.cloud
[2025/03/09 10:29:53] [ info] [sp] stream processor started
[2025/03/09 10:30:01] [debug] [input chunk] update output instances with new chunk size diff=47, records=1, input=http.0
[2025/03/09 10:30:01] [debug] [input chunk] update output instances with new chunk size diff=49, records=1, input=http.0
[2025/03/09 10:30:01] [debug] [input chunk] update output instances with new chunk size diff=48, records=1, input=http.0
[2025/03/09 10:30:01] [debug] [task] created task=0x600001474000 id=0 OK
[cloudant_output] In FLBPluginFlushCtx
[cloudant_output] In sendToCloudant
[cloudant_output] In generateUniqueID for cloudant doc
[cloudant_output] In generateUniqueID for cloudant doc
[cloudant_output] In generateUniqueID for cloudant doc
[cloudant_output] Successfully sent all records to Cloudant.
[2025/03/09 10:30:04] [debug] [out flush] cb_destroy coro_id=0
[2025/03/09 10:30:04] [debug] [task] destroy task=0x600001474000 (task_id=0)
[2025/03/09 10:30:31] [debug] [input chunk] update output instances with new chunk size diff=47, records=1, input=http.0
[2025/03/09 10:30:31] [debug] [task] created task=0x600001474000 id=0 OK
[cloudant_output] In FLBPluginFlushCtx
[cloudant_output] In sendToCloudant
[cloudant_output] In generateUniqueID for cloudant doc
[cloudant_output] Successfully sent all records to Cloudant.
[2025/03/09 10:30:32] [debug] [out flush] cb_destroy coro_id=1
[2025/03/09 10:30:32] [debug] [task] destroy task=0x600001474000 (task_id=0)
^C[2025/03/09 10:30:40] [engine] caught signal (SIGINT)
[cloudant_output] In FLBPluginExit
[cloudant_output] Plugin exiting

```

</details>

---
### **Next Steps**
- Enhance logging for debugging failures.
- Implement retries and concurrency for better performance in Kubernetes environments.
- Run this in Kubernetes as `kind: DaemonSet`.

🚀 **Happy Logging with Fluent Bit & Cloudant!**


