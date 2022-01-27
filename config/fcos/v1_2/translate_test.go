// Copyright 2022 Red Hat, Inc
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
// limitations under the License.)

package v1_2

import (
	"fmt"
	"testing"

	baseutil "github.com/coreos/butane/base/util"
	base "github.com/coreos/butane/base/v0_3"
	"github.com/coreos/butane/config/common"
	"github.com/coreos/butane/translate"

	"github.com/coreos/ignition/v2/config/util"
	"github.com/coreos/ignition/v2/config/v3_2/types"
	"github.com/coreos/vcontext/path"
	"github.com/coreos/vcontext/report"
	"github.com/stretchr/testify/assert"
)

// TestTranslateConfig tests translating the Butane config.
func TestTranslateConfig(t *testing.T) {
	tests := []struct {
		in         Config
		out        types.Config
		exceptions []translate.Translation
		report     report.Report
	}{
		// empty config
		{
			Config{},
			types.Config{
				Ignition: types.Ignition{
					Version: "3.2.0",
				},
			},
			[]translate.Translation{
				{path.New("yaml", "version"), path.New("json", "ignition", "version")},
			},
			report.Report{},
		},
		// partition number for the `root` label is incorrect
		{
			Config{
				Config: base.Config{
					Storage: base.Storage{
						Disks: []base.Disk{
							{
								Device: "/dev/vda",
								Partitions: []base.Partition{
									{
										Label:   util.StrToPtr("root"),
										SizeMiB: util.IntToPtr(12000),
										Resize:  util.BoolToPtr(true),
									},
									{
										Label:   util.StrToPtr("var-home"),
										SizeMiB: util.IntToPtr(10240),
									},
								},
							},
						},
						Filesystems: []base.Filesystem{
							{
								Device:         "/dev/disk/by-partlabel/var-home",
								Format:         util.StrToPtr("xfs"),
								Path:           util.StrToPtr("/var/home"),
								Label:          util.StrToPtr("var-home"),
								WipeFilesystem: util.BoolToPtr(false),
							},
						},
					},
				},
			},
			types.Config{
				Ignition: types.Ignition{
					Version: "3.2.0",
				},
				Storage: types.Storage{
					Disks: []types.Disk{
						{
							Device: "/dev/vda",
							Partitions: []types.Partition{
								{
									Label:   util.StrToPtr("root"),
									SizeMiB: util.IntToPtr(12000),
									Resize:  util.BoolToPtr(true),
								},
								{
									Label:   util.StrToPtr("var-home"),
									SizeMiB: util.IntToPtr(10240),
								},
							},
						},
					},
					Filesystems: []types.Filesystem{
						{
							Device:         "/dev/disk/by-partlabel/var-home",
							Format:         util.StrToPtr("xfs"),
							Path:           util.StrToPtr("/var/home"),
							Label:          util.StrToPtr("var-home"),
							WipeFilesystem: util.BoolToPtr(false),
						},
					},
				},
			},
			[]translate.Translation{
				{path.New("yaml", "version"), path.New("json", "ignition", "version")},
				{path.New("yaml", "storage", "disks", 0, "partitions", 0, "label"), path.New("json", "storage", "disks", 0, "partitions", 0, "label")},
				{path.New("yaml", "storage", "disks", 0, "partitions", 0, "size_mib"), path.New("json", "storage", "disks", 0, "partitions", 0, "sizeMiB")},
				{path.New("yaml", "storage", "disks", 0, "partitions", 0, "resize"), path.New("json", "storage", "disks", 0, "partitions", 0, "resize")},
				{path.New("yaml", "storage", "disks", 0, "partitions", 1, "label"), path.New("json", "storage", "disks", 0, "partitions", 1, "label")},
				{path.New("yaml", "storage", "disks", 0, "partitions", 1, "size_mib"), path.New("json", "storage", "disks", 0, "partitions", 1, "sizeMiB")},
				{path.New("yaml", "storage", "disks", 0, "partitions", 0), path.New("json", "storage", "disks", 0, "partitions", 0)},
				{path.New("yaml", "storage", "disks", 0), path.New("json", "storage", "disks", 0)},
				{path.New("yaml", "storage", "filesystems", 0, "device"), path.New("json", "storage", "filesystems", 0, "device")},
				{path.New("yaml", "storage", "filesystems", 0, "format"), path.New("json", "storage", "filesystems", 0, "format")},
				{path.New("yaml", "storage", "filesystems", 0, "path"), path.New("json", "storage", "filesystems", 0, "path")},
				{path.New("yaml", "storage", "filesystems", 0, "label"), path.New("json", "storage", "filesystems", 0, "label")},
				{path.New("yaml", "storage", "filesystems", 0, "wipe_filesystem"), path.New("json", "storage", "filesystems", 0, "wipeFilesystem")},
				{path.New("yaml", "storage", "filesystems", 0), path.New("json", "storage", "filesystems", 0)},
				{path.New("yaml", "storage", "filesystems"), path.New("json", "storage", "filesystems")},
				{path.New("yaml", "storage"), path.New("json", "storage")},
			},
			report.Report{
				Entries: []report.Entry{
					{
						Kind:    report.Warn,
						Message: common.ErrWrongPartitionNumber.Error(),
						Context: path.New("json", "storage", "disks", 0, "partitions", 0, "label"),
					},
				},
			},
		},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("translate %d", i), func(t *testing.T) {
			actual, translations, r := test.in.ToIgn3_2Unvalidated(common.TranslateOptions{})
			assert.Equal(t, test.out, actual, "translation mismatch")
			assert.Equal(t, test.report, r, "report mismatch")
			baseutil.VerifyTranslations(t, translations, test.exceptions)
			assert.NoError(t, translations.DebugVerifyCoverage(actual), "incomplete TranslationSet coverage")
		})
	}
}