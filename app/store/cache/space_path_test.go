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

package cache

import (
	"context"
	"testing"

	"github.com/harness/gitness/app/store"
	"github.com/harness/gitness/types"

	"github.com/stretchr/testify/require"
)

type recordingSpacePathCache struct {
	evicted string
}

func (c *recordingSpacePathCache) Stats() (int64, int64) {
	return 0, 0
}

func (c *recordingSpacePathCache) Get(_ context.Context, _ string) (*types.SpacePath, error) {
	return nil, nil
}

func (c *recordingSpacePathCache) Evict(_ context.Context, key string) {
	c.evicted = key
}

func TestSpacePathCacheEvictNormalizesKey(t *testing.T) {
	t.Parallel()

	inner := &recordingSpacePathCache{}
	cache := spacePathCache{
		inner:                   inner,
		spacePathTransformation: store.ToLowerSpacePathTransformation,
	}

	cache.Evict(context.Background(), "Root/Child")

	require.Equal(t, "root/child", inner.evicted)
}
