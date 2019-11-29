package config

type envs []string

func (e envs) set(args ...string) envs {
	if len(args)%2 == 0 {
		for i := 0; i < len(args); i = i + 2 {
			e = append(e, args[i])
			_ = Setenv(args[i], args[i+1])
		}
	}
	return e
}

func (e envs) unset() {
	for _, key := range e {
		_ = Unsetenv(key)
	}
}
