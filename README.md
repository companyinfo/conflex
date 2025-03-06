# Conflex

Conflex (pronounced /ˈkɒnflɛks/) is a powerful and versatile configuration management package for Go that simplifies
handling application settings across different environments and formats.

## Features

- Simple and intuitive API for easy integration
- Support for multiple configuration formats (e.g., JSON, YAML)
- Support for multiple configuration sources (e.g., files, environment variables)
- Support for multiple configuration dumping destinations (files)
- Smart configuration merging with override policies
- Remote configuration sources (e.g., Consul)
- Automatic environment variable binding with customizable prefixes
- Deep nested configuration support with dot notation
- Type-safe configuration access with default values
- Comprehensive error handling and logging
- Performance optimized for large configurations

## Installation

```shell
go get companyinfo.dev/conflex
```

## Quick Start

```go
package main

import (
	"companyinfo.dev/conflex"
	"companyinfo.dev/conflex/codec"
	"context"
	"log"
)

func main() {
	cfg, err := conflex.New(
		conflex.WithFileSource("config.yaml", codec.TypeYAML),
		conflex.WithFileSource("config.json", codec.TypeJSON),
		// By default, it will use the Consul API client and 
		// lookup CONSUL_HTTP_ADDR and CONSUL_HTTP_TOKEN environment variables
		conflex.WithConsulSource(nil, "staging/service", codec.TypeJSON),
	)
	if err != nil {
		log.Fatalf("failed to create configuration: %v", err)
	}

	err = cfg.Load(context.Background())
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	// Access configuration values
	port := cfg.GetInt("server.port")
	host := cfg.GetString("server.host")
	log.Printf("Server is running on %s:%d", host, port)
}

```

## Roadmap and Future Plans

- [ ] Support for more configuration formats:
    - [ ] TOML (Tom's Obvious, Minimal Language)
    - [ ] HCL (HashiCorp Configuration Language)
    - [ ] INI (INI Configuration)
- [ ] Support for more configuration sources:
    - [ ] Command line flags (e.g., `--host=localhost`)
    - [ ] Vault (HashiCorp Vault)
    - [ ] Etcd
    - [ ] Zookeeper (Apache ZooKeeper)
    - [ ] Redis/Valkey (Redis Key-Value Store)
    - [ ] Memcached (Memcached Key-Value Store)
- [ ] Support for hot reloading of configuration changes
- [ ] Support for decryption of sensitive configuration values (e.g., SOPS)
- [ ] Support for configuration validation with JSON Schema

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
