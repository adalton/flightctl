package service

import (
	"encoding/json"
	"net/http"

	api "github.com/flightctl/flightctl/api/v1beta1"
)

// RenewDeviceCertificate handles certificate renewal requests from devices
func (h *ServiceHandler) RenewDeviceCertificate(w http.ResponseWriter, r *http.Request, deviceName string) {
	// Parse the renewal request from the request body
	var renewalReq api.CertificateRenewalRequest
	if err := json.NewDecoder(r.Body).Decode(&renewalReq); err != nil {
		h.log.Errorf("Failed to decode renewal request: %v", err)
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Get the client certificate from the TLS connection for authentication
	clientCert := getClientCertFromRequest(r)
	if clientCert == nil {
		h.log.Warn("No client certificate provided for renewal")
		writeErrorResponse(w, http.StatusUnauthorized, "Client certificate required")
		return
	}

	// Note: The actual integration with certrotation.Handler will be done as part of proper
	// service initialization. For now, we acknowledge the request but return not implemented.
	_ = renewalReq
	_ = deviceName
	_ = clientCert

	// TODO: Initialize the certificate rotation handler properly
	// For now, this is a placeholder that needs to be wired up with the actual handler
	// when the service is initialized. This should be done as part of T021.
	h.log.Warnf("Certificate renewal endpoint called for device %s, but handler not yet initialized", deviceName)
	writeErrorResponse(w, http.StatusNotImplemented, "Certificate renewal not yet fully implemented")

	// Placeholder for actual implementation:
	// handler := h.certRotationHandler
	// response, statusCode, err := handler.HandleRenewalRequest(r.Context(), deviceName, serviceReq, clientCert)
	// if err != nil {
	//     h.log.Errorf("Certificate renewal failed: %v", err)
	//     writeErrorResponse(w, statusCode, err.Error())
	//     return
	// }
	// writeJSONResponse(w, http.StatusOK, response)
}

// writeErrorResponse writes an error response to the client
func writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(api.Status{
		Code:    int32(statusCode),
		Message: message,
	})
}

// writeJSONResponse writes a JSON response to the client
func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

// getClientCertFromRequest extracts the client certificate from the TLS connection
func getClientCertFromRequest(r *http.Request) *http.Request {
	// This needs proper implementation based on how FlightCTL handles mTLS
	// For now returning nil to indicate not implemented
	return nil
}
