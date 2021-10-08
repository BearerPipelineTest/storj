// Copyright (C) 2020 Storj Labs, Inc.
// See LICENSE for copying information.

package admin_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"storj.io/common/testcontext"
	"storj.io/storj/private/testplanet"
	"storj.io/storj/satellite"
	"storj.io/storj/satellite/console"
)

func TestUserGet(t *testing.T) {
	testplanet.Run(t, testplanet.Config{
		SatelliteCount:   1,
		StorageNodeCount: 0,
		UplinkCount:      1,
		Reconfigure: testplanet.Reconfigure{
			Satellite: func(_ *zap.Logger, _ int, config *satellite.Config) {
				config.Admin.Address = "127.0.0.1:0"
			},
		},
	}, func(t *testing.T, ctx *testcontext.Context, planet *testplanet.Planet) {
		sat := planet.Satellites[0]
		address := sat.Admin.Admin.Listener.Addr()
		project := planet.Uplinks[0].Projects[0]

		projLimit, err := sat.DB.Console().Users().GetProjectLimit(ctx, project.Owner.ID)
		require.NoError(t, err)

		coupons, err := sat.DB.StripeCoinPayments().Coupons().ListByUserID(ctx, project.Owner.ID)
		require.NoError(t, err)

		couponsMarshaled, err := json.Marshal(coupons)
		require.NoError(t, err)

		link := "http://" + address.String() + "/api/users/" + project.Owner.Email
		expectedBody := `{` +
			fmt.Sprintf(`"user":{"id":"%s","fullName":"User uplink0_0","email":"%s","projectLimit":%d},`, project.Owner.ID, project.Owner.Email, projLimit) +
			fmt.Sprintf(`"projects":[{"id":"%s","name":"uplink0_0","description":"","ownerId":"%s"}],`, project.ID, project.Owner.ID) +
			fmt.Sprintf(`"coupons":%s}`, couponsMarshaled)

		assertReq(ctx, t, link, http.MethodGet, "", http.StatusOK, expectedBody, planet.Satellites[0].Config.Console.AuthToken)

		link = "http://" + address.String() + "/api/users/" + "user-not-exist@not-exist.test"
		body := assertReq(ctx, t, link, http.MethodGet, "", http.StatusNotFound, "", planet.Satellites[0].Config.Console.AuthToken)
		require.Contains(t, string(body), "does not exist")
	})
}

func TestUserAdd(t *testing.T) {
	testplanet.Run(t, testplanet.Config{
		SatelliteCount:   1,
		StorageNodeCount: 0,
		UplinkCount:      1,
		Reconfigure: testplanet.Reconfigure{
			Satellite: func(_ *zap.Logger, _ int, config *satellite.Config) {
				config.Admin.Address = "127.0.0.1:0"
			},
		},
	}, func(t *testing.T, ctx *testcontext.Context, planet *testplanet.Planet) {
		address := planet.Satellites[0].Admin.Admin.Listener.Addr()
		email := "alice+2@mail.test"

		body := strings.NewReader(fmt.Sprintf(`{"email":"%s","fullName":"Alice Test","password":"123a123"}`, email))
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://"+address.String()+"/api/users", body)
		require.NoError(t, err)
		req.Header.Set("Authorization", planet.Satellites[0].Config.Console.AuthToken)

		response, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, response.StatusCode)
		require.Equal(t, "application/json", response.Header.Get("Content-Type"))

		responseBody, err := ioutil.ReadAll(response.Body)
		require.NoError(t, err)
		require.NoError(t, response.Body.Close())

		var output console.User

		err = json.Unmarshal(responseBody, &output)
		require.NoError(t, err)

		user, err := planet.Satellites[0].DB.Console().Users().Get(ctx, output.ID)
		require.NoError(t, err)
		require.Equal(t, email, user.Email)
	})
}

func TestUserAdd_sameEmail(t *testing.T) {
	testplanet.Run(t, testplanet.Config{
		SatelliteCount:   1,
		StorageNodeCount: 0,
		UplinkCount:      1,
		Reconfigure: testplanet.Reconfigure{
			Satellite: func(_ *zap.Logger, _ int, config *satellite.Config) {
				config.Admin.Address = "127.0.0.1:0"
			},
		},
	}, func(t *testing.T, ctx *testcontext.Context, planet *testplanet.Planet) {
		address := planet.Satellites[0].Admin.Admin.Listener.Addr()
		email := "alice+2@mail.test"

		body := strings.NewReader(fmt.Sprintf(`{"email":"%s","fullName":"Alice Test","password":"123a123"}`, email))
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://"+address.String()+"/api/users", body)
		require.NoError(t, err)
		req.Header.Set("Authorization", planet.Satellites[0].Config.Console.AuthToken)

		response, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, response.StatusCode)
		require.Equal(t, "application/json", response.Header.Get("Content-Type"))

		responseBody, err := ioutil.ReadAll(response.Body)
		require.NoError(t, err)
		require.NoError(t, response.Body.Close())

		var output console.User

		err = json.Unmarshal(responseBody, &output)
		require.NoError(t, err)

		user, err := planet.Satellites[0].DB.Console().Users().Get(ctx, output.ID)
		require.NoError(t, err)
		require.Equal(t, email, user.Email)

		// Add same user again, this should fail
		body = strings.NewReader(fmt.Sprintf(`{"email":"%s","fullName":"Alice Test","password":"123a123"}`, email))
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, "http://"+address.String()+"/api/users", body)
		require.NoError(t, err)
		req.Header.Set("Authorization", planet.Satellites[0].Config.Console.AuthToken)

		response, err = http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusConflict, response.StatusCode)
		require.Equal(t, "application/json", response.Header.Get("Content-Type"))
		require.NoError(t, response.Body.Close())
	})
}

func TestUserUpdate(t *testing.T) {
	testplanet.Run(t, testplanet.Config{
		SatelliteCount:   1,
		StorageNodeCount: 0,
		UplinkCount:      1,
		Reconfigure: testplanet.Reconfigure{
			Satellite: func(_ *zap.Logger, _ int, config *satellite.Config) {
				config.Admin.Address = "127.0.0.1:0"
			},
		},
	}, func(t *testing.T, ctx *testcontext.Context, planet *testplanet.Planet) {
		address := planet.Satellites[0].Admin.Admin.Listener.Addr()
		user, err := planet.Satellites[0].DB.Console().Users().GetByEmail(ctx, planet.Uplinks[0].Projects[0].Owner.Email)
		require.NoError(t, err)

		t.Run("OK", func(t *testing.T) {
			// Updat user data.
			link := fmt.Sprintf("http://"+address.String()+"/api/users/%s", user.Email)
			body := `{"email":"alice+2@mail.test", "shortName":"Newbie"}`
			responseBody := assertReq(ctx, t, link, http.MethodPut, body, http.StatusOK, "", planet.Satellites[0].Config.Console.AuthToken)
			require.Len(t, responseBody, 0)

			updatedUser, err := planet.Satellites[0].DB.Console().Users().Get(ctx, user.ID)
			require.NoError(t, err)
			require.Equal(t, "alice+2@mail.test", updatedUser.Email)
			require.Equal(t, user.FullName, updatedUser.FullName)
			require.NotEqual(t, "Newbie", user.ShortName)
			require.Equal(t, "Newbie", updatedUser.ShortName)
			require.Equal(t, user.ID, updatedUser.ID)
			require.Equal(t, user.Status, updatedUser.Status)
			require.Equal(t, user.ProjectLimit, updatedUser.ProjectLimit)

			// Update rate limit.
			link = "http://" + address.String() + "/api/users/alice+2@mail.test"
			newLimit := 50
			body = fmt.Sprintf(`{"projectLimit":%d}`, newLimit)
			responseBody = assertReq(ctx, t, link, http.MethodPut, body, http.StatusOK, "", planet.Satellites[0].Config.Console.AuthToken)
			require.Len(t, responseBody, 0)

			updatedUserRate, err := planet.Satellites[0].DB.Console().Users().Get(ctx, user.ID)
			require.NoError(t, err)
			require.Equal(t, updatedUser.Email, updatedUserRate.Email)
			require.Equal(t, updatedUser.ID, updatedUserRate.ID)
			require.Equal(t, updatedUser.Status, updatedUserRate.Status)
			require.Equal(t, newLimit, updatedUserRate.ProjectLimit)
		})

		t.Run("Not found", func(t *testing.T) {
			link := "http://" + address.String() + "/api/users/user-not-exists@not-exists.test"
			body := `{"email":"alice+2@mail.test", "shortName":"Newbie"}`
			responseBody := assertReq(ctx, t, link, http.MethodPut, body, http.StatusNotFound, "", planet.Satellites[0].Config.Console.AuthToken)
			require.Contains(t, string(responseBody), "does not exist")
		})
	})
}

func TestUserDelete(t *testing.T) {
	testplanet.Run(t, testplanet.Config{
		SatelliteCount:   1,
		StorageNodeCount: 0,
		UplinkCount:      1,
		Reconfigure: testplanet.Reconfigure{
			Satellite: func(_ *zap.Logger, _ int, config *satellite.Config) {
				config.Admin.Address = "127.0.0.1:0"
			},
		},
	}, func(t *testing.T, ctx *testcontext.Context, planet *testplanet.Planet) {
		address := planet.Satellites[0].Admin.Admin.Listener.Addr()
		user, err := planet.Satellites[0].DB.Console().Users().GetByEmail(ctx, planet.Uplinks[0].Projects[0].Owner.Email)
		require.NoError(t, err)

		// Deleting the user should fail, as project exists
		link := fmt.Sprintf("http://"+address.String()+"/api/users/%s", user.Email)
		body := assertReq(ctx, t, link, http.MethodDelete, "", http.StatusConflict, "", planet.Satellites[0].Config.Console.AuthToken)
		require.Greater(t, len(body), 0)

		err = planet.Satellites[0].DB.Console().Projects().Delete(ctx, planet.Uplinks[0].Projects[0].ID)
		require.NoError(t, err)

		// Deleting the user should pass, as no project exists for given user.
		body = assertReq(ctx, t, link, http.MethodDelete, "", http.StatusOK, "", planet.Satellites[0].Config.Console.AuthToken)
		require.Len(t, body, 0)

		// Deleting unexisting user returns Not Found.
		body = assertReq(ctx, t, link, http.MethodDelete, "", http.StatusNotFound, "", planet.Satellites[0].Config.Console.AuthToken)
		require.Contains(t, string(body), "does not exist")
	})
}
