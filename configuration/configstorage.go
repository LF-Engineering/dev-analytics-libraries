package configuration

type ConfigStorage interface {
	Get(key Key) (string, error)
	Set(key Key,val string) error
}
