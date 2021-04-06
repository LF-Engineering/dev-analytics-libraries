package configurations

import (
	"github.com/LF-Engineering/dev-analytics-libraries/aws/ssm"
)

// Configuration ...
type Configuration struct {
Ssm *ssm.SSM
}

// NewClient ...
func NewProvider() (*Configuration, error) {

	s, err := ssm.NewSSMClient()
	if err != nil {
		return nil, err
	}
	config := new(Configuration)
config.Ssm = s
	return config, nil
}

// GetConfigValue ...
func (c *Configuration) GetConfigValue(name string, decryption bool) (string, error){
	param := c.Ssm.Param(name, decryption, false, "", "", "")

	val, err := param.GetValue()
	if err!= nil {
		return "", err
	}
	// val = os.Getenv(name)

	return val, nil
}


// SetConfigValue ...
func (c *Configuration) SetConfigValue(name string, dataType string, paramType string, value string) (string, error){
	param := c.Ssm.Param(name, false, false, dataType, paramType, value)

	val, err := param.SetValue()
	if err!= nil {
		return "", err
	}

	return val, nil
}

// UpdateConfigValue ...
func (c *Configuration) UpdateConfigValue(name string, dataType string, paramType string, value string, overwrite bool) (string, error){
	param := c.Ssm.Param(name, false, overwrite, dataType, paramType, value)

	val, err := param.UpdateValue()
	if err!= nil {
		return "", err
	}

	return val, nil
}
