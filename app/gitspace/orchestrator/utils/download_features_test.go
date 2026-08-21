// Copyright 2023 Harness, Inc.
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

package utils

import (
	"archive/tar"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnpackTarballRejectsEscapingPaths(t *testing.T) {
	tests := []struct {
		name      string
		entryName func(base string) string
	}{
		{
			name: "parent traversal",
			entryName: func(string) string {
				return "../outside.txt"
			},
		},
		{
			name: "nested parent traversal",
			entryName: func(string) string {
				return "nested/../../outside.txt"
			},
		},
		{
			name: "absolute path",
			entryName: func(base string) string {
				return filepath.Join(base, "outside.txt")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := t.TempDir()
			outputDir := filepath.Join(base, "feature")
			tarball := filepath.Join(base, "feature.tar")
			outside := filepath.Join(base, "outside.txt")

			writeTestTarball(t, tarball, tt.entryName(base), "outside")

			err := unpackTarball(tarball, outputDir)
			require.Error(t, err)
			require.Contains(t, err.Error(), "invalid tar entry path")

			_, statErr := os.Stat(outside)
			require.ErrorIs(t, statErr, os.ErrNotExist)
		})
	}
}

func TestUnpackTarballAllowsNestedPaths(t *testing.T) {
	base := t.TempDir()
	outputDir := filepath.Join(base, "feature")
	tarball := filepath.Join(base, "feature.tar")

	writeTestTarball(t, tarball, "nested/file.txt", "inside")

	require.NoError(t, unpackTarball(tarball, outputDir))
	contents, err := os.ReadFile(filepath.Join(outputDir, "nested", "file.txt"))
	require.NoError(t, err)
	require.Equal(t, "inside", string(contents))
}

func writeTestTarball(t *testing.T, tarball, entryName, contents string) {
	t.Helper()

	f, err := os.Create(tarball)
	require.NoError(t, err)

	tw := tar.NewWriter(f)
	body := []byte(contents)
	require.NoError(t, tw.WriteHeader(&tar.Header{
		Name:     entryName,
		Mode:     0o644,
		Size:     int64(len(body)),
		Typeflag: tar.TypeReg,
	}))
	_, err = tw.Write(body)
	require.NoError(t, err)
	require.NoError(t, tw.Close())
	require.NoError(t, f.Close())
}
