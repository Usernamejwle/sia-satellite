package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mike76-dev/sia-satellite/modules"
)

type (
	// GatewayGET contains the fields returned by a GET call to "/gateway".
	GatewayGET struct {
		NetAddress modules.NetAddress `json:"netaddress"`
		Peers      []modules.Peer     `json:"peers"`
		Online     bool               `json:"online"`
	}

	// GatewayBlocklistPOST contains the information needed to set the Blocklist.
	// of the gateway
	GatewayBlocklistPOST struct {
		Action    string   `json:"action"`
		Addresses []string `json:"addresses"`
	}

	// GatewayBlocklistGET contains the Blocklist of the gateway.
	GatewayBlocklistGET struct {
		Blocklist []string `json:"blocklist"`
	}
)

// RegisterRoutesGateway is a helper function to register all gateway routes.
func RegisterRoutesGateway(router *httprouter.Router, g modules.Gateway, requiredPassword string) {
	router.GET("/gateway", func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		gatewayHandler(g, w, req, ps)
	})
	router.POST("/gateway/connect/:netaddress", RequirePassword(func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		gatewayConnectHandler(g, w, req, ps)
	}, requiredPassword))
	router.POST("/gateway/disconnect/:netaddress", RequirePassword(func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		gatewayDisconnectHandler(g, w, req, ps)
	}, requiredPassword))
	router.GET("/gateway/blocklist", func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		gatewayBlocklistHandlerGET(g, w, req, ps)
	})
	router.POST("/gateway/blocklist", RequirePassword(func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		gatewayBlocklistHandlerPOST(g, w, req, ps)
	}, requiredPassword))
}

// gatewayHandler handles the API call asking for the gateway status.
func gatewayHandler(gateway modules.Gateway, w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	peers := gateway.Peers()
	// nil slices are marshalled as 'null' in JSON, whereas 0-length slices are
	// marshalled as '[]'. The latter is preferred, indicating that the value
	// exists but contains no elements.
	if peers == nil {
		peers = make([]modules.Peer, 0)
	}
	WriteJSON(w, GatewayGET{gateway.Address(), peers, gateway.Online()})
}

// gatewayConnectHandler handles the API call to add a peer to the gateway.
func gatewayConnectHandler(gateway modules.Gateway, w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
	addr := modules.NetAddress(ps.ByName("netaddress"))
	err := gateway.ConnectManual(addr)
	if err != nil {
		WriteError(w, Error{err.Error()}, http.StatusBadRequest)
		return
	}

	WriteSuccess(w)
}

// gatewayDisconnectHandler handles the API call to remove a peer from the gateway.
func gatewayDisconnectHandler(gateway modules.Gateway, w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
	addr := modules.NetAddress(ps.ByName("netaddress"))
	err := gateway.DisconnectManual(addr)
	if err != nil {
		WriteError(w, Error{err.Error()}, http.StatusBadRequest)
		return
	}

	WriteSuccess(w)
}

// gatewayBlocklistHandlerGET handles the API call to get the gateway's
// blocklist.
func gatewayBlocklistHandlerGET(gateway modules.Gateway, w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	// Get Blocklist.
	blocklist, err := gateway.Blocklist()
	if err != nil {
		WriteError(w, Error{"unable to get blocklist mode: " + err.Error()}, http.StatusBadRequest)
		return
	}
	WriteJSON(w, GatewayBlocklistGET{
		Blocklist: blocklist,
	})
}

// gatewayBlocklistHandlerPOST handles the API call to modify the gateway's
// blocklist.
//
// Addresses will be passed in as an array of strings, comma separated net
// addresses.
func gatewayBlocklistHandlerPOST(gateway modules.Gateway, w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	// Parse parameters.
	var params GatewayBlocklistPOST
	err := json.NewDecoder(req.Body).Decode(&params)
	if err != nil {
		WriteError(w, Error{"invalid parameters: " + err.Error()}, http.StatusBadRequest)
		return
	}

	switch params.Action {
	case "append":
		// Check that addresses where submitted.
		if len(params.Addresses) == 0 {
			WriteError(w, Error{"no addresses submitted to append or remove"}, http.StatusBadRequest)
			return
		}
		// Add addresses to Blocklist.
		if err := gateway.AddToBlocklist(params.Addresses); err != nil {
			WriteError(w, Error{"failed to add addresses to the blocklist: " + err.Error()}, http.StatusBadRequest)
			return
		}
	case "remove":
		// Check that addresses where submitted.
		if len(params.Addresses) == 0 {
			WriteError(w, Error{"no addresses submitted to append or remove"}, http.StatusBadRequest)
			return
		}
		// Remove addresses from the Blocklist.
		if err := gateway.RemoveFromBlocklist(params.Addresses); err != nil {
			WriteError(w, Error{"failed to remove addresses from the blocklist: " + err.Error()}, http.StatusBadRequest)
			return
		}
	case "set":
		// Set Blocklist.
		if err := gateway.SetBlocklist(params.Addresses); err != nil {
			WriteError(w, Error{"failed to set the blocklist: " + err.Error()}, http.StatusBadRequest)
			return
		}
	default:
		WriteError(w, Error{"invalid parameters: " + err.Error()}, http.StatusBadRequest)
		return
	}

	WriteSuccess(w)
}
