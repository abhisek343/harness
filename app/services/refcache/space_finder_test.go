// Copyright 2026 Harness, Inc.
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

package refcache

import (
	"context"
	"errors"
	"testing"

	storecache "github.com/harness/gitness/app/store/cache"
	gitness_store "github.com/harness/gitness/store"
	"github.com/harness/gitness/types"

	"github.com/stretchr/testify/require"
)

type testSpacePathCache struct {
	values  []*types.SpacePath
	gets    int
	evicted []string
}

func (c *testSpacePathCache) Stats() (int64, int64) {
	return 0, 0
}

func (c *testSpacePathCache) Get(_ context.Context, _ string) (*types.SpacePath, error) {
	value := c.values[c.gets]
	c.gets++
	return value, nil
}

func (c *testSpacePathCache) Evict(_ context.Context, key string) {
	c.evicted = append(c.evicted, key)
}

type testSpaceIDCache struct {
	spaces map[int64]*types.SpaceCore
	errs   map[int64]error
}

func (c *testSpaceIDCache) Stats() (int64, int64) {
	return 0, 0
}

func (c *testSpaceIDCache) Get(_ context.Context, key int64) (*types.SpaceCore, error) {
	if err := c.errs[key]; err != nil {
		return nil, err
	}
	return c.spaces[key], nil
}

func (c *testSpaceIDCache) Evict(_ context.Context, _ int64) {}

func TestSpaceFinderFindByRefRefreshesStalePathMapping(t *testing.T) {
	t.Parallel()

	pathCache := &testSpacePathCache{
		values: []*types.SpacePath{
			{Value: "project", SpaceID: 1},
			{Value: "project", SpaceID: 2},
		},
	}
	idCache := &testSpaceIDCache{
		spaces: map[int64]*types.SpaceCore{
			2: {ID: 2, Path: "project"},
		},
		errs: map[int64]error{
			1: gitness_store.ErrResourceNotFound,
		},
	}

	finder := NewSpaceFinder(
		idCache,
		pathCache,
		nil,
		storecache.Evictor[*types.SpaceCore]{},
	)

	space, err := finder.FindByRef(context.Background(), "project")

	require.NoError(t, err)
	require.Equal(t, int64(2), space.ID)
	require.Equal(t, 2, pathCache.gets)
	require.Equal(t, []string{"project"}, pathCache.evicted)
}

func TestSpaceFinderFindByRefDoesNotRetryNonNotFoundError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("cache unavailable")
	pathCache := &testSpacePathCache{
		values: []*types.SpacePath{{Value: "project", SpaceID: 1}},
	}
	idCache := &testSpaceIDCache{
		spaces: map[int64]*types.SpaceCore{},
		errs: map[int64]error{
			1: expectedErr,
		},
	}

	finder := NewSpaceFinder(
		idCache,
		pathCache,
		nil,
		storecache.Evictor[*types.SpaceCore]{},
	)

	space, err := finder.FindByRef(context.Background(), "project")

	require.Nil(t, space)
	require.ErrorIs(t, err, expectedErr)
	require.Equal(t, 1, pathCache.gets)
	require.Empty(t, pathCache.evicted)
}
