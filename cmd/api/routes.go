package main

import (
	"expvar"
	"github.com/julienschmidt/httprouter"
	"github.com/pistolricks/go-api-template/ui"
	"net/http"
)

func (app *application) routes() http.Handler {

	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	/* DO NOT FORGET THE *filepath for the fileserver */

	router.Handler(http.MethodGet, "/static/*filepath", http.FileServerFS(ui.Files))

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodGet, "/v1/vendors", app.listVendorsHandler)
	router.HandlerFunc(http.MethodPost, "/v1/vendors", app.requirePermission("vendors:write", app.createVendorHandler))
	router.HandlerFunc(http.MethodGet, "/v1/vendors/:id", app.requirePermission("vendors:write", app.showVendorHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/vendors/:id", app.requirePermission("vendors:write", app.updateVendorHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/vendors/:id", app.requirePermission("vendors:write", app.deleteVendorHandler))

	router.HandlerFunc(http.MethodGet, "/v1/addresses/create", app.requirePermission("vendors:write", app.showAddressForm))
	router.HandlerFunc(http.MethodPost, "/v1/addresses", app.requirePermission("vendors:write", app.createAddressHandler))

	router.HandlerFunc(http.MethodGet, "/v1/contents", app.requirePermission("vendors:write", app.listContentsHandler))
	router.HandlerFunc(http.MethodPost, "/v1/upload/image", app.requirePermission("vendors:write", app.uploadImageHandler))

	router.HandlerFunc(http.MethodGet, "/v1/users/activate", app.showActivateUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/password", app.updateUserPasswordHandler)

	router.HandlerFunc(http.MethodPost, "/v1/users/logout", app.userLogoutHandler)

	router.HandlerFunc(http.MethodPost, "/v1/tokens/activation", app.createActivationTokenHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/password-reset", app.createPasswordResetTokenHandler)

	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())
	return app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router)))))
}
