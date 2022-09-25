// SPDX-FileCopyrightText: 2022 Red Hat, Inc. <sd-mt-sre@redhat.com>
//
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type FailedRequestError int

func (e FailedRequestError) Error() string {
	return fmt.Sprintf("request failed with status %d", e)
}

func GetLatestTag(ctx context.Context, owner, repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("constructing request: %w", err)
	}

	var client http.Client

	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("retrieving latest tag: %w", err)
	}

	defer res.Body.Close()

	if status := res.StatusCode; status != http.StatusOK {
		return "", FailedRequestError(status)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("reading response body: %w", err)
	}

	release := struct {
		TagName string `json:"tag_name"`
	}{}

	if err := json.Unmarshal(data, &release); err != nil {
		return "", fmt.Errorf("unmarshalling response body: %w", err)
	}

	return release.TagName, nil
}

func DownloadFile(ctx context.Context, url, out string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("constructing request: %w", err)
	}

	var client http.Client

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("downloading file: %w", err)
	}

	defer res.Body.Close()

	if status := res.StatusCode; status != http.StatusOK {
		return FailedRequestError(status)
	}

	f, err := os.Create(out)
	if err != nil {
		return fmt.Errorf("creating file %q: %w", out, err)
	}

	defer f.Close()

	if _, err := io.Copy(f, res.Body); err != nil {
		return fmt.Errorf("copying response: %w", err)
	}

	return nil
}
