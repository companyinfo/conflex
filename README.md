# Conflex

[![Conflex CI](https://github.com/companyinfo/conflex/actions/workflows/ci.yaml/badge.svg)](https://github.com/companyinfo/conflex/actions/workflows/ci.yaml) [![Go Report Card](https://goreportcard.com/badge/go.companyinfo.dev/conflex)](https://goreportcard.com/report/go.companyinfo.dev/conflex) [![Go Reference](https://pkg.go.dev/badge/go.companyinfo.dev/conflex.svg)](https://pkg.go.dev/go.companyinfo.dev/conflex)

Conflex (pronounced /ˈkɒnflɛks/) is a powerful and versatile configuration management package for Go that simplifies
handling application settings across different environments and formats.

## Features

- **Easy Integration**: Simple and intuitive API.
- **Flexible Sources**: Load from files, environment variables (with custom prefixes), Consul, and easily extend with custom sources.
- **Format Agnostic**: Supports JSON, YAML, and other formats via extensible codecs.
- **Hierarchical Merging**: Configurations from multiple sources are merged, with later sources overriding earlier ones.
- **Struct Binding**: Automatically map configuration data to Go structs.
- **Dot Notation Access**: Navigate nested configuration easily (e.g., `config.GetString("database.host")`).
- **Type-Safe Retrieval**: Get values as specific types (`string`, `int`, `bool`, etc.), with error-returning options for robust handling.
- **Configuration Dumping**: Save the effective configuration to files or other custom destinations.
- **Clear Error Handling**: Provides comprehensive error information for easier debugging.

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
	Port int    `config:"server.port"`
	Host string `config:"server.host"`
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

## Testing & Best Practices

- Use the testify suite for unit and integration tests (see `*_test.go` files).
- Mock sources and dumpers for isolated tests.
- Always check errors from `Load` and `Dump`.
- For concurrency, `Conflex` is thread-safe for `Load` and `Get`.

## Troubleshooting & FAQ

**Q: Why is my struct not being populated?**
- Make sure you pass a pointer to your struct to `WithBinding`.
- Check your struct tags: use `config:"fieldname"`.

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
    - [ ] Configuration validation (e.g., using JSON Schema)

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
