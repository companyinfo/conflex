# Conflex

[![Conflex CI](https://github.com/companyinfo/conflex/actions/workflows/ci.yaml/badge.svg)](https://github.com/companyinfo/conflex/actions/workflows/ci.yaml) [![codecov](https://codecov.io/gh/companyinfo/conflex/graph/badge.svg?token=JwgTSjIfXS)](https://codecov.io/gh/companyinfo/conflex)
 [![Go Report Card](https://goreportcard.com/badge/go.companyinfo.dev/conflex)](https://goreportcard.com/report/go.companyinfo.dev/conflex) [![Go Reference](https://pkg.go.dev/badge/go.companyinfo.dev/conflex.svg)](https://pkg.go.dev/go.companyinfo.dev/conflex)

Conflex (pronounced /ˈkɒnflɛks/) is a powerful and versatile configuration management package for Go that simplifies
handling application settings across different environments and formats.

> **Conflex is designed to help Go applications follow best practices for configuration management as recommended by the [Twelve-Factor App methodology](https://12factor.net/), especially [Factor III: Config](https://12factor.net/config).**

---

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Error Handling](#error-handling)
- [Advanced Usage](#advanced-usage)
  - [Environment Variable Naming Conventions](#environment-variable-naming-conventions)
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
- **Format Agnostic**: Supports JSON, YAML, TOML, and other formats via extensible codecs.
- **Type Casting**: Built-in caster codecs for automatic type conversion (bool, int, float, time, duration, etc.).
- **Hierarchical Merging**: Configurations from multiple sources are merged, with later sources overriding earlier ones.
- **Struct Binding**: Automatically map configuration data to Go structs.
- **Built-in Validation**: Validate configuration using struct methods, JSON Schemas, or custom functions.
- **Dot Notation Access**: Navigate nested configuration easily (e.g., `config.GetString("database.host")`).
- **Type-Safe Retrieval**: Get values as specific types (`string`, `int`, `bool`, etc.), with error-returning options for robust handling.
- **Configuration Dumping**: Save the effective configuration to files or other custom destinations.
- **Clear Error Handling**: Provides comprehensive error information for easier debugging.
- **Thread-Safe**: Safe for concurrent access and configuration loading in multi-goroutine applications.
- **Nil-Safe Operations**: All getter methods handle nil Conflex instances gracefully, returning appropriate zero values or errors.
- **Consistent Return Types**: Error versions of getter methods return empty types (empty slices, maps) instead of nil for missing keys.

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
- **Context validation**: Both `Load()` and `Dump()` methods validate that context is not nil.
- **Error handling**: All methods return descriptive errors for easier debugging.

### Built-in Codecs

Conflex comes with several built-in codecs:

- **JSON**: `codec.TypeJSON` - Standard JSON format
- **YAML**: `codec.TypeYAML` - YAML format
- **TOML**: `codec.TypeTOML` - TOML format
- **Environment Variables**: `codec.TypeEnvVar` - For environment variable parsing

#### Caster Codecs

Conflex also provides caster codecs for automatic type conversion:

- **Boolean**: `codec.TypeCasterBool` - Converts to bool
- **Integer**: `codec.TypeCasterInt`, `codec.TypeCasterInt8`, `codec.TypeCasterInt16`, `codec.TypeCasterInt32`, `codec.TypeCasterInt64`
- **Unsigned Integer**: `codec.TypeCasterUint`, `codec.TypeCasterUint8`, `codec.TypeCasterUint16`, `codec.TypeCasterUint32`, `codec.TypeCasterUint64`
- **Float**: `codec.TypeCasterFloat32`, `codec.TypeCasterFloat64`
- **String**: `codec.TypeCasterString` - Converts to string
- **Time**: `codec.TypeCasterTime` - Converts to time.Time
- **Duration**: `codec.TypeCasterDuration` - Converts to time.Duration

> **Note:** Environment variable codec (`codec.TypeEnvVar`) only supports decoding. Attempting to encode will return an error indicating that encoding to environment variables is not supported.

## Error Handling

Conflex provides comprehensive error handling with detailed context information through the `ConfigError` type.

### ConfigError Structure

```go
type ConfigError struct {
    Source    string // The source where the error occurred (e.g., "source[0]", "json-schema", "binding")
    Field     string // The specific field where the error occurred (optional)
    Operation string // The operation being performed (e.g., "load", "validate", "bind", "merge")
    Err       error  // The underlying error
}
```

### Error Examples

```go
// Source loading error
err := cfg.Load(context.Background())
// Error: "config error in source[0] during load: failed to read file: no such file or directory"

// Validation error
err := cfg.Load(context.Background())
// Error: "config error in json-schema during validate: invalid schema"

// Binding error
err := cfg.Load(context.Background())
// Error: "config error in binding during bind: failed to decode configuration"
```

### Getter Method Error Handling

Getter methods come in two variants:

1. **Non-error versions**: Return zero values for missing keys or nil instances

   ```go
   cfg.GetString("nonexistent") // Returns empty string
   cfg.GetInt("nonexistent")    // Returns 0
   cfg.GetBool("nonexistent")   // Returns false
   ```

2. **Error versions**: Return errors for missing keys or nil instances

   ```go
   cfg.GetStringE("nonexistent") // Returns ("", error)
   cfg.GetIntE("nonexistent")    // Returns (0, error)
   cfg.GetBoolE("nonexistent")   // Returns (false, error)
   ```

When called on a nil Conflex instance, error versions return "conflex instance is nil" error.

## Advanced Usage

### Struct Binding

Bind configuration directly to your own struct:

```go
type Config struct {
    Port int    `conflex:"port"`
    Host string `conflex:"host"`
}

var c Config
cfg, _ := conflex.New(
    conflex.WithFileSource("config.yaml", codec.TypeYAML),
    conflex.WithBinding(&c),
)
cfg.Load(context.Background())
// c.Port and c.Host are now populated
```

### Environment Variable Naming Conventions

Conflex provides powerful environment variable support that automatically maps environment variables to nested configuration structures. This follows the [Twelve-Factor App methodology](https://12factor.net/config) for configuration management.

#### Basic Usage

```go
cfg, _ := conflex.New(
    conflex.WithFileSource("config.yaml", codec.TypeYAML),
    conflex.WithOSEnvVarSource("MYAPP_"), // Only env vars with prefix MYAPP_
)
```

#### Naming Convention Rules

Conflex uses a **hierarchical naming convention** where underscores (`_`) in environment variable names create nested configuration structures:

1. **Environment variables are converted to lowercase**
2. **Underscores (`_`) create nested levels**
3. **Empty parts (consecutive underscores) are filtered out**
4. **Values are automatically trimmed of whitespace**

#### Examples

| Environment Variable | Configuration Path | Value |
|---------------------|-------------------|-------|
| `MYAPP_SERVER_PORT` | `server.port` | `8080` |
| `MYAPP_DATABASE_HOST` | `database.host` | `localhost` |
| `MYAPP_DATABASE_USER_NAME` | `database.user.name` | `admin` |
| `MYAPP_FOO__BAR` | `foo.bar` | `value` |
| `MYAPP_A_B_C_D` | `a.b.c.d` | `nested` |

#### Struct Field Mapping

When using struct binding, environment variables map directly to struct fields using the `conflex` tag:

```go
type Config struct {
    Port     int    `conflex:"port"`
    Host     string `conflex:"host"`
    Database struct {
        Host     string `conflex:"host"`
        Port     int    `conflex:"port"`
        Username string `conflex:"username"`
        Password string `conflex:"password"`
    } `conflex:"database"`
}
```

**Environment variables needed:**

```bash
export MYAPP_PORT=8080
export MYAPP_HOST=localhost
export MYAPP_DATABASE_HOST=db.example.com
export MYAPP_DATABASE_PORT=5432
export MYAPP_DATABASE_USERNAME=admin
export MYAPP_DATABASE_PASSWORD=secret123
```

#### Advanced Examples

**Complex Nested Configuration:**

```go
type AppConfig struct {
    Server struct {
        Host string `conflex:"host"`
        Port int    `conflex:"port"`
        TLS  struct {
            Enabled  bool   `conflex:"enabled"`
            CertFile string `conflex:"cert_file"`
            KeyFile  string `conflex:"key_file"`
        } `conflex:"tls"`
    } `conflex:"server"`
    Database struct {
        Primary struct {
            Host     string `conflex:"host"`
            Port     int    `conflex:"port"`
            Database string `conflex:"database"`
        } `conflex:"primary"`
        Replica struct {
            Host     string `conflex:"host"`
            Port     int    `conflex:"port"`
            Database string `conflex:"database"`
        } `conflex:"replica"`
    } `conflex:"database"`
}
```

**Required environment variables:**

```bash
export MYAPP_SERVER_HOST=0.0.0.0
export MYAPP_SERVER_PORT=8080
export MYAPP_SERVER_TLS_ENABLED=true
export MYAPP_SERVER_TLS_CERT_FILE=/etc/ssl/certs/server.crt
export MYAPP_SERVER_TLS_KEY_FILE=/etc/ssl/private/server.key
export MYAPP_DATABASE_PRIMARY_HOST=primary.db.example.com
export MYAPP_DATABASE_PRIMARY_PORT=5432
export MYAPP_DATABASE_PRIMARY_DATABASE=myapp
export MYAPP_DATABASE_REPLICA_HOST=replica.db.example.com
export MYAPP_DATABASE_REPLICA_PORT=5432
export MYAPP_DATABASE_REPLICA_DATABASE=myapp
```

#### Edge Cases and Special Handling

**Consecutive Underscores:**

- `MYAPP_FOO__BAR` → `foo.bar` (empty parts filtered out)
- `MYAPP_A___B` → `a.b` (multiple empty parts filtered)

**Type Conflicts:**
If an environment variable creates a conflict between scalar and nested values, the nested structure takes precedence:

```bash
export MYAPP_FOO=scalar_value
export MYAPP_FOO_BAR=nested_value
# Result: foo.bar = "nested_value" (scalar "foo" is overwritten)
```

**Whitespace Handling:**

- Keys and values are automatically trimmed of whitespace
- `MYAPP_KEY = value` → `key = "value"`

#### Best Practices

1. **Use Descriptive Prefixes:** Always use application-specific prefixes to avoid conflicts

   ```bash
   # Good
   export MYAPP_DATABASE_HOST=localhost
   export WEBAPP_DATABASE_HOST=localhost
   
   # Avoid
   export DATABASE_HOST=localhost  # Too generic
   ```

2. **Consistent Naming:** Use consistent naming patterns across your application

   ```bash
   # Consistent pattern
   export MYAPP_SERVER_HOST=localhost
   export MYAPP_SERVER_PORT=8080
   export MYAPP_SERVER_TIMEOUT=30s
   ```

3. **Documentation:** Document your environment variables in your application's README

   ```bash
   # Required environment variables:
   # MYAPP_SERVER_HOST - Server hostname (default: localhost)
   # MYAPP_SERVER_PORT - Server port (default: 8080)
   # MYAPP_DATABASE_HOST - Database hostname
   # MYAPP_DATABASE_PORT - Database port (default: 5432)
   ```

4. **Validation:** Use struct validation to ensure required environment variables are set

   ```go
   func (c *Config) Validate() error {
       if c.Server.Host == "" {
           return errors.New("MYAPP_SERVER_HOST is required")
       }
       if c.Server.Port <= 0 {
           return errors.New("MYAPP_SERVER_PORT must be positive")
       }
       return nil
   }
   ```

#### Merging and Precedence

- Multiple sources are merged; later sources override earlier ones.
- Environment variables can override file/remote config:

```go
cfg, _ := conflex.New(
    conflex.WithFileSource("config.yaml", codec.TypeYAML),
    conflex.WithOSEnvVarSource("MYAPP_"), // Only env vars with prefix MYAPP_
)
```

### Content Sources

Load configuration from byte slices (useful for testing or dynamic configuration):

```go
configData := []byte(`{"server": {"port": 8080, "host": "localhost"}}`)
cfg, _ := conflex.New(
    conflex.WithContentSource(configData, codec.TypeJSON),
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

#### Configurable File Permissions

You can customize file permissions when dumping configuration:

```go
import "go.companyinfo.dev/conflex/dumper"

// Use default permissions (0644)
fileDumper := dumper.NewFile("config.yaml", encoder)

// Use custom permissions
fileDumper := dumper.NewFileWithPermissions("config.yaml", encoder, 0600)
```

The default file permissions are defined by the `DefaultFilePermissions` constant (0644).

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

**Note:** If you need type conversion functionality, consider using the built-in caster codecs (e.g., `codec.TypeCasterInt`, `codec.TypeCasterBool`) instead of creating a custom codec for simple type casting.

### When to Use a Custom Codec

- Supporting formats not built-in (e.g., XML, encrypted configs)
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
    Port int    `conflex:"port"`
    Host string `conflex:"host"`
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

**TOML (`config.toml`):**

```toml
[server]
port = 8080
host = "localhost"

feature_enabled = true
```

## Testing & Best Practices

- Use the testify suite for unit and integration tests (see `*_test.go` files).
- Mock sources and dumpers for isolated tests.
- Always check errors from `Load` and `Dump`.
- For concurrency, `Conflex` is thread-safe for `Load` and `Get`.

## Troubleshooting & FAQ

**Q: Why is my struct not being populated?**

- Make sure you pass a pointer to your struct to `WithBinding`.
- Check your struct tags: use `conflex:"fieldname"` for the field name within its context.

**Q: How do I override config with environment variables?**

- Use `WithOSEnvVarSource("PREFIX_")` and set env vars like `PREFIX_SERVER_PORT=8080`.
- Environment variables follow the naming convention: `PREFIX_SECTION_SUBSECTION_KEY=value`
- For nested structures, use underscores: `PREFIX_DATABASE_USER_NAME=admin` maps to `database.user.name`

**Q: How do environment variables map to struct fields?**

- Environment variables are converted to lowercase and split by underscores
- Use the `conflex` tag to map to the field name within its struct context: `conflex:"port"`
- Example: `MYAPP_SERVER_PORT=8080` with tag `conflex:"port"` in a struct tagged with `conflex:"server"` populates the struct field

**Q: What happens with consecutive underscores in environment variable names?**

- Consecutive underscores are filtered out: `MYAPP_FOO__BAR` becomes `foo.bar`
- This allows for cleaner environment variable names while maintaining the same configuration structure

**Q: How do I add a new config source or dumper?**

- Implement the `Source` or `Dumper` interface and pass it to `WithSource` or `WithDumper`.

**Q: How do I access nested values?**

- Use dot notation: `cfg.Get("outer.inner.key")`.

**Q: What happens if a source returns nil or an error?**

- If a source returns an error, `Load` will return it. If a source returns nil, it is skipped.

**Q: What happens when I call getter methods on a nil Conflex instance?**

- Non-error versions return appropriate zero values (empty string, 0, false, etc.)
- Error versions return "conflex instance is nil" error
- This prevents panic and provides graceful degradation

**Q: What do getter methods return for missing keys?**

- Non-error versions return zero values for the type (empty string, 0, false, empty slices/maps)
- Error versions return the zero value plus an error describing the missing key
- Error versions for slice/map types return empty slices/maps instead of nil for consistency

## Roadmap and Future Plans

- [ ] **Additional Configuration Formats:**
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
