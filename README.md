# go-config [![Build Status](https://travis-ci.org/Yalantis/go-config.svg?branch=master)](https://travis-ci.org/Yalantis/go-config)

go-config allows to initialize configuration in flexible way using from default, file, environment variables value.  
### Initialization
Done in three steps:  
  1. init with value from `default` tag
  2. merge with config file if `filepath` is provided
  3. override with environment variables which stored under `envconfig` tag

### Supported file extensions
- json

### Supported types
- Standard types: `bool`, `float`, `int`(`uint`), `slice`, `string`
- `time.Duration`, `time.Time`: full support with aliases `config.Duration`, `config.Time`
- Custom types, slice of custom types

### Usage

#### Default value
```go
type Server struct {
    Addr string `default:"localhost:8080"`
}
```
#### Environment value
```go
type Server struct {
    Addr string `envconfig:"SERVER_ADDR"`
}
```
#### Combined default, json, env
```go
type Server struct {
    Addr string `json:"addr" envconfig:"SERVER_ADDR" default:"localhost:8080"`
}
```
#### Slice
Default strings separator is comma. 
```
REDIS_ADDR=127.0.0.1:6377,127.0.0.1:6378,127.0.0.1:6379
```
```go
type Redis struct {
    Addrs []string `json:"addrs" envconfig:"REDIS_ADDR" default:"localhost:6378,localhost:6379"`
}
```
Slice of structs could be parsed from environment by defining `envprefix`.  
Every ENV group override element stored at `index` of slice or append new one.  
Sparse slices are not allowed.
```go
var cfg struct {
...
Replicas []Postgres `json:"replicas" envprefix:"REPLICAS"`
...
}
```
Environment key should has next pattern:  
`${envprefix}_${index}_${envconfig}` or `${envprefix}_${index}_${StructFieldName}`
```
REPLICAS_0_POSTGRES_USER=replica REPLICAS_2_USER=replica
```
#### `time.Duration`, `time.Time`
In case using json file you have to use aliases `config.Duration`, `config.Time`, that properly unmarshal it self
```go
type NATS struct {
    ...
    ReconnectInterval config.Duration `json:"reconnect_interval" envconfig:"NATS_RECONNECT_INTERVAL" default:"2s"`
}
```
Otherwise `time.Duration`, `time.Time` might be used directly:
```go
var cfg struct {
    ReadTimeout  time.Duration `envconfig:"READ_TIMEOUT"  default:"1s"`
    WriteTimeout time.Duration `envconfig:"WRITE_TIMEOUT" default:"10s"`
}
```
