// Copyright 2023 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
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
//
// SPDX-License-Identifier: Apache-2.0

//go:build integration
// +build integration

package database_test

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/require"
)

func TestSavedQueries_ListSavedQueries(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)

		savedQueriesFilter = model.QueryParameterFilter{
			Name:         "id",
			Operator:     model.GreaterThan,
			Value:        "4",
			IsStringData: false,
		}
		savedQueriesFilterMap = model.QueryParameterFilterMap{savedQueriesFilter.Name: model.QueryParameterFilters{savedQueriesFilter}}
	)

	userUUID, err := uuid.NewV4()
	require.Nil(t, err)

	for i := 0; i < 7; i++ {
		if _, err := dbInst.CreateSavedQuery(testCtx, userUUID, fmt.Sprintf("saved_query_%d", i), "", ""); err != nil {
			t.Fatalf("Error creating audit log: %v", err)
		}
	}

	if _, count, err := dbInst.ListSavedQueries(testCtx, userUUID, "", model.SQLFilter{}, 0, 10); err != nil {
		t.Fatalf("Failed to list all saved queries: %v", err)
	} else if count != 7 {
		t.Fatalf("Expected 7 saved queries to be returned")
	} else if filter, err := savedQueriesFilterMap.BuildSQLFilter(); err != nil {
		t.Fatalf("Failed to generate SQL Filter: %v", err)
		// Limit is set to 1 to verify that count is total filtered count, not response size
	} else if _, count, err = dbInst.ListSavedQueries(testCtx, userUUID, "", filter, 0, 1); err != nil {
		t.Fatalf("Failed to list filtered saved queries: %v", err)
	} else if count != 3 {
		t.Fatalf("Expected 3 saved queries to be returned")
	}
}

func TestSavedQueries_IsSavedQuerySharedToUser(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)

	user1, err := dbInst.CreateUser(testCtx, model.User{
		PrincipalName: userPrincipal,
	})
	require.NoError(t, err)

	query, err := dbInst.CreateSavedQuery(testCtx, user1.ID, "Test Query", "TESTING", "Example")
	require.NoError(t, err)

	_, err = dbInst.CreateSavedQueryPermissionToUser(testCtx, query.ID, user1.ID)
	require.NoError(t, err)

	isShared, err := dbInst.IsSavedQuerySharedToUser(testCtx, query.ID, user1.ID)
	require.NoError(t, err)
	assert.True(t, isShared)
}
