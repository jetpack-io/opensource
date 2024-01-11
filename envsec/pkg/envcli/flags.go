// Copyright 2023 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envcli

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.jetpack.io/envsec/internal/build"
	"go.jetpack.io/envsec/pkg/awsfed"
	"go.jetpack.io/envsec/pkg/envsec"
	"go.jetpack.io/envsec/pkg/stores/jetstore"
	"go.jetpack.io/envsec/pkg/stores/ssmstore"
	"go.jetpack.io/pkg/auth/session"
	"go.jetpack.io/pkg/envvar"
	"go.jetpack.io/pkg/id"
	"go.jetpack.io/typeid"
)

// to be composed into xyzCmdFlags structs
type configFlags struct {
	projectID string
	orgID     string
	envName   string
}

func (f *configFlags) register(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(
		&f.projectID,
		"project-id",
		"",
		"project id by which to namespace secrets",
	)

	cmd.PersistentFlags().StringVar(
		&f.orgID,
		"org-id",
		"",
		"organization id by which to namespace secrets",
	)

	cmd.PersistentFlags().StringVar(
		&f.envName,
		"environment",
		"dev",
		"environment name, one of: dev, preview, prod",
	)
}

func (f *configFlags) validateProjectID(orgID id.OrgID) (string, error) {
	if f.projectID != "" {
		return f.projectID, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", errors.WithStack(err)
	}
	config, err := (&envsec.Envsec{
		WorkingDir: wd,
		IsDev:      build.IsDev,
	}).ProjectConfig()
	if errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf(
			"project ID not specified. You must run `envsec init` or specify --project-id in this directory",
		)
	} else if err != nil {
		return "", errors.WithStack(err)
	}

	if config.OrgID != orgID {
		// Validate that the project ID belongs to the org ID
		return "", errors.Errorf(
			"Project ID %s does not belong to organization %s",
			config.ProjectID,
			orgID,
		)
	}
	return config.ProjectID.String(), nil
}

type CmdConfig struct {
	envsec   *envsec.Envsec
	envID    envsec.EnvID
	envNames []string
}

func (f *configFlags) genConfig(cmd *cobra.Command) (*CmdConfig, error) {
	if bootstrappedConfig != nil {
		return bootstrappedConfig, nil
	}

	ctx := cmd.Context()
	var tok *session.Token
	var err error

	wd, err := os.Getwd()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	envsecInstance := defaultEnvsec(cmd, wd)

	var store envsec.Store
	if envvar.Bool("ENVSEC_USE_AWS_STORE") {
		// Temporary hack to enable the AWS store
		ssmConfig, err := awsfed.GenSSMConfigFromToken(ctx, tok, true /*useCache*/)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		store, err = ssmstore.New(ctx, ssmConfig)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	} else {
		store = &jetstore.JetpackAPIStore{}
	}

	if err = envsecInstance.SetStore(ctx, store); err != nil {
		return nil, errors.WithStack(err)
	}

	if err != nil {
		return nil, errors.WithStack(err)
	}

	if tok != nil && f.orgID == "" {
		f.orgID = tok.IDClaims().OrgID
	}

	orgID, err := typeid.Parse[id.OrgID](f.orgID)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	projectID, err := f.validateProjectID(orgID)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	envid, err := envsec.NewEnvID(projectID, f.orgID, f.envName)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	envNames := []string{"dev", "prod", "preview"}
	if cmd.Flags().Changed(environmentFlagName) {
		envNames = []string{envid.EnvName}
	}

	return &CmdConfig{
		envsec:   envsecInstance,
		envID:    envid,
		envNames: envNames,
	}, nil
}

var bootstrappedConfig *CmdConfig

// BootstrapConfig is used to set the config for all commands that use genConfig
// Useful for using envsec programmatically.
func BootstrapConfig(cmdConfig *CmdConfig) {
	bootstrappedConfig = cmdConfig
}
