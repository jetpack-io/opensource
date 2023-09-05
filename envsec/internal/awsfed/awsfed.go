package awsfed

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/service/cognitoidentity"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentity/types"
	"github.com/golang-jwt/jwt/v5"
	"go.jetpack.io/envsec/internal/envvar"
	"go.jetpack.io/envsec/internal/filecache"
)

const cacheKey = "awsfed"

type AWSFed struct {
	AccountId      string
	IdentityPoolId string
	Provider       string
	Region         string
}

func New() *AWSFed {
	return &AWSFed{
		AccountId:      "984256416385",
		IdentityPoolId: "us-west-2:8111c156-085b-4ac5-b94d-f823205f6261",
		Provider: envvar.Get(
			"ENVSEC_AUTH_DOMAIN",
			"accounts.jetpack.io",
		),
		Region: "us-west-2",
	}
}

func (a *AWSFed) AWSCreds(
	ctx context.Context,
	token *jwt.Token,
) (*types.Credentials, error) {
	cache := filecache.New("envsec")
	if cachedCreds, err := cache.Get(cacheKey); err == nil {
		var creds types.Credentials
		if err := json.Unmarshal(cachedCreds, &creds); err == nil {
			return &creds, nil
		}
	}

	svc := cognitoidentity.New(cognitoidentity.Options{
		Region: a.Region,
	})

	logins := map[string]string{a.Provider: token.Raw}
	getIdoutput, err := svc.GetId(
		ctx,
		&cognitoidentity.GetIdInput{
			AccountId:      &a.AccountId,
			IdentityPoolId: &a.IdentityPoolId,
			Logins:         logins,
		},
	)
	if err != nil {
		return nil, err
	}

	output, err := svc.GetCredentialsForIdentity(
		ctx,
		&cognitoidentity.GetCredentialsForIdentityInput{
			IdentityId: getIdoutput.IdentityId,
			Logins:     logins,
		},
	)
	if err != nil {
		return nil, err
	}

	if creds, err := json.Marshal(output.Credentials); err != nil {
		return nil, err
	} else if err := cache.SetT(
		cacheKey,
		creds,
		*output.Credentials.Expiration,
	); err != nil {
		return nil, err
	}

	return output.Credentials, nil
}
