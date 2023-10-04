package awsfed

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cognitoidentity"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentity/types"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"go.jetpack.io/envsec"
	"go.jetpack.io/envsec/internal/envvar"
	"go.jetpack.io/envsec/internal/filecache"
	"go.jetpack.io/pkg/sandbox/auth/session"
)

const cacheKeyPrefix = "awsfed"

type AWSFed struct {
	AccountID      string
	IdentityPoolID string
	LegacyProvider string
	Provider       string
	Region         string
}

func New() *AWSFed {
	return &AWSFed{
		AccountID:      "984256416385",
		IdentityPoolID: "us-west-2:8111c156-085b-4ac5-b94d-f823205f6261",
		LegacyProvider: "auth.jetpack.io",
		Provider: envvar.Get(
			"ENVSEC_AUTH_DOMAIN",
			"accounts.jetpack.io",
		),
		Region: "us-west-2",
	}
}

func (a *AWSFed) AWSCreds(
	ctx context.Context,
	tok *session.Token,
) (*types.Credentials, error) {
	cache := filecache.New("jetpack.io/envsec")
	if cachedCreds, err := cache.Get(cacheKey(tok)); err == nil {
		var creds types.Credentials
		if err := json.Unmarshal(cachedCreds, &creds); err == nil {
			return &creds, nil
		}
	}

	svc := cognitoidentity.New(cognitoidentity.Options{
		Region: a.Region,
	})

	logins := map[string]string{}
	if tok.IDClaims() == nil {
		// skip
	} else if tok.IDClaims().Issuer == fmt.Sprintf("https://%s/", a.LegacyProvider) {
		logins[a.LegacyProvider] = tok.IDToken
	} else {
		logins[a.Provider] = tok.IDToken
	}

	getIdoutput, err := svc.GetId(
		ctx,
		&cognitoidentity.GetIdInput{
			AccountId:      &a.AccountID,
			IdentityPoolId: &a.IdentityPoolID,
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
		cacheKey(tok),
		creds,
		*output.Credentials.Expiration,
	); err != nil {
		return nil, err
	}

	return output.Credentials, nil
}

// AWSCredsFromIdToken behaves similar to AWSCreds but it takes JWT from input
// rather than reading from a file or cache. This is to allow web services use
// this package without having to write every user's JWT in a cache or a file.
func (a *AWSFed) AWSCredsFromIdToken(
	ctx context.Context,
	idToken *jwt.Token,
) (*types.Credentials, error) {

	svc := cognitoidentity.New(cognitoidentity.Options{
		Region: a.Region,
	})

	logins := map[string]string{
		a.Provider: idToken.Raw,
	}

	getIdoutput, err := svc.GetId(
		ctx,
		&cognitoidentity.GetIdInput{
			AccountId:      &a.AccountID,
			IdentityPoolId: &a.IdentityPoolID,
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

	return output.Credentials, nil
}

func cacheKey(t *session.Token) string {
	id := ""
	if claims := t.IDClaims(); claims != nil && claims.OrgID != "" {
		id = claims.OrgID
	} else {
		id = fmt.Sprintf("%x", sha256.Sum256([]byte(t.IDToken)))
	}

	return fmt.Sprintf("%s-%s", cacheKeyPrefix, id)
}

func GenSSMConfigForUser(
	ctx context.Context,
	tok *session.Token,
) (*envsec.SSMConfig, error) {
	if tok == nil {
		return &envsec.SSMConfig{}, nil
	}
	fed := New()
	creds, err := fed.AWSCreds(ctx, tok)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &envsec.SSMConfig{
		AccessKeyID:     *creds.AccessKeyId,
		SecretAccessKey: *creds.SecretKey,
		SessionToken:    *creds.SessionToken,
		Region:          fed.Region,
	}, nil
}

func GenSSMConfigFromIdToken(
	ctx context.Context,
	tok *jwt.Token,
) (*envsec.SSMConfig, error) {
	if tok == nil {
		return &envsec.SSMConfig{}, nil
	}
	fed := New()
	creds, err := fed.AWSCredsFromIdToken(ctx, tok)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &envsec.SSMConfig{
		AccessKeyID:     *creds.AccessKeyId,
		SecretAccessKey: *creds.SecretKey,
		SessionToken:    *creds.SessionToken,
		Region:          fed.Region,
	}, nil
}
