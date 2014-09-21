package config

type EnvFile struct {
	Vars map[string]string
	// mask File.Content, it shouldn't be used.
	Content interface{} `json:"-" yaml:"-"`
	*File
}
