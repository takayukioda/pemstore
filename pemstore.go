package pemstore

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

type awsSsmStore struct {
	ssm    *ssm.SSM
	prefix string
}

const DEFAULT_STORE_PREFIX = "pemstore"
const DEFAULT_PROFILE = "default"
const PREFIX_DELIMITER = "/"

func New(profile *string, mfaEnabled bool, prefix *string) *awsSsmStore {
	options := session.Options{
		Profile:           DEFAULT_PROFILE,
		SharedConfigState: session.SharedConfigEnable,
	}
	if profile != nil && *profile != "" {
		options.Profile = aws.StringValue(profile)
	}
	if mfaEnabled {
		options.AssumeRoleTokenProvider = stscreds.StdinTokenProvider
	}

	storePrefix := DEFAULT_STORE_PREFIX
	if prefix != nil && *prefix != "" {
		storePrefix = *prefix
	}

	sess := session.Must(session.NewSessionWithOptions(options))
	return &awsSsmStore{
		ssm:    ssm.New(sess),
		prefix: storePrefix,
	}
}

func (p awsSsmStore) storeKey(suffix string) string {
	return PREFIX_DELIMITER + p.prefix + PREFIX_DELIMITER + suffix
}

func (p awsSsmStore) listParameters(params []*ssm.ParameterMetadata, token *string, initial bool) ([]*ssm.ParameterMetadata, error) {
	if !initial && token == nil {
		return params, nil
	}
	filter := &ssm.ParameterStringFilter{
		Key:    aws.String("Name"),
		Option: aws.String("BeginsWith"),
		Values: aws.StringSlice([]string{p.storeKey("")}),
	}
	output, err := p.ssm.DescribeParameters(&ssm.DescribeParametersInput{
		NextToken:        token,
		ParameterFilters: []*ssm.ParameterStringFilter{filter},
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

func (p awsSsmStore) List() ([]string, error) {
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

func (p awsSsmStore) Get(key string, decryption bool) (string, error) {
	output, err := p.ssm.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String(p.storeKey(key)),
		WithDecryption: aws.Bool(decryption),
	})
	if err != nil {
		return "", err
	}
	return aws.StringValue(output.Parameter.Value), nil
}

func (p awsSsmStore) Exists(key string) (bool, error) {
	filter := ssm.ParameterStringFilter{
		Key:    aws.String("Name"),
		Values: aws.StringSlice([]string{p.storeKey(key)}),
	}
	output, err := p.ssm.DescribeParameters(&ssm.DescribeParametersInput{
		ParameterFilters: []*ssm.ParameterStringFilter{&filter},
	})
	if err != nil {
		return false, err
	}
	return len(output.Parameters) != 0, nil
}

func (p awsSsmStore) Store(key string, data []byte, overwrite bool) error {
	_, err := p.ssm.PutParameter(&ssm.PutParameterInput{
		Type:  aws.String("SecureString"),
		Name:  aws.String(p.storeKey(key)),
		Value: aws.String(string(data)),
	})
	return err
}

func (p awsSsmStore) Remove(key string) error {
	_, err := p.ssm.DeleteParameter(&ssm.DeleteParameterInput{
		Name: aws.String(p.storeKey(key)),
	})
	return err
}
