package ssm

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	log "github.com/sirupsen/logrus"
)

// SSM implements the SSM API interface.
type SSM struct {
	client ssmiface.SSMAPI
}

// Session returns new aws session
func Session() (*session.Session, error) {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-west-2"
	}

	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region:      aws.String(region),
			Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), ""),
		},
		SharedConfigState: session.SharedConfigEnable,
	})

	svc := session.Must(sess, err)
	return svc, err
}

// NewSSMClient Returns an SSM client
func NewSSMClient() (*SSM, error) {
	// Create AWS Session
	sess, err := Session()
	if err != nil {
		log.Warnf("error while initializing aws session: %+v ", err)
		return nil, err
	}
	ssmClient := &SSM{ssm.New(sess)}
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
	_, err := ssmClient.PutParameter(&ssm.PutParameterInput{
		Name:     &p.Name,
		DataType: &p.DataType,
		Type:     &p.Type,
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
	_, err := p.ssmClient.client.PutParameter(&ssm.PutParameterInput{
		Overwrite: &p.Overwrite,
		Name:      &p.Name,
		DataType:  &p.DataType,
		Type:      &p.Type,
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
	parameter, err := ssmClient.GetParameter(&ssm.GetParameterInput{
		Name:           &p.Name,
		WithDecryption: &p.WithDecryption,
	})
	if err != nil {
		log.Warnf("error getting ssm parameter [%+v] : %+v\n", p.Name, err)
		return "", err
	}

	return *parameter.Parameter.Value, nil
}
