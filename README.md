# Conflex

[![Conflex CI](https://github.com/companyinfo/conflex/actions/workflows/ci.yaml/badge.svg)](https://github.com/companyinfo/conflex/actions/workflows/ci.yaml) ![Codecov](https://img.shields.io/codecov/c/github/companyinfo/conflex)
 [![Go Report Card](https://goreportcard.com/badge/go.companyinfo.dev/conflex)](https://goreportcard.com/report/go.companyinfo.dev/conflex) [![Go Reference](https://pkg.go.dev/badge/go.companyinfo.dev/conflex.svg)](https://pkg.go.dev/go.companyinfo.dev/conflex)

Conflex (pronounced /ˈkɒnflɛks/) is a powerful and versatile configuration management package for Go that simplifies
handling application settings across different environments and formats.

> **Conflex is designed to help Go applications follow best practices for configuration management as recommended by the [Twelve-Factor App methodology](https://12factor.net/), especially [Factor III: Config](https://12factor.net/config).**

---

## Table of Contents
- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Advanced Usage](#advanced-usage)
- [Custom Codecs](#custom-codecs)
- [Validation](#validation)
- [Real-World Example](#real-world-example)
- [Sample Configuration Files](#sample-configuration-files)
- [Testing & Best Practices](#testing--best-practices)
- [Troubleshooting & FAQ](#troubleshooting--faq)
- [Roadmap and Future Plans](#roadmap-and-future-plans)
- [Contributing](#contributing)
- [License](#license)

---

## Features

- **Easy Integration**: Simple and intuitive API.
- **Flexible Sources**: Load from files, environment variables (with custom prefixes), Consul, and easily extend with custom sources.
- **Format Agnostic**: Supports JSON, YAML, and other formats via extensible codecs.
- **Hierarchical Merging**: Configurations from multiple sources are merged, with later sources overriding earlier ones.
- **Struct Binding**: Automatically map configuration data to Go structs.
- **Built-in Validation**: Validate configuration using struct methods, JSON Schemas, or custom functions.
- **Dot Notation Access**: Navigate nested configuration easily (e.g., `config.GetString("database.host")`).
- **Type-Safe Retrieval**: Get values as specific types (`string`, `int`, `bool`, etc.), with error-returning options for robust handling.
- **Configuration Dumping**: Save the effective configuration to files or other custom destinations.
- **Clear Error Handling**: Provides comprehensive error information for easier debugging.
- **Thread-Safe**: Safe for concurrent access and configuration loading in multi-goroutine applications.

## Installation

```shell
go get go.companyinfo.dev/conflex
```

## Quick Start

Here's a minimal example to get you started:

```go
package main

import (
	"go.companyinfo.dev/conflex"
	"go.companyinfo.dev/conflex/codec"
	"context"
	"log"
)

func main() {
	cfg, err := conflex.New(
		conflex.WithFileSource("config.yaml", codec.TypeYAML),
		conflex.WithFileSource("config.json", codec.TypeJSON),
		// Optionally add remote sources, e.g. Consul
		conflex.WithConsulSource("staging/service", codec.TypeJSON),
	)
	if err != nil {
		log.Fatalf("failed to create configuration: %v", err)
	}

	if err := cfg.Load(context.Background()); err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	// Access configuration values
	port := cfg.GetInt("server.port")
	host := cfg.GetString("server.host")
	log.Printf("Server is running on %s:%d", host, port)
}
```

### How it works
- **Sources** are loaded in order; later sources override earlier ones.
- **Dot notation** allows deep access: `cfg.Get("database.host")`.
- **Type-safe accessors**: `GetString`, `GetInt`, `GetBool`, etc.

## Advanced Usage

### Struct Binding
Bind configuration directly to your own struct:

```go
type Config struct {
	Port int    `conflex:"server.port"`
	Host string `conflex:"server.host"`
}

var c Config
cfg, _ := conflex.New(
	conflex.WithFileSource("config.yaml", codec.TypeYAML),
	conflex.WithBinding(&c),
)
cfg.Load(context.Background())
// c.Port and c.Host are now populated
```

### Merging and Precedence
- Multiple sources are merged; later sources override earlier ones.
- Environment variables can override file/remote config:

```go
cfg, _ := conflex.New(
	conflex.WithFileSource("config.yaml", codec.TypeYAML),
	conflex.WithOSEnvVarSource("MYAPP_"), // Only env vars with prefix MYAPP_
)
```

### Remote Sources (Consul)

```go
cfg, _ := conflex.New(
	conflex.WithConsulSource("production/service", codec.TypeJSON),
)
```

> **Note:** By default, the Consul source will use the Consul API client and automatically look up the `CONSUL_HTTP_ADDR` and `CONSUL_HTTP_TOKEN` environment variables for configuration. You can override these by setting the appropriate environment variables or configuring the Consul client manually.

### Dumping Configuration

```go
cfg, _ := conflex.New(
	conflex.WithFileSource("config.yaml", codec.TypeYAML),
	conflex.WithFileDumper("out.yaml", codec.TypeYAML),
)
cfg.Load(context.Background())
cfg.Dump(context.Background()) // Writes merged config to out.yaml
```

## Custom Codecs

Conflex allows you to extend configuration support to any format by registering your own codecs.

### Implementing a Custom Codec

A codec must implement the following interface:

```go
type Codec interface {
    Encode(v any) ([]byte, error)
    Decode(data []byte, v any) error
}
```

### Example: Registering a Custom Codec

Suppose you want to support a custom format called `mytype`:

```go
package mycodec

import (
    "go.companyinfo.dev/conflex/codec"
)

type MyCodec struct{}

func (MyCodec) Encode(v any) ([]byte, error) {
    // ... your encoding logic ...
}

func (MyCodec) Decode(data []byte, v any) error {
    // ... your decoding logic ...
}

func init() {
    codec.RegisterEncoder("mytype", MyCodec{})
    codec.RegisterDecoder("mytype", MyCodec{})
}
```

Then, in your application:

```go
import (
    _ "yourmodule/mycodec" // ensure init() runs
    "go.companyinfo.dev/conflex"
    "go.companyinfo.dev/conflex/codec"
)

cfg, _ := conflex.New(
    conflex.WithFileSource("config.mytype", "mytype"),
)
```

### When to Use a Custom Codec
- Supporting formats not built-in (e.g., TOML, XML, encrypted configs)
- Integrating with legacy or proprietary configuration formats
- Adding validation or transformation logic during encode/decode

**Tip:** If you build a useful codec, consider contributing it back to the community!

## Validation

Conflex supports configuration validation to help catch errors early and ensure your application runs with correct settings.

### 1. Struct-Based Validation

If your binding struct implements the following interface, Conflex will call `Validate()` after binding:

```go
type Validator interface {
    Validate() error
}
```

**Example:**
```go
type MyConfig struct {
    Port int `conflex:"port"`
}

func (c *MyConfig) Validate() error {
    if c.Port <= 0 {
        return errors.New("port must be positive")
    }
    return nil
}

cfg, _ := conflex.New(
    conflex.WithFileSource("config.yaml", codec.TypeYAML),
    conflex.WithBinding(&myConfig),
)
err := cfg.Load(context.Background()) // Will return error if validation fails
```

### 2. JSON Schema Validation (for Maps)

> **What is JSON Schema?**  
> [JSON Schema](https://json-schema.org/) is a standard for describing the structure and validation rules of JSON data. It allows you to define required fields, data types, value constraints, and more, making it easy to validate configuration files and catch errors early.  
> Learn more at [json-schema.org](https://json-schema.org/).

You can validate the loaded configuration map against a JSON Schema using [`github.com/santhosh-tekuri/jsonschema/v6`](https://github.com/santhosh-tekuri/jsonschema):

> **Note:** JSON Schema validation in Conflex is applied to the merged configuration map (`map[string]any`), not directly to Go structs. Schema validation happens before any struct binding. If you want to validate your struct, use the struct-based `Validate() error` method described above.

```go
schemaBytes, err := os.ReadFile("schema.json")
if err != nil {
    log.Fatalf("failed to read schema: %v", err)
}
cfg, _ := conflex.New(
    conflex.WithFileSource("config.yaml", codec.TypeYAML),
    conflex.WithJSONSchema(schemaBytes),
)
```

### 3. Custom Validation Functions

You can register a custom validation function for either the bound struct or the config map:

```go
cfg, _ := conflex.New(
    conflex.WithFileSource("config.yaml", codec.TypeYAML),
    conflex.WithValidator(func(cfg map[string]any) error {
        if cfg["port"].(int) <= 0 {
            return errors.New("port must be positive")
        }
        return nil
    }),
)
```

### Summary Table

| Validation Type         | For Structs         | For Maps           | How to Use                        |
|------------------------|---------------------|--------------------|-----------------------------------|
| Interface-based        | `Validate() error`  | —                  | Implement on struct               |
| JSON Schema            | —                   | Yes                | `WithJSONSchema(schema)`          |
| Custom Function        | Yes                 | Yes                | `WithValidator(func) error`       |

**Tip:** Validation helps prevent misconfiguration and makes your application more robust!

## Real-World Example

This example demonstrates merging multiple sources (file, environment, Consul), using struct binding, validation, and a custom codec.

```go
import (
    "context"
    "go.companyinfo.dev/conflex"
    "go.companyinfo.dev/conflex/codec"
    _ "yourmodule/mycodec" // Register your custom codec
)

type Config struct {
    Port int    `conflex:"server.port"`
    Host string `conflex:"server.host"`
}

func (c *Config) Validate() error {
    if c.Port <= 0 {
        return errors.New("port must be positive")
    }
    return nil
}

var c Config
cfg, err := conflex.New(
    conflex.WithFileSource("config.yaml", codec.TypeYAML),
    conflex.WithFileSource("config.json", codec.TypeJSON),
    conflex.WithOSEnvVarSource("MYAPP_"),
    conflex.WithConsulSource("production/service", codec.TypeJSON),
    conflex.WithFileSource("config.mytype", "mytype"), // custom codec
    conflex.WithBinding(&c),
    conflex.WithValidator(func(m map[string]any) error {
        if m["feature_enabled"] != true {
            return errors.New("feature_enabled must be true")
        }
        return nil
    }),
)
if err != nil {
    log.Fatalf("failed to create configuration: %v", err)
}
if err := cfg.Load(context.Background()); err != nil {
    log.Fatalf("failed to load configuration: %v", err)
}
```

## Sample Configuration Files

**YAML (`config.yaml`):**
```yaml
server:
  port: 8080
  host: localhost
feature_enabled: true
```

**JSON (`config.json`):**
```json
{
  "server": {
    "port": 8080,
    "host": "localhost"
  },
  "feature_enabled": true
}
```

## Testing & Best Practices

- Use the testify suite for unit and integration tests (see `*_test.go` files).
- Mock sources and dumpers for isolated tests.
- Always check errors from `Load` and `Dump`.
- For concurrency, `Conflex` is thread-safe for `Load` and `Get`.

## Troubleshooting & FAQ

**Q: Why is my struct not being populated?**
- Make sure you pass a pointer to your struct to `WithBinding`.
- Check your struct tags: use `conflex:"fieldname"`.

**Q: How do I override config with environment variables?**
- Use `WithOSEnvVarSource("PREFIX_")` and set env vars like `PREFIX_SERVER_PORT=8080`.

**Q: How do I add a new config source or dumper?**
- Implement the `Source` or `Dumper` interface and pass it to `WithSource` or `WithDumper`.

**Q: How do I access nested values?**
- Use dot notation: `cfg.Get("outer.inner.key")`.

**Q: What happens if a source returns nil or an error?**
- If a source returns an error, `Load` will return it. If a source returns nil, it is skipped.

## Roadmap and Future Plans

- [ ] **Additional Configuration Formats:**
    - [ ] TOML (Tom's Obvious, Minimal Language)
    - [ ] HCL (HashiCorp Configuration Language)
    - [ ] INI
- [ ] **Additional Configuration Sources:**
    - [ ] Command-line flags (e.g., `--host=localhost`)
    - [ ] HashiCorp Vault
    - [ ] Etcd
    - [ ] Apache ZooKeeper
    - [ ] Redis / Valkey
    - [ ] Memcached
- [ ] **Advanced Features:**
    - [ ] Hot reloading of configuration changes
    - [ ] Decryption of sensitive configuration values (e.g., SOPS integration)

## Contributing

Contributions are welcome! Here's how you can contribute:

1. Fork the repository
2. Create a new branch (`git checkout -b feature/improvement`)
3. Make your changes
4. Commit your changes (`git commit -am 'Add new feature'`)
5. Push to the branch (`git push origin feature/improvement`)
6. Create a Pull Request

Please make sure to:
- Follow the existing code style
- Add tests if applicable
- Update documentation as needed
- Include a clear description of your changes in the PR

## License

Copyright &copy; 2025 Company.info

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
