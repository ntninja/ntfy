package cmd

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestNewYamlSourceFromFile(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "server.yml")
	contents := `
# Normal options
listen-https: ":10443"

# Note the underscore!
listen_http: ":1080"

# OMG this is allowed now ...
K: /some/file.pem
`
	require.Nil(t, os.WriteFile(filename, []byte(contents), 0600))

	ctx, err := newYamlSourceFromFile(filename, flagsServe)
	require.Nil(t, err)

	listenHTTPS, err := ctx.String("listen-https")
	require.Nil(t, err)
	require.Equal(t, ":10443", listenHTTPS)

	listenHTTP, err := ctx.String("listen-http") // No underscore!
	require.Nil(t, err)
	require.Equal(t, ":1080", listenHTTP)

	keyFile, err := ctx.String("key-file") // Long option!
	require.Nil(t, err)
	require.Equal(t, "/some/file.pem", keyFile)
}

func TestNewYamlSourceFromFileWithInclude(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "server.yml")
	contents := `
# Baseline option
listen-https: ":10443"

# Overshadowed by this include
include: "server-override.yml"
`
	require.Nil(t, os.WriteFile(filename, []byte(contents), 0600))

	filename = filepath.Join(t.TempDir(), "server-override.yml")
	contents = `
# Override
listen-https: ":10444"
`
	require.Nil(t, os.WriteFile(filename, []byte(contents), 0600))

	ctx, err := newYamlSourceFromFile(filename, flagsServe)
	require.Nil(t, err)

	listenHTTPS, err := ctx.String("listen-https")
	require.Nil(t, err)
	require.Equal(t, ":10444", listenHTTPS)
}


func TestNewYamlSourceFromFileWithIncludes(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "server.yml")
	contents := fmt.Sprintf(`
# Items added by includes only
include:
 - "%s"  # Using an absolute path here for coverage
 - "server-override2.yml"
`, filepath.Join(t.TempDir(), "server-override1.yml"))
	require.Nil(t, os.WriteFile(filename, []byte(contents), 0600))

	filename = filepath.Join(t.TempDir(), "server-override1.yml")
	contents = `
listen-https: ":10443"
`
	require.Nil(t, os.WriteFile(filename, []byte(contents), 0600))

	filename = filepath.Join(t.TempDir(), "server-override2.yml")
	contents = `
# Overrides previous include
listen-https: ":10444"

# Adds extra option
listen-http: ":10888"
`
	require.Nil(t, os.WriteFile(filename, []byte(contents), 0600))

	ctx, err := newYamlSourceFromFile(filename, flagsServe)
	require.Nil(t, err)

	listenHTTPS, err := ctx.String("listen-https")
	require.Nil(t, err)
	require.Equal(t, ":10444", listenHTTPS)

	listenHTTP, err := ctx.String("listen-http")
	require.Nil(t, err)
	require.Equal(t, ":10888", listenHTTP)
}
