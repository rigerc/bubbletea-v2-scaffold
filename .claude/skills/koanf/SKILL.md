---
name: koanf
description: >-
  Generates, explains, and debugs configuration management code using the koanf Go library (github.com/knadh/koanf/v2). Use when building Go applications that need to load configuration from files (YAML, TOML, JSON, HCL, dotenv), environment variables, command-line flags, remote stores (Vault, Consul, etcd, S3, AWS AppConfig/ParameterStore), or in-memory sources. Covers provider chaining, config merging, struct unmarshaling, file watching, default values, custom parsers/providers, and migration from viper.
allowed-tools: Read, Write, Edit, Bash(go *)
metadata:
  category: go-cli
  version: "2.0"
---

## Overview

koanf is a lightweight, cleaner alternative to spf13/viper with better abstractions, extensibility, and far fewer dependencies. It separates configuration concerns into:

- **Providers**: Sources that provide configuration (files, env vars, flags, remote stores)
- **Parsers**: Convert raw bytes to nested maps (JSON, YAML, TOML, HCL, dotenv)
- **Koanf**: The core that merges configs and provides type-safe accessors

Key features:
- Case-sensitive keys (unlike viper)
- No forced lowercasing of keys
- Detached dependencies (install only what you need)
- Thread-safe with mutex protection
- Provider chaining with merge semantics
- File watching for hot-reload
- Custom merge strategies

## Installation

```bash
# Core library
go get -u github.com/knadh/koanf/v2

# Providers (install only what you need)
go get -u github.com/knadh/koanf/providers/file
go get -u github.com/knadh/koanf/providers/env/v2
go get -u github.com/knadh/koanf/providers/posflag
go get -u github.com/knadh/koanf/providers/confmap
go get -u github.com/knadh/koanf/providers/structs
go get -u github.com/knadh/koanf/providers/rawbytes
go get -u github.com/knadh/koanf/providers/s3
go get -u github.com/knadh/koanf/providers/vault/v2
go get -u github.com/knadh/koanf/providers/consul/v2
go get -u github.com/knadh/koanf/providers/etcd/v2
go get -u github.com/knadh/koanf/providers/parameterstore/v2

# Parsers (install only what you need)
go get -u github.com/knadh/koanf/parsers/yaml
go get -u github.com/knadh/koanf/parsers/toml
go get -u github.com/knadh/koanf/parsers/json
go get -u github.com/knadh/koanf/parsers/hcl
go get -u github.com/knadh/koanf/parsers/dotenv
```

## Basic Usage Pattern

```go
package main

import (
    "log"
    
    "github.com/knadh/koanf/v2"
    "github.com/knadh/koanf/parsers/yaml"
    "github.com/knadh/koanf/providers/file"
)

var k = koanf.New(".")

func main() {
    if err := k.Load(file.Provider("config.yaml"), yaml.Parser()); err != nil {
        log.Fatalf("error loading config: %v", err)
    }
    
    host := k.String("server.host")
    port := k.Int("server.port")
}
```

## Type-Safe Getters

All getters return zero values if key doesn't exist. Use `Must*` variants to panic on missing/zero values.

```go
k.String("key")           // string
k.Strings("key")          // []string
k.Int("key")              // int
k.Ints("key")             // []int
k.Int64("key")            // int64
k.Int64s("key")           // []int64
k.Float64("key")          // float64
k.Float64s("key")         // []float64
k.Bool("key")             // bool
k.Bools("key")            // []bool
k.Duration("key")         // time.Duration (parses "1h30m" or nanoseconds)
k.Time("key", layout)     // time.Time (parses string with layout)
k.Bytes("key")            // []byte
k.StringMap("key")        // map[string]string
k.StringsMap("key")       // map[string][]string
k.IntMap("key")           // map[string]int
k.BoolMap("key")          // map[string]bool
k.Get("key")              // any (raw value)
k.Exists("key")           // bool
k.Keys()                  // []string (all flattened keys)
k.All()                   // map[string]any (flattened)
k.Raw()                   // map[string]any (nested)
```

## Provider Chaining and Merge Order

Load multiple providers in sequence. Later providers override earlier ones.

```go
k := koanf.New(".")

k.Load(confmap.Provider(map[string]any{
    "server.port": 8080,
    "server.host": "localhost",
}, "."), nil)

k.Load(file.Provider("config.yaml"), yaml.Parser())

k.Load(env.Provider(".", env.Opt{Prefix: "APP_"}), nil)

k.Load(posflag.Provider(flagSet, ".", k), nil)
```

## Loading from Files

### YAML

```go
import "github.com/knadh/koanf/parsers/yaml"

k.Load(file.Provider("config.yaml"), yaml.Parser())
```

### TOML

```go
import "github.com/knadh/koanf/parsers/toml"

k.Load(file.Provider("config.toml"), toml.Parser())
```

### JSON

```go
import "github.com/knadh/koanf/parsers/json"

k.Load(file.Provider("config.json"), json.Parser())
```

### HCL

```go
import "github.com/knadh/koanf/parsers/hcl"

k.Load(file.Provider("config.hcl"), hcl.Parser(true))
```

### dotenv

```go
import "github.com/knadh/koanf/parsers/dotenv"

k.Load(file.Provider(".env"), dotenv.Parser())
```

## Environment Variables

```go
import "github.com/knadh/koanf/providers/env/v2"

k.Load(env.Provider(".", env.Opt{
    Prefix: "MYAPP_",
    TransformFunc: func(k, v string) (string, any) {
        k = strings.ToLower(strings.TrimPrefix(k, "MYAPP_"))
        k = strings.ReplaceAll(k, "_", ".")
        if strings.Contains(v, " ") {
            return k, strings.Split(v, " ")
        }
        return k, v
    },
    EnvironFunc: func() []string {
        return os.Environ()
    },
}), nil)
```

## Command-Line Flags

### spf13/pflag

```go
import (
    "github.com/knadh/koanf/providers/posflag"
    flag "github.com/spf13/pflag"
)

f := flag.NewFlagSet("config", flag.ContinueOnError)
f.String("host", "localhost", "server host")
f.Int("port", 8080, "server port")
f.Parse(os.Args[1:])

k.Load(posflag.Provider(f, ".", k), nil)
```

### stdlib flag

```go
import "github.com/knadh/koanf/providers/basicflag"

f := flag.NewFlagSet("config", flag.ContinueOnError)
f.String("host", "localhost", "server host")
f.Parse(os.Args[1:])

k.Load(basicflag.Provider(f, "."), nil)
```

## Default Values

### Using confmap Provider

```go
import "github.com/knadh/koanf/providers/confmap"

k.Load(confmap.Provider(map[string]any{
    "server.host": "localhost",
    "server.port": 8080,
}, "."), nil)
```

### Using structs Provider

```go
import "github.com/knadh/koanf/providers/structs"

type Config struct {
    Server struct {
        Host string `koanf:"host"`
        Port int    `koanf:"port"`
    } `koanf:"server"`
}

defaults := Config{}
defaults.Server.Host = "localhost"
defaults.Server.Port = 8080

k.Load(structs.Provider(defaults, "koanf"), nil)
```

## Struct Unmarshaling

```go
type ServerConfig struct {
    Host string `koanf:"host"`
    Port int    `koanf:"port"`
}

type Config struct {
    Server ServerConfig `koanf:"server"`
    Debug  bool         `koanf:"debug"`
}

var cfg Config
k.Unmarshal("", &cfg)
```

### Advanced Unmarshaling

```go
k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{
    Tag:       "koanf",
    FlatPaths: false,
    DecoderConfig: &mapstructure.DecoderConfig{
        TagName: "koanf",
        // ... custom decoder config
    },
})
```

### Flat Paths Unmarshaling

```go
type FlatConfig struct {
    ServerHost string `koanf:"server.host"`
    ServerPort int    `koanf:"server.port"`
}

var cfg FlatConfig
k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{
    Tag:       "koanf",
    FlatPaths: true,
})
```

## File Watching (Hot Reload)

```go
f := file.Provider("config.yaml")

if err := k.Load(f, yaml.Parser()); err != nil {
    log.Fatal(err)
}

f.Watch(func(event any, err error) {
    if err != nil {
        log.Printf("watch error: %v", err)
        return
    }
    
    log.Println("config changed, reloading...")
    k = koanf.New(".")
    k.Load(f, yaml.Parser())
})

defer f.Unwatch()
```

## Remote Stores

### HashiCorp Vault

```go
import "github.com/knadh/koanf/providers/vault/v2"

provider, err := vault.Provider(vault.Config{
    Address:     "http://localhost:8200",
    Token:       "my-token",
    Path:        "secret/data/my-app",
    Timeout:     10 * time.Second,
    ExcludeMeta: true,
})
if err != nil {
    log.Fatal(err)
}

k.Load(provider, nil)
```

### HashiCorp Consul

```go
import (
    "github.com/hashicorp/consul/api"
    "github.com/knadh/koanf/providers/consul/v2"
)

provider, err := consul.Provider(consul.Config{
    Key:     "my-app/config",
    Recurse: true,
    Cfg:     api.DefaultConfig(),
})
if err != nil {
    log.Fatal(err)
}

k.Load(provider, nil)
```

### etcd

```go
import "github.com/knadh/koanf/providers/etcd/v2"

provider, err := etcd.Provider(etcd.Config{
    Endpoints:   []string{"localhost:2379"},
    DialTimeout: 5 * time.Second,
    Key:         "my-app/config",
    Prefix:      true,
})
if err != nil {
    log.Fatal(err)
}

k.Load(provider, nil)
```

### AWS S3

```go
import "github.com/knadh/koanf/providers/s3"

k.Load(s3.Provider(s3.Config{
    AccessKey: os.Getenv("AWS_ACCESS_KEY"),
    SecretKey: os.Getenv("AWS_SECRET_KEY"),
    Region:    os.Getenv("AWS_REGION"),
    Bucket:    "my-config-bucket",
    ObjectKey: "config.json",
}), json.Parser())
```

### AWS Parameter Store

```go
import "github.com/knadh/koanf/providers/parameterstore/v2"

provider, err := parameterstore.Provider(parameterstore.Config[ssm.GetParametersByPathInput]{
    Delim: "/",
    Input: ssm.GetParametersByPathInput{
        Path:           aws.String("/my-app/"),
        WithDecryption: aws.Bool(true),
    },
    Callback: func(key, value string) (string, any) {
        return strings.TrimPrefix(key, "/my-app/"), value
    },
})
if err != nil {
    log.Fatal(err)
}

k.Load(provider, nil)
```

## Raw Bytes Provider

```go
import "github.com/knadh/koanf/providers/rawbytes"

b := []byte(`{"server": {"host": "localhost", "port": 8080}}`)
k.Load(rawbytes.Provider(b), json.Parser())
```

## Marshaling (Config to Bytes)

```go
b, err := k.Marshal(json.Parser())
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(b))
```

## Cut and Merge Operations

```go
subConfig := k.Cut("server")

subConfig.String("host")

k.Merge(subConfig)

k.MergeAt(subConfig, "parent.child")

k.Set("server.host", "new-host")

k.Delete("server")
```

## Strict Merging

```go
k := koanf.NewWithConf(koanf.Conf{
    Delim:       ".",
    StrictMerge: true,
})

if err := k.Load(file.Provider("config1.yaml"), yaml.Parser()); err != nil {
    log.Fatal(err)
}

if err := k.Load(file.Provider("config2.yaml"), yaml.Parser()); err != nil {
    log.Fatal(err)
}
```

## Custom Merge Function

```go
k.Load(file.Provider("override.yaml"), yaml.Parser(), koanf.WithMergeFunc(func(src, dest map[string]any) error {
    for key, value := range src {
        dest[key] = value
    }
    return nil
}))
```

## Custom Providers

```go
type MyProvider struct{}

func (p *MyProvider) ReadBytes() ([]byte, error) {
    return []byte(`{"key": "value"}`), nil
}

func (p *MyProvider) Read() (map[string]any, error) {
    return map[string]any{"key": "value"}, nil
}

k.Load(&MyProvider{}, nil)
```

## Custom Parsers

```go
type MyParser struct{}

func (p *MyParser) Unmarshal(b []byte) (map[string]any, error) {
    var out map[string]any
    if err := json.Unmarshal(b, &out); err != nil {
        return nil, err
    }
    return out, nil
}

func (p *MyParser) Marshal(m map[string]any) ([]byte, error) {
    return json.Marshal(m)
}

k.Load(file.Provider("config.myformat"), &MyParser{})
```

## Migration from viper

Key differences when migrating:

1. Keys are case-sensitive (viper lowercases them)
2. Install providers/parsers separately (viper bundles everything)
3. No implicit ordering - you control load order
4. Getters return copies, not references (safe to mutate)
5. `koanf` tag instead of `mapstructure`
6. No built-in remote support (use providers)

```go
// viper
viper.SetConfigFile("config.yaml")
viper.ReadInConfig()
viper.GetString("server.host")

// koanf
k := koanf.New(".")
k.Load(file.Provider("config.yaml"), yaml.Parser())
k.String("server.host")
```

## Common Patterns

### Layered Configuration

```go
func loadConfig() (*koanf.Koanf, error) {
    k := koanf.New(".")
    
    if err := k.Load(structs.Provider(defaultConfig{}, "koanf"), nil); err != nil {
        return nil, err
    }
    
    if err := k.Load(file.Provider("/etc/myapp/config.yaml"), yaml.Parser()); err != nil && !os.IsNotExist(err) {
        return nil, err
    }
    
    if err := k.Load(file.Provider("./config.yaml"), yaml.Parser()); err != nil && !os.IsNotExist(err) {
        return nil, err
    }
    
    if err := k.Load(env.Provider(".", env.Opt{Prefix: "MYAPP_"}), nil); err != nil {
        return nil, err
    }
    
    if err := k.Load(posflag.Provider(flagSet, ".", k), nil); err != nil {
        return nil, err
    }
    
    return k, nil
}
```

### Configuration Validation

```go
func validateConfig(k *koanf.Koanf) error {
    if !k.Exists("server.host") {
        return errors.New("server.host is required")
    }
    if port := k.Int("server.port"); port < 1 || port > 65535 {
        return fmt.Errorf("server.port must be between 1 and 65535, got %d", port)
    }
    return nil
}
```

## Best Practices

1. **Load order matters**: defaults -> file -> env -> flags
2. **Use confmap for defaults**: simpler than structs for flat defaults
3. **Pass k to posflag.Provider**: handles default flag values correctly
4. **Watch with mutex**: concurrent reads during reload need synchronization
5. **Validate after load**: koanf doesn't validate, do it explicitly
6. **Use Must* getters sparingly**: prefer explicit error handling
7. **Install only needed providers/parsers**: keeps binaries small

## Available Providers

| Provider | Package | Description |
|----------|---------|-------------|
| file | providers/file | Read from filesystem |
| fs | providers/fs | Read from fs.FS (embed.FS) |
| env | providers/env/v2 | Environment variables |
| posflag | providers/posflag | spf13/pflag flags |
| basicflag | providers/basicflag | stdlib flag |
| confmap | providers/confmap | Map[string]any |
| structs | providers/structs | Struct with tags |
| rawbytes | providers/rawbytes | []byte slice |
| s3 | providers/s3 | AWS S3 |
| vault | providers/vault/v2 | HashiCorp Vault |
| consul | providers/consul/v2 | HashiCorp Consul |
| etcd | providers/etcd/v2 | CNCF etcd |
| parameterstore | providers/parameterstore/v2 | AWS SSM Parameter Store |
| appconfig | providers/appconfig/v2 | AWS AppConfig |
| azkeyvault | providers/azkeyvault | Azure Key Vault |
| cliflagv2 | providers/cliflagv2 | urfave/cli v2 |
| cliflagv3 | providers/cliflagv3 | urfave/cli v3 |
| kiln | providers/kiln | Environment with dotenv |

## Available Parsers

| Parser | Package | Description |
|--------|---------|-------------|
| yaml | parsers/yaml | YAML |
| json | parsers/json | JSON |
| toml | parsers/toml | TOML (go-toml v1) |
| toml/v2 | parsers/toml/v2 | TOML (go-toml v2) |
| dotenv | parsers/dotenv | dotenv files |
| hcl | parsers/hcl | HashiCorp HCL |
| hjson | parsers/hjson | HJSON |
| huml | parsers/huml | HUML |
| nestedtext | parsers/nestedtext | NestedText |
