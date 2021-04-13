package configuration

// ConfigStorage ...
type ConfigStorage interface {
	Get(key Key) (string, error)
	Set(key Key,val string) error
}
