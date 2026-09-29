package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/evcc-io/evcc/core"
	"github.com/evcc-io/evcc/core/site"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNamedSiteConfigRoutes(t *testing.T) {
	home, office := core.NewSite(), core.NewSite()
	home.SetTitle("Home")
	office.SetTitle("Office")

	router := mux.NewRouter()
	api := router.PathPrefix("/api/config").Subrouter()
	registerSiteConfigHandlers(api, map[string]site.API{"home": home, "office": office})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/config/sites/office", nil))
	require.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{"title":"Office","grid":"","pv":null,"battery":null,"aux":null,"ext":null,"consumer":null,"curtail":null}`, response.Body.String())

	response = httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, "/api/config/sites/office", strings.NewReader(`{"title":"Branch"}`))
	router.ServeHTTP(response, request)
	require.Equal(t, http.StatusAccepted, response.Code)
	assert.Equal(t, "Home", home.GetTitle())
	assert.Equal(t, "Branch", office.GetTitle())
}
