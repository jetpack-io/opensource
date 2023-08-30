// Copyright 2022 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envsec

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"go.jetpack.io/envsec/internal/debug"
)

type SSMStore struct {
	store *parameterStore
}

// SSMStore implements interface Store (compile-time check)
var _ Store = (*SSMStore)(nil)

func newSSMStore(ctx context.Context, config *SSMConfig) (*SSMStore, error) {
	paramStore, err := newParameterStore(ctx, config)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	store := &SSMStore{
		store: paramStore,
	}
	return store, nil
}

func (s *SSMStore) List(ctx context.Context, envId EnvId) ([]EnvVar, error) {
	return s.store.ListByTags(ctx, envId)
}

func (s *SSMStore) Get(ctx context.Context, envId EnvId, name string) (string, error) {
	vars, err := s.GetAll(ctx, envId, []string{name})
	if err != nil {
		return "", errors.WithStack(err)
	}
	if len(vars) == 0 {
		return "", nil
	}
	return vars[0].Value, nil
}

func (s *SSMStore) GetAll(ctx context.Context, envId EnvId, names []string) ([]EnvVar, error) {
	return s.store.getAll(ctx, envId, names)
}

func (s *SSMStore) Set(
	ctx context.Context,
	envId EnvId,
	name string,
	value string,
) error {
	path := varPath(envId, name)

	// New parameter definition
	tags := buildTags(envId, name)
	parameter := &parameter{
		tags: tags,
		id:   path,
	}
	return s.store.newParameter(ctx, parameter, value)
}

func (s *SSMStore) SetAll(ctx context.Context, envId EnvId, values map[string]string) error {
	// For now we implement by issuing multiple calls to Set()
	// Make more efficient either by implementing a batch call to the underlying API, or
	// by concurrently calling Set()

	var multiErr error
	for name, value := range values {
		err := s.Set(ctx, envId, name, value)
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
	}
	return multiErr
}

func (s *SSMStore) Delete(ctx context.Context, envId EnvId, name string) error {
	return s.DeleteAll(ctx, envId, []string{name})
}

func (s *SSMStore) DeleteAll(ctx context.Context, envId EnvId, names []string) error {
	return s.store.deleteAll(ctx, envId, names)
}

func buildFilters(envId EnvId) []types.ParameterStringFilter {
	filters := []types.ParameterStringFilter{
		{
			Key:    lo.ToPtr("Path"),
			Option: lo.ToPtr("Recursive"),
			Values: []string{projectPath(envId)},
		},
	}
	if envId.ProjectId != "" {
		filters = append(filters, types.ParameterStringFilter{
			Key:    lo.ToPtr("tag:project-id"),
			Values: []string{envId.ProjectId},
		})
	}
	if envId.OrgId != "" {
		filters = append(filters, types.ParameterStringFilter{
			Key:    lo.ToPtr("tag:org-id"),
			Values: []string{envId.OrgId},
		})
	}
	if envId.EnvName != "" {
		filters = append(filters, types.ParameterStringFilter{
			Key:    lo.ToPtr("tag:env-name"),
			Values: []string{strings.ToLower(envId.EnvName)},
		})
	}
	debug.Log("filters: %v\n\n", filters)
	return filters
}
func buildTags(envId EnvId, varName string) []types.Tag {
	tags := []types.Tag{}
	if envId.ProjectId != "" {
		tags = append(tags, types.Tag{
			Key:   lo.ToPtr("project-id"),
			Value: lo.ToPtr(envId.ProjectId),
		})
	}
	if envId.OrgId != "" {
		tags = append(tags, types.Tag{
			Key:   lo.ToPtr("org-id"),
			Value: lo.ToPtr(envId.OrgId),
		})
	}
	if envId.EnvName != "" {
		tags = append(tags, types.Tag{
			Key:   lo.ToPtr("env-name"),
			Value: lo.ToPtr(strings.ToLower(envId.EnvName)),
		})
	}

	if varName != "" {
		tags = append(tags, types.Tag{
			Key:   lo.ToPtr("name"),
			Value: lo.ToPtr(varName),
		})
	}

	return tags
}
