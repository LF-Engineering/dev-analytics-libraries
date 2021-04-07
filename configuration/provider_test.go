package configuration

import "testing"

func TestGet(t *testing.T) {
	s := NewLocalConfigStorage()
	srv := NewProvider(s)

	srv.Set(EsUrl,)
}
