package configuration

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

// Get configuration value by key
func (c *Configuration) Get(key Key) (string, error) {
	return c.configStorage.Get(key)
}

// Set configuration value
func (c *Configuration) Set(key Key, val string) error {
	return c.configStorage.Set(key, val)
}
