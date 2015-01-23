// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package configdrive

import (
	"reflect"
	"testing"

	"github.com/coreos/coreos-cloudinit/datasource"
	"github.com/coreos/coreos-cloudinit/datasource/test"
)

func TestFetchMetadata(t *testing.T) {
	for _, tt := range []struct {
		root  string
		files test.MockFilesystem

		metadata datasource.Metadata
	}{
		{
			root:  "/",
			files: test.MockFilesystem{"/openstack/latest/meta_data.json": `{"ignore": "me"}`},
		},
		{
			root:     "/",
			files:    test.MockFilesystem{"/openstack/latest/meta_data.json": `{"hostname": "host"}`},
			metadata: datasource.Metadata{Hostname: "host"},
		},
		{
			root:  "/media/configdrive",
			files: test.MockFilesystem{"/media/configdrive/openstack/latest/meta_data.json": `{"hostname": "host", "network_config": {"content_path": "path"}, "public_keys":{"1": "key1", "2": "key2"}}`},
			metadata: datasource.Metadata{
				Hostname:          "host",
				NetworkConfigPath: "path",
				SSHPublicKeys: map[string]string{
					"1": "key1",
					"2": "key2",
				},
			},
		},
	} {
		cd := configDrive{tt.root, tt.files.ReadFile}
		metadata, err := cd.FetchMetadata()
		if err != nil {
			t.Fatalf("bad error for %q: want %v, got %q", tt, nil, err)
		}
		if !reflect.DeepEqual(tt.metadata, metadata) {
			t.Fatalf("bad metadata for %q: want %#v, got %#v", tt, tt.metadata, metadata)
		}
	}
}

func TestFetchUserdata(t *testing.T) {
	for _, tt := range []struct {
		root  string
		files test.MockFilesystem

		userdata string
	}{
		{
			"/",
			test.MockFilesystem{},
			"",
		},
		{
			"/",
			test.MockFilesystem{"/openstack/latest/user_data": "userdata"},
			"userdata",
		},
		{
			"/media/configdrive",
			test.MockFilesystem{"/media/configdrive/openstack/latest/user_data": "userdata"},
			"userdata",
		},
	} {
		cd := configDrive{tt.root, tt.files.ReadFile}
		userdata, err := cd.FetchUserdata()
		if err != nil {
			t.Fatalf("bad error for %q: want %v, got %q", tt, nil, err)
		}
		if string(userdata) != tt.userdata {
			t.Fatalf("bad userdata for %q: want %q, got %q", tt, tt.userdata, userdata)
		}
	}
}

func TestConfigRoot(t *testing.T) {
	for _, tt := range []struct {
		root       string
		configRoot string
	}{
		{
			"/",
			"/openstack",
		},
		{
			"/media/configdrive",
			"/media/configdrive/openstack",
		},
	} {
		cd := configDrive{tt.root, nil}
		if configRoot := cd.ConfigRoot(); configRoot != tt.configRoot {
			t.Fatalf("bad config root for %q: want %q, got %q", tt, tt.configRoot, configRoot)
		}
	}
}

func TestNewDatasource(t *testing.T) {
	for _, tt := range []struct {
		root       string
		expectRoot string
	}{
		{
			root:       "",
			expectRoot: "",
		},
		{
			root:       "/media/configdrive",
			expectRoot: "/media/configdrive",
		},
	} {
		service := NewDatasource(tt.root)
		if service.root != tt.expectRoot {
			t.Fatalf("bad root (%q): want %q, got %q", tt.root, tt.expectRoot, service.root)
		}
	}
}
