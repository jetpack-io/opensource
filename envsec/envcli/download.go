// Copyright 2022 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envcli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.jetpack.io/envsec/tux"
)

type downloadOptions struct {
	format string
}

func DownloadCmd(provider configProvider) *cobra.Command {
	opts := &downloadOptions{}
	command := &cobra.Command{
		Use:   "download <file1>",
		Short: "Download environment variables into the specified file",
		Long:  "Download environment variables stored into the specified file (most commonly a .env file). The format of the file is one NAME=VALUE per line.",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.format == "json" || opts.format == "env" {
				return nil
			}
			return errors.Wrapf(errUnsupportedFormat, "format: %s", opts.format)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdCfg, err := provider(cmd.Context())
			if err != nil {
				return errors.WithStack(err)
			}
			envVars, err := cmdCfg.Store.List(cmd.Context(), cmdCfg.EnvId)
			if err != nil {
				return errors.WithStack(err)
			}

			if len(envVars) == 0 {
				err = tux.WriteHeader(cmd.OutOrStdout(),
					"[DONE] There are no environment variables to download for environment: %s\n",
					strings.ToLower(cmdCfg.EnvId.EnvName),
				)
				return errors.WithStack(err)
			}

			wd, err := os.Getwd()
			if err != nil {
				return errors.WithStack(err)
			}
			// A single relativeFilePath is guaranteed to be there.
			filePath := filepath.Join(wd, args[0] /* relativeFilePath */)

			envVarMap := map[string]string{}
			for _, envVar := range envVars {
				envVarMap[envVar.Name] = envVar.Value
			}

			var contents []byte
			if opts.format == "json" {
				contents, err = encodeToJSON(envVarMap)
			} else {
				contents, err = encodeToDotEnv(envVarMap)
			}

			if err != nil {
				return errors.WithStack(err)
			}

			err = os.WriteFile(filePath, contents, 0644)
			if err != nil {
				return errors.WithStack(err)
			}
			err = tux.WriteHeader(cmd.OutOrStdout(),
				"[DONE] Downloaded environment variables to %v for environment: %s\n",
				strings.Join(tux.QuotedTerms(args), ", "),
				strings.ToLower(cmdCfg.EnvId.EnvName),
			)
			if err != nil {
				return errors.WithStack(err)
			}

			return nil
		},
	}

	command.Flags().StringVarP(
		&opts.format, "format", "f", "env", "File format: env or json")

	return command
}

func encodeToJSON(m map[string]string) ([]byte, error) {
	b := new(bytes.Buffer)
	encoder := json.NewEncoder(b)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(m); err != nil {
		return nil, errors.WithStack(err)
	}
	return b.Bytes(), nil
}

func encodeToDotEnv(m map[string]string) ([]byte, error) {
	envContents, err := godotenv.Marshal(m)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return []byte(envContents), nil
}
