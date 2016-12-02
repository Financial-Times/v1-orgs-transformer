package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/service-status-go/gtg"
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
		writeJSONMessage(writer, err.Error(), http.StatusInternalServerError)
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
		writeJSONMessage(writer, err.Error(), http.StatusInternalServerError)
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
		writeJSONMessage(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}

func writeJSONMessage(w http.ResponseWriter, msg string, statusCode int) {
	w.WriteHeader(statusCode)
	fmt.Fprintln(w, fmt.Sprintf("{\"message\": \"%s\"}", msg))
}

// ADMIN HANDLERS

func (h *orgsHandler) getOrgCount(writer http.ResponseWriter, req *http.Request) {
	if !h.service.isInitialised() {
		writer.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	count, err := h.service.orgCount()
	if err != nil {
		log.Errorf("Error calling orgCount service: %s", err.Error())
		writeJSONMessage(writer, err.Error(), http.StatusInternalServerError)
	}
	fmt.Fprint(writer, count)
}

func (h *orgsHandler) getOrgIds(writer http.ResponseWriter, req *http.Request) {
	if !h.service.isInitialised() {
		writer.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	orgUUIDs, err := h.service.orgIds()
	if err != nil {
		writeJSONMessage(writer, err.Error(), http.StatusInternalServerError)
	}

	writer.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(writer)
	for _, u := range orgUUIDs {
		enc.Encode(u)
	}
}

func (h *orgsHandler) reloadOrgs(writer http.ResponseWriter, req *http.Request) {
	err := h.service.orgReload()
	if err != nil {
		writeJSONMessage(writer, err.Error(), http.StatusInternalServerError)
	}
	writeJSONMessage(writer, "Reload successful", http.StatusOK)
}

func (h *orgsHandler) HealthCheck() v1a.Check {
	return v1a.Check{
		BusinessImpact:   "Unable to respond to requests",
		Name:             "Check service has finished initilising.",
		PanicGuide:       "https://sites.google.com/a/ft.com/ft-technology-service-transition/home/run-book-library/v1-people-transformer",
		Severity:         1,
		TechnicalSummary: "Cannot serve any content as data not loaded.",
		Checker: func() (string, error) {
			if h.service.isInitialised() {
				return "Service is up and running", nil
			}
			return "Error as service initilising", errors.New("Service is initilising.")
		},
	}
}

func (h *orgsHandler) getGTG() gtg.Status {
	if h.service.isInitialised() && h.service.isDataLoaded() {
		return gtg.Status{GoodToGo: true}
	}
	return gtg.Status{GoodToGo: false}
}
