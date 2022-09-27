// SPDX-FileCopyrightText: 2022 Red Hat, Inc. <sd-mt-sre@redhat.com>
//
// SPDX-License-Identifier: Apache-2.0

package scm

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"

	"github.com/google/go-github/v43/github"
)

func NewGitHubClient(opts ...GitHubOption) *GitHubClient {
	var cfg GitHubConfig

	cfg.Option(opts...)

	return &GitHubClient{
		cfg: cfg,
		gh:  github.NewClient(nil),
	}
}

type GitHubClient struct {
	cfg GitHubConfig
	gh  *github.Client
}

func (c *GitHubClient) GetLatestVersion(ctx context.Context) (string, error) {
	rel, _, err := c.gh.Repositories.GetLatestRelease(ctx, c.cfg.Orginaztion, c.cfg.Repository)
	if err != nil {
		return "", fmt.Errorf("getting latest release: %w", err)
	}

	return rel.GetTagName(), nil
}

func (c *GitHubClient) GetLatestPluginBinary(ctx context.Context) ([]byte, error) {
	id, err := c.getLatestAssetID(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving latest binary id: %w", err)
	}

	rc, _, err := c.gh.Repositories.DownloadReleaseAsset(ctx, c.cfg.Orginaztion, c.cfg.Repository, id, http.DefaultClient)
	if err != nil {
		return nil, fmt.Errorf("downloading release asset: %w", err)
	}

	defer rc.Close()

	unzipped, err := gzip.NewReader(rc)
	if err != nil {
		return nil, fmt.Errorf("opening compressed data: %w", err)
	}

	archive := tar.NewReader(unzipped)

	for {
		next, err := archive.Next()
		if errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("binary not found: %w", err)
		}

		if next.Name == c.cfg.TargetBinary {
			break
		}
	}

	data, err := io.ReadAll(archive)
	if err != nil {
		return nil, fmt.Errorf("reading compressed data: %w", err)
	}

	return data, nil
}

var ErrBinaryNotFound = errors.New("binary not found")

func (c *GitHubClient) getLatestAssetID(ctx context.Context) (int64, error) {
	rel, _, err := c.gh.Repositories.GetLatestRelease(ctx, c.cfg.Orginaztion, c.cfg.Repository)
	if err != nil {
		return 0, fmt.Errorf("fetching latest release: %w", err)
	}

	for _, asset := range rel.Assets {
		if !c.matchesTargetBinary(asset.GetName()) {
			continue
		}

		return asset.GetID(), nil
	}

	return 0, ErrBinaryNotFound
}

func (c *GitHubClient) matchesTargetBinary(name string) bool {
	sfx := fmt.Sprintf("%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)

	return strings.HasPrefix(name, c.cfg.TargetBinary) && strings.HasSuffix(name, sfx)
}

type GitHubConfig struct {
	Orginaztion  string
	Repository   string
	TargetBinary string
}

func (c *GitHubConfig) Option(opts ...GitHubOption) {
	for _, opt := range opts {
		opt.ConfigureGitHub(c)
	}
}

type GitHubOption interface {
	ConfigureGitHub(*GitHubConfig)
}

type WithOrganization string

func (o WithOrganization) ConfigureGitHub(c *GitHubConfig) {
	c.Orginaztion = string(o)
}

type WithRepository string

func (r WithRepository) ConfigureGitHub(c *GitHubConfig) {
	c.Repository = string(r)
}

type WithTargetBinary string

func (b WithTargetBinary) ConfigureGitHub(c *GitHubConfig) {
	c.TargetBinary = string(b)
}
