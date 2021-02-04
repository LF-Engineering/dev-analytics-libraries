package ssm

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	log "github.com/sirupsen/logrus"
)

// GetParameter takes in the variable key and checks OS for it
// if there is no environment variable for the key, it will pull
// from Parameter Store (Encrypted)
// Variables in parameter store must all be encrypted
func GetParameter(key string) (string, error) {
	value := os.Getenv(key)
	if value != "" {
		return value, nil
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-west-2"
	}

	withDecryption := true
	sess, err := session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String(region)},
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		log.Warn("Error getting a new session: ", err)
		return "", err
	}

	ssmsvc := ssm.New(sess, aws.NewConfig().WithRegion(region))
	param, err := ssmsvc.GetParameter(&ssm.GetParameterInput{
		Name:           &key,
		WithDecryption: &withDecryption,
	})
	if err != nil {
		log.Warn("Error getting SSM parameter: ", err)
		return "", err
	}

	return *param.Parameter.Value, nil
}
