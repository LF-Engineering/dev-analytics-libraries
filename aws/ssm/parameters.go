package ssm

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// SSM implements the SSM API interface.
type SSM struct {
	client ssm.Client
}

// NewSSMClient Returns an SSM client
func NewSSMClient() (*SSM, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil{
		return nil, err
	}

	ssmClient := &SSM{*ssm.NewFromConfig(cfg)}
	return ssmClient, nil
}

// Param struct
type Param struct {
	// Get Param fields
	Name           string
	WithDecryption bool
	ssmClient      *SSM

	// update/insert param fields
	Overwrite bool   // set to true when updating param, dont set when inserting param
	DataType  string // can be text or aws:ec2:image
	Type      string //e.g SecureString
	Value     string // value of ssm parameter
}

// Param function creates the struct for querying the ssm parameter store
// name param is used to set the ssm key
// decryption param determine whither Return decrypted or encrypted values for secure string. is ignored for String and StringList
// overwrite param determine whither to overwrite existing value or not. used only with update value.
// dataType param specify data type. valid values are: text | aws:ec2:image
// paramType param specify data type. valid values are: String | StringList | SecureString
// value param specify the actual value for parameter
func (s *SSM) Param(name string, decryption bool, overwrite bool, dataType string, paramType string, value string) *Param {
	return &Param{
		Name:           name,
		WithDecryption: decryption,
		ssmClient:      s,
		Overwrite:      overwrite,
		DataType:       dataType,
		Type:           paramType,
		Value:          value,
	}
}

// SetValue Creates new SSM parameter
func (p *Param) SetValue() (string, error) {
	ssmClient := p.ssmClient.client
	_, err := ssmClient.PutParameter(context.TODO() ,&ssm.PutParameterInput{
		Name:     &p.Name,
		DataType: &p.DataType,
		Type:     types.ParameterType(p.Type),
		Value:    &p.Value,
	})
	if err != nil {
		log.Warnf("error creating ssm parameter [%+v] : %+v\n", p.Name, err)
		return "", err
	}
	return fmt.Sprintf("successfully created ssm param: %+v", p.Name), nil
}

// UpdateValue Updates SSM parameter value
func (p *Param) UpdateValue() (string, error) {
	_, err := p.ssmClient.client.PutParameter(context.TODO(), &ssm.PutParameterInput{
		Overwrite: p.Overwrite,
		Name:      &p.Name,
		DataType:  &p.DataType,
		Value:     &p.Value,
	})
	if err != nil {
		log.Warnf("error updating ssm parameter [%+v] : %+v\n", p.Name, err)
		return "", err
	}

	return fmt.Sprintf("successfully updated ssm param: %+v", p.Name), nil
}

// GetValue Returns SSM parameter value
func (p *Param) GetValue() (string, error) {
	ssmClient := p.ssmClient.client
	parameter, err := ssmClient.GetParameter(context.TODO(), &ssm.GetParameterInput{
		Name:           &p.Name,
		WithDecryption: p.WithDecryption,
	})
	if err != nil {
		log.Warnf("error getting ssm parameter [%+v] : %+v\n", p.Name, err)
		return "", err
	}

	return *parameter.Parameter.Value, nil
}
