package pemstore

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

type AwsSsmStore struct {
	ssm *ssm.SSM
}

func New(profile string, mfaEnabled bool) *AwsSsmStore {
	options := session.Options{
		Profile:           profile,
		SharedConfigState: session.SharedConfigEnable,
	}
	if mfaEnabled {
		options.AssumeRoleTokenProvider = stscreds.StdinTokenProvider
	}
	sess := session.Must(session.NewSessionWithOptions(options))
	return &AwsSsmStore{
		ssm: ssm.New(sess),
	}
}

func (p AwsSsmStore) listParameters(params []*ssm.ParameterMetadata, token *string, initial bool) ([]*ssm.ParameterMetadata, error) {
	if !initial && token == nil {
		return params, nil
	}
	output, err := p.ssm.DescribeParameters(&ssm.DescribeParametersInput{
		NextToken: token,
	})
	if err != nil {
		return nil, err
	}
	if initial {
		params = make([]*ssm.ParameterMetadata, 0, len(output.Parameters))
	}
	for i, max := 0, len(output.Parameters); i < max; i++ {
		params = append(params, output.Parameters[i])
	}
	return p.listParameters(params, output.NextToken, false)
}

func (p AwsSsmStore) List() ([]string, error) {
	params, err := p.listParameters(nil, nil, true)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(params))
	for i, l := 0, len(params); i < l; i++ {
		// for _, param := range output.Parameters {
		names = append(names, aws.StringValue(params[i].Name))
	}
	return names, nil
}

func (p AwsSsmStore) Get(key string, decryption bool) (string, error) {
	output, err := p.ssm.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String(key),
		WithDecryption: aws.Bool(decryption),
	})
	if err != nil {
		return "", err
	}
	return aws.StringValue(output.Parameter.Value), nil
}

func (p AwsSsmStore) Exists(key string) (bool, error) {
	filter := ssm.ParameterStringFilter{
		Key:    aws.String("Name"),
		Values: aws.StringSlice([]string{key}),
	}
	output, err := p.ssm.DescribeParameters(&ssm.DescribeParametersInput{
		ParameterFilters: []*ssm.ParameterStringFilter{&filter},
	})
	if err != nil {
		return false, err
	}
	return len(output.Parameters) != 0, nil
}

func (p AwsSsmStore) Store(key string, data []byte, overwrite bool) error {
	_, err := p.ssm.PutParameter(&ssm.PutParameterInput{
		Type: aws.String("SecureString"),
		Name: aws.String(key),
		Value: aws.String(string(data)),
	})
	return err
}

func (p AwsSsmStore) Remove(key string) error {
	_, err := p.ssm.DeleteParameter(&ssm.DeleteParameterInput{
		Name: aws.String(key),
	})
	return err
}
