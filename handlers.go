package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

type orgsHandler struct {
	service orgsService
}

func newOrgsHandler(service orgsService) orgsHandler {
	return orgsHandler{service: service}
}

func (h *orgsHandler) getOrgs(writer http.ResponseWriter, req *http.Request) {
	if !h.service.isInitialised() {
		writer.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	obj, err := h.service.getOrgs()
	if err != nil {
		log.Errorf("Error calling getOrgs service: %s", err.Error())
		writeJSONError(writer, err.Error(), http.StatusInternalServerError)
	}
	writeJSONResponse(obj, true, writer)
}

func (h *orgsHandler) getOrgByUUID(writer http.ResponseWriter, req *http.Request) {
	if !h.service.isInitialised() {
		writer.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(req)
	uuid := vars["uuid"]

	obj, found, err := h.service.getOrgByUUID(uuid)
	if err != nil {
		writeJSONError(writer, err.Error(), http.StatusInternalServerError)
	}
	writeJSONResponse(obj, found, writer)
}

func writeJSONResponse(obj interface{}, found bool, writer http.ResponseWriter) {
	writer.Header().Add("Content-Type", "application/json")

	if !found {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	enc := json.NewEncoder(writer)
	if err := enc.Encode(obj); err != nil {
		log.Errorf("Error on json encoding=%v\n", err)
		writeJSONError(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}

func writeJSONError(w http.ResponseWriter, errorMsg string, statusCode int) {
	w.WriteHeader(statusCode)
	fmt.Fprintln(w, fmt.Sprintf("{\"message\": \"%s\"}", errorMsg))
}

// ADMIN HANDLERS

func (h *orgsHandler) getOrgCount(writer http.ResponseWriter, req *http.Request) {
	if !h.service.isInitialised() {
		writer.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	obj, err := h.service.orgCount()
	if err != nil {
		log.Errorf("Error calling orgCount service: %s", err.Error())
		writeJSONError(writer, err.Error(), http.StatusInternalServerError)
	}
	writeJSONResponse(obj, true, writer)
}

func (h *orgsHandler) getOrgIds(writer http.ResponseWriter, req *http.Request) {
	if !h.service.isInitialised() {
		writer.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	orgUUIDs, err := h.service.orgIds()
	if err != nil {
		writeJSONError(writer, err.Error(), http.StatusInternalServerError)
	}

	writer.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(writer)
	for _, u := range orgUUIDs {
		enc.Encode(u)
	}
}

func (h *orgsHandler) reloadOrgs(writer http.ResponseWriter, req *http.Request) {

}
