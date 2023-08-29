// Copyright 2023 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envcli

import (
	"context"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.jetpack.io/envsec"
	"go.jetpack.io/envsec/internal/awsfed"
)

// to be composed into xyzCmdFlags structs
type configFlags struct {
	projectId string
	envName   string
}

func (f *configFlags) register(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(
		&f.projectId,
		"project-id",
		"",
		"Project id to namespace secrets by",
	)

	// cmd.PersistentFlags().StringVar(
	// 	&f.orgId,
	// 	"org-id",
	// 	"",
	// 	"Organization id to namespace secrets by",
	// )

	cmd.PersistentFlags().StringVar(
		&f.envName,
		"environment",
		"dev",
		"Environment name, such as dev or prod",
	)
}

type cmdConfig struct {
	Store envsec.Store
	EnvId envsec.EnvId
}

func (f *configFlags) genConfig(ctx context.Context) (*cmdConfig, error) {
	user, err := newAuthenticator().GetUser()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	awsFederated := awsfed.NewAWSFed()
	ssmConfig, err := awsFederated.GetSSMConfig(user.GetAccessToken())
	if err != nil {
		return nil, errors.WithStack(err)
	}
	s, err := envsec.NewStore(ctx, ssmConfig)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	envid, err := envsec.NewEnvId(user.GetOrgId(), f.envName)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &cmdConfig{
		Store: s,
		EnvId: envid,
	}, nil
}
