package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/service-status-go/gtg"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
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
		writeJSONMessageWithStatus(writer, err.Error(), http.StatusInternalServerError)
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
		writeJSONMessageWithStatus(writer, err.Error(), http.StatusInternalServerError)
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
		writeJSONMessageWithStatus(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}

func writeJSONMessageWithStatus(w http.ResponseWriter, msg string, statusCode int) {
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
		writeJSONMessageWithStatus(writer, err.Error(), http.StatusInternalServerError)
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
		writeJSONMessageWithStatus(writer, err.Error(), http.StatusInternalServerError)
	}

	writer.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(writer)
	for _, u := range orgUUIDs {
		enc.Encode(u)
	}
}

func (h *orgsHandler) reloadOrgs(writer http.ResponseWriter, req *http.Request) {

	go func() {
		if err := h.service.orgReload(); err != nil {
			log.Errorf("ERROR reloading cache: %v", err.Error())
		}
	}()
	writeJSONMessageWithStatus(writer, "Reloading V1 organisations", http.StatusAccepted)
}

func (h *orgsHandler) HealthCheck() fthealth.Check {
	return fthealth.Check{
		BusinessImpact:   "Unable to respond to requests",
		Name:             "Check service has finished initilising.",
		PanicGuide:       "https://sites.google.com/a/ft.com/ft-technology-service-transition/home/run-book-library/v1-people-transformer",
		Severity:         1,
		TechnicalSummary: "Cannot serve any content as data not loaded.",
		Checker:          h.serviceInitialisedChecker,
	}
}

func (h *orgsHandler) serviceInitialisedChecker() (string, error) {
	if h.service.isInitialised() {
		return "Service is up and running", nil
	}
	return "Error as service initialising", errors.New("Service is initialising")
}

func (h *orgsHandler) dataLoadedChecker() (string, error) {
	if h.service.isDataLoaded() {
		return "Data loading is completed", nil
	}
	return "Error as loading data", errors.New("Data is loading")
}

func (h *orgsHandler) GTG() gtg.Status {
	isInitialisedCheck := func() gtg.Status {
		return gtgCheck(h.serviceInitialisedChecker)
	}

	dataLoadedCheck := func() gtg.Status {
		return gtgCheck(h.dataLoadedChecker)
	}

	return gtg.FailFastParallelCheck([]gtg.StatusChecker{
		isInitialisedCheck,
		dataLoadedCheck,
	})()
}

func gtgCheck(handler func() (string, error)) gtg.Status {
	if _, err := handler(); err != nil {
		return gtg.Status{GoodToGo: false, Message: err.Error()}
	}
	return gtg.Status{GoodToGo: true}
}
