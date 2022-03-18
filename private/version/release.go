// Copyright (C) 2020 Storj Labs, Inc.
// See LICENSE for copying information.

package version

import _ "unsafe" // needed for go:linkname

//go:linkname buildTimestamp storj.io/private/version.buildTimestamp
var buildTimestamp string = "1647619366"

//go:linkname buildCommitHash storj.io/private/version.buildCommitHash
var buildCommitHash string = "5e5e2ba0e39a0e283893112dd3855d8f4dd3a444"

//go:linkname buildVersion storj.io/private/version.buildVersion
var buildVersion string = "v1.50.4"

//go:linkname buildRelease storj.io/private/version.buildRelease
var buildRelease string = "true"

// ensure that linter understands that the variables are being used.
func init() { use(buildTimestamp, buildCommitHash, buildVersion, buildRelease) }

func use(...interface{}) {}
