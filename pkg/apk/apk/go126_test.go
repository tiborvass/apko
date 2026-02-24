package apk

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
	"strings"
	"path/filepath"
)

func TestGo126(t *testing.T) {
	packages := []string{"go~1.26"}
	arches := []string{"aarch64"}
	repos := []string{"https://packages.wolfi.dev/os", "https://packages.cgr.dev/extras"}
	wolfiSigningKey := "https://packages.wolfi.dev/os/wolfi-signing.rsa.pub"
	extrasSigningKey := "https://packages.cgr.dev/extras/chainguard-extras.rsa.pub"
	r := Releases{
		Architectures: arches,
		LatestStable: "main",
		ReleaseBranches: []ReleaseBranch{
			{
				ReleaseBranch: "main",
				GitBranch: "main",
				Arches: arches,
				Repos: []Repo{
					{Name: "os"},
					{Name: "extras"},
				},
				Keys: map[string][]RepoKeys{
					"aarch64": {
						{URL: wolfiSigningKey},
						{URL: extrasSigningKey},
					},
					"x86_64": {
						{URL: wolfiSigningKey},
						{URL: extrasSigningKey},
					},
				},
			},
		},
	}
	b := r.GetReleaseBranch("main")

	ctx := context.TODO()
	for _, arch := range arches {
		t.Run(fmt.Sprintf("arch: %s", arch), func(t *testing.T) {
			keyURLs := b.KeysFor(arch, time.Now())
			keys := map[string][]byte{}
			for _, u := range keyURLs {
				res, err := http.Get(u)
				if err != nil {
					t.Fatalf("fetching key %q: %v", u, err)
				}
				defer res.Body.Close()
				keyBytes, err := io.ReadAll(res.Body)
				if err != nil {
					t.Fatalf("reading key %q: %v", u, err)
				}
				u = strings.ReplaceAll(u, "%40", "@")
				keys[filepath.Base(u)] = keyBytes
			}
			indexes, err := GetRepositoryIndexes(ctx, repos, keys, arch, WithHTTPClient(http.DefaultClient))
			if err != nil {
				t.Fatalf("getting repository indexes: %v", err)
			}
			pkgResolver := NewPkgResolver(ctx, indexes)
			pkgs, conflicts, err := pkgResolver.GetPackagesWithDependencies(ctx, packages, nil)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("pkgs: %v, conflicts: %v", pkgs, conflicts)
		})
	}
	
}
