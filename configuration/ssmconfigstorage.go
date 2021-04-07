package configuration

import "github.com/LF-Engineering/dev-analytics-libraries/aws/ssm"

type SSMConfigStorage struct {
	ssmClient *ssm.SSM
}

// NewSSMConfigStorage ...
func NewSSMConfigStorage() (*SSMConfigStorage, error) {
	s, err := ssm.NewSSMClient()
	if err != nil {
		return nil, err
	}

	return &SSMConfigStorage{
		ssmClient: s,
	}, nil
}

// Get ...
func (s *SSMConfigStorage) Get(key Key) (string, error) {
	return s.ssmClient.Param(string(key), true, false, "", "", "").GetValue()
}

// Set ...
func (s *SSMConfigStorage) Set(key Key, val string) error {
	_, err := s.ssmClient.Param(string(key), true, true, "", "", val).UpdateValue()
	return err
}
