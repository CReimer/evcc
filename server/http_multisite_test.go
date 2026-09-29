package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/evcc-io/evcc/core"
	"github.com/evcc-io/evcc/core/session"
	"github.com/evcc-io/evcc/db"
	"github.com/evcc-io/evcc/hems/smartgrid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNamedSiteRoutes(t *testing.T) {
	srv := NewHTTPd("", nil, Customization{})
	srv.RegisterNamedSiteHandlers("office", core.NewSite())

	request := httptest.NewRequest(http.MethodPost, "/api/sites/office/residualpower/100", nil)
	response := httptest.NewRecorder()
	srv.Router().ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `100`, response.Body.String())
}

func TestThreeSiteControlRoutesStayIndependent(t *testing.T) {
	home, office, cabin := core.NewSite(), core.NewSite(), core.NewSite()
	srv := NewHTTPd("", nil, Customization{})
	srv.RegisterPrimarySiteHandlers("home", home)
	srv.RegisterNamedSiteHandlers("home", home)
	srv.RegisterNamedSiteHandlers("office", office)
	srv.RegisterNamedSiteHandlers("cabin", cabin)

	for path, value := range map[string]string{
		"/api/residualpower/100":              "100",
		"/api/sites/office/residualpower/200": "200",
		"/api/sites/cabin/residualpower/300":  "300",
	} {
		response := httptest.NewRecorder()
		srv.Router().ServeHTTP(response, httptest.NewRequest(http.MethodPost, path, nil))
		require.Equal(t, http.StatusOK, response.Code)
		assert.JSONEq(t, value, response.Body.String())
	}
	assert.Equal(t, 100.0, home.GetResidualPower())
	assert.Equal(t, 200.0, office.GetResidualPower())
	assert.Equal(t, 300.0, cabin.GetResidualPower())
}

func TestNamedSiteSessionsAreScoped(t *testing.T) {
	require.NoError(t, db.NewInstance("sqlite", ":memory:"))
	require.NoError(t, db.Instance.Create(&[]session.Session{
		{Site: "home", Loadpoint: "garage", ChargedEnergy: 1},
		{Site: "office", Loadpoint: "parking", ChargedEnergy: 2},
	}).Error)

	srv := NewHTTPd("", nil, Customization{})
	srv.RegisterPrimarySiteHandlers("home", core.NewSite())
	srv.RegisterNamedSiteHandlers("office", core.NewSite())

	for path, want := range map[string]string{
		"/api/sessions":              "garage",
		"/api/sites/office/sessions": "parking",
	} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		response := httptest.NewRecorder()
		srv.Router().ServeHTTP(response, request)
		require.Equal(t, http.StatusOK, response.Code)
		var got session.Sessions
		require.NoError(t, json.Unmarshal(response.Body.Bytes(), &got))
		require.Len(t, got, 1)
		assert.Equal(t, want, got[0].Loadpoint)
	}
}

func TestNamedGridSessionsAreScoped(t *testing.T) {
	require.NoError(t, db.NewInstance("sqlite", ":memory:"))
	require.NoError(t, db.Instance.Create(&[]smartgrid.GridSession{
		{Site: "home", Type: smartgrid.Dim},
		{Site: "office", Type: smartgrid.Curtail},
	}).Error)

	srv := NewHTTPd("", nil, Customization{})
	srv.RegisterNamedSiteHandlers("office", core.NewSite())
	response := httptest.NewRecorder()
	srv.Router().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/sites/office/gridsessions", nil))
	require.Equal(t, http.StatusOK, response.Code)
	var got smartgrid.GridSessions
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &got))
	require.Len(t, got, 1)
	assert.Equal(t, "office", got[0].Site)
}
