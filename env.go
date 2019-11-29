package config

import "syscall"

var lookupEnv = syscall.Getenv
var environ = syscall.Environ
var Setenv = syscall.Setenv
var Unsetenv = syscall.Unsetenv
