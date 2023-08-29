// Copyright 2022 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envsec

import (
	"path"
	"strings"
)

const PATH_PREFIX = "/jetpack-data/env"

func varPath(envId EnvId, varName string) string {
	return path.Join(projectPath(envId), envId.ProjectId, envId.EnvName, varName)
}

func orgPath(envId EnvId) string {
	// because of aws permissions orgPath needs to have a trailing `/`
	return path.Join(PATH_PREFIX, envId.OrgId) + "/"
}

func projectPath(envId EnvId) string {
	// because of aws permissions projectPath needs to have a trailing `/`
	return path.Join(orgPath(envId), envId.ProjectId) + "/"
}

func envPath(envId EnvId) string {
	// because of aws permissions projectPath needs to have a trailing `/`
	return path.Join(projectPath(envId), envId.EnvName) + "/"
}

func nameFromPath(path string) string {
	subpaths := strings.Split(path, "/")
	if len(subpaths) == 0 {
		return ""
	}
	return subpaths[len(subpaths)-1]
}
