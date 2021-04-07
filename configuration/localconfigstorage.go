package configuration

import "errors"

// LocalConfigStorage ...
type LocalConfigStorage struct {
	configs map[Key]string
}

// NewLocalConfigStorage ...
func NewLocalConfigStorage() *LocalConfigStorage {
	return &LocalConfigStorage{
		configs: make(map[Key]string),
	}
}

// Get ...
func (s *LocalConfigStorage) Get(key Key) (string, error) {
	v, ok := s.configs[key]
	if !ok {
		return "", errors.New("config key not found")
	}
	return v, nil
}

// Set ...
func (s *LocalConfigStorage) Set(key Key, val string) error {
	s.configs[key]=val
	return nil
}
