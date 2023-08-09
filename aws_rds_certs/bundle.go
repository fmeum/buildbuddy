package aws_rds_certs

import (
	"embed"
	"io/fs"

	"github.com/buildbuddy-io/buildbuddy/server/util/fileresolver"
)

// NB: Include everything in bazel `embedsrcs` with `*`.
//
//go:embed *
var all embed.FS

func Get() fs.FS {
	return fileresolver.New(all, "aws_rds_certs")
}
