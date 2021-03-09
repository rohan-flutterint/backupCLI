// Copyright 2019 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	"context"
	"fmt"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/pingcap/errors"
	"go.uber.org/zap"

	"github.com/pingcap/br/pkg/lightning/log"
)

const None = "None"

// Version information.
var (
	ReleaseVersion = None
	BuildTS        = None
	GitHash        = None
	GitBranch      = None
	GoVersion      = None
)

// GetRawInfo do what its name tells
func GetRawInfo() string {
	var info string
	info += fmt.Sprintf("Release Version: %s\n", ReleaseVersion)
	info += fmt.Sprintf("Git Commit Hash: %s\n", GitHash)
	info += fmt.Sprintf("Git Branch: %s\n", GitBranch)
	info += fmt.Sprintf("UTC Build Time: %s\n", BuildTS)
	info += fmt.Sprintf("Go Version: %s\n", GoVersion)
	return info
}

// PrintInfo prints some information of the app, like git hash, binary build time, etc.
func PrintInfo(app string, callback func()) {
	oldLevel := log.SetLevel(zap.InfoLevel)
	defer log.SetLevel(oldLevel)

	log.L().Info("Welcome to "+app,
		zap.String("Release Version", ReleaseVersion),
		zap.String("Git Commit Hash", GitHash),
		zap.String("Git Branch", GitBranch),
		zap.String("UTC Build Time", BuildTS),
		zap.String("Go Version", GoVersion),
	)

	if callback != nil {
		callback()
	}
}

// FetchPDVersion get pd version
func FetchPDVersion(ctx context.Context, tls *TLS, pdAddr string) (*semver.Version, error) {
	var rawVersion string
	err := tls.WithHost(pdAddr).GetJSON(ctx, "/pd/api/v1/config/cluster-version", &rawVersion)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return semver.NewVersion(rawVersion)
}

func ExtractTiDBVersion(version string) (*semver.Version, error) {
	// version format: "5.7.10-TiDB-v2.1.0-rc.1-7-g38c939f"
	//                               ^~~~~~~~~^ we only want this part
	// version format: "5.7.10-TiDB-v2.0.4-1-g06a0bf5"
	//                               ^~~~^
	// version format: "5.7.10-TiDB-v2.0.7"
	//                               ^~~~^
	// version format: "5.7.25-TiDB-v3.0.0-beta-211-g09beefbe0-dirty"
	//                               ^~~~~~~~~^
	// The version is generated by `git describe --tags` on the TiDB repository.
	versions := strings.Split(strings.TrimSuffix(version, "-dirty"), "-")
	end := len(versions)
	switch end {
	case 3, 4:
	case 5, 6:
		end -= 2
	default:
		return nil, errors.Errorf("not a valid TiDB version: %s", version)
	}
	rawVersion := strings.Join(versions[2:end], "-")
	rawVersion = strings.TrimPrefix(rawVersion, "v")
	return semver.NewVersion(rawVersion)
}
