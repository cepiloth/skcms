// Copyright 2018 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"go.skia.org/infra/task_scheduler/go/specs"
)

var (
	TASKS = []string{
		"skcms-Linux",
		"skcms-Mac",
		"skcms-Win",
	}
)

func addTask(b *specs.TasksCfgBuilder, task string) {
	dimensions := map[string][]string{
		// It's nice to test on Skylake Xeons, which support AVX-512.
		"skcms-Linux": []string{"os:Linux", "cpu:x86-64-Skylake_GCE"},
		"skcms-Mac":   []string{"os:Mac"},
		// We think there's something amiss building on Win7 or Win8 bots, so restrict to 2016.
		"skcms-Win": []string{"os:Windows-2016Server"},
	}
	packages := map[string][]*specs.CipdPackage{
		"skcms-Linux": []*specs.CipdPackage{
			&specs.CipdPackage{
				Name:    "infra/ninja/linux-amd64",
				Path:    "ninja",
				Version: "version:1.8.2",
			},
			&specs.CipdPackage{
				Name:    "skia/bots/android_ndk_linux",
				Path:    "ndk",
				Version: "version:14",
			},
			&specs.CipdPackage{
				Name:    "skia/bots/clang_linux",
				Path:    "clang_linux",
				Version: "version:12",
			},
			&specs.CipdPackage{
				Name:    "skia/bots/mips64el_toolchain_linux",
				Path:    "mips64el_toolchain_linux",
				Version: "version:4",
			},
		},
		"skcms-Mac": []*specs.CipdPackage{
			&specs.CipdPackage{
				Name:    "infra/ninja/mac-amd64",
				Path:    "ninja",
				Version: "version:1.8.2",
			},
			&specs.CipdPackage{
				Name:    "skia/bots/android_ndk_darwin",
				Path:    "ndk",
				Version: "version:8",
			},
			// Copied from
			// https://skia.googlesource.com/skia/+/30a4e3da4bf341d5968b8cdf5bc2260e7f0d4b04/infra/bots/gen_tasks.go#206
			// https://chromium.googlesource.com/chromium/tools/build/+/e19b7d9390e2bb438b566515b141ed2b9ed2c7c2/scripts/slave/recipe_modules/ios/api.py#317
			// This package is really just an installer for XCode.
			&specs.CipdPackage{
				Name:    "infra/tools/mac_toolchain/${platform}",
				Path:    "mac_toolchain",
				Version: "git_revision:796d2b92cff93fc2059623ce0a66284373ceea0a",
			},
		},
		"skcms-Win": []*specs.CipdPackage{
			&specs.CipdPackage{
				Name:    "skia/bots/win_ninja",
				Path:    "ninja",
				Version: "version:2",
			},
			&specs.CipdPackage{
				Name:    "skia/bots/win_toolchain",
				Path:    "win_toolchain",
				Version: "version:9",
			},
			&specs.CipdPackage{
				Name:    "skia/bots/clang_win",
				Path:    "clang_win",
				Version: "version:8",
			},
		},
	}

	caches := map[string][]*specs.Cache{
		"skcms-Mac": []*specs.Cache{
			// Use a different cache from Skia so that there's less churn if we update
			// the Xcode version for one without updating the other.
			&specs.Cache{
				Name: "xcode_skcms",
				Path: "cache/Xcode_skcms.app",
			},
		},
	}

	command := []string{"python", "skcms/infra/bots/bot.py"}
	for _, p := range packages[task] {
		command = append(command, p.Path)
	}
	for _, c := range caches[task] {
		command = append(command, c.Path)
	}

	b.MustAddTask(task, &specs.TaskSpec{
		Caches:       caches[task],
		CipdPackages: packages[task],
		Command:      command,
		Dimensions:   append(dimensions[task], "gpu:none", "pool:Skia"),
		Isolate:      "bot.isolate",
		MaxAttempts:  1,
		// This service account gives mac_toolchain access to the Xcode CIPD packages.
		ServiceAccount: "skia-external-compile-tasks@skia-swarming-bots.iam.gserviceaccount.com",
	})

	b.MustAddJob(task, &specs.JobSpec{
		TaskSpecs: []string{task},
	})
}

func main() {
	b := specs.MustNewTasksCfgBuilder()
	for _, task := range TASKS {
		addTask(b, task)
	}

	b.MustAddJob("skcms", &specs.JobSpec{
		TaskSpecs: TASKS,
	})
	b.MustFinish()
}
