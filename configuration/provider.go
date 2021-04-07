package configuration

// todo: problems
// 1. mode (local or online)
// two constructors NewProvider, NewDevelopProvider
// 2. key names
// enum
// 3. config storage
// configStorage interface

// Configuration ...
type Configuration struct {
	configStorage ConfigStorage
}

// NewProvider ...
func NewProvider(configStorage ConfigStorage) *Configuration {
	return &Configuration{
		configStorage: configStorage,
	}
}

// Get ...
func (c *Configuration) Get(key Key) (string, error) {
	return c.configStorage.Get(key)
}

// Get ...
func (c *Configuration) Set(key Key, val string) error {
	return c.configStorage.Set(key, val)
}