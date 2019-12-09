package config

import "os"

var lookupEnv = os.LookupEnv
var environ = os.Environ
var Setenv = os.Setenv
var Unsetenv = os.Unsetenv

//var ExpandEnv = os.ExpandEnv
