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

// Session returns new aws session
/*func Session() (*session.Session, error) {
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
*/

// NewSSMClient Returns an SSM client
func NewSSMClient() (*SSM, error) {
	// Create AWS Session
	/*sess, err := Session()
	if err != nil {
		log.Warnf("error while initializing aws session: %+v ", err)
		return nil, err
	}*/

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil{
		return nil, err
	}
	//ssmClient := ssm.NewFromConfig(cfg)

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
		Type:     types.ParameterTypeSecureString,
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
