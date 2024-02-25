package handler

import (
	"fmt"
	"html/template"
	"log"
	"myapp/back/service"
	"net/http"
	"strconv"
)

type DatabaseCheck struct {
	databaseService *service.DS
}

func NewDatabaseCheck(databaseService *service.DS) *DatabaseCheck {
	return &DatabaseCheck{
		databaseService: databaseService,
	}
}

func (h *DatabaseCheck) Index(w http.ResponseWriter, r *http.Request) {
	// JSON parse
	tmpl, err := template.ParseFiles("l0/index.html") //HTML
	if err != nil {
		http.Error(w, "Parse error", http.StatusInternalServerError)
		log.Println("Parse erro in template:", err)
		return
	}

	// Display template
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "display error", http.StatusInternalServerError)
		log.Println("display error template:", err)
		return
	}
}

func (h *DatabaseCheck) GetInfo(w http.ResponseWriter, r *http.Request) {
	numberStr := r.FormValue("number")
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		http.Error(w, "Invalid number", http.StatusBadRequest)
		return
	}

	json, err := h.databaseService.GetInfo(number)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// JSON parse
	tmpl, err := template.ParseFiles("L0/index.html")
	if err != nil {
		http.Error(w, "Parse error", http.StatusInternalServerError)
		log.Println("Parse error template:", err)
		return
	}

	// Info
	data := struct {
		Info   string
		Number int
	}{
		Info:   json,
		Number: number,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Display error", http.StatusInternalServerError)
		log.Println("Display error template:", err)
		return
	}
}

func (h *DatabaseCheck) AddData(w http.ResponseWriter, r *http.Request) {
	// Get Info
	json := r.FormValue("json_data")

	// Call method
	id, err := h.databaseService.AddData(json)
	if err != nil {
		http.Error(w, "Add data error", http.StatusInternalServerError)
		log.Printf("Add data error: %v", err)
		return
	}
	fmt.Fprintf(w, "Added new data. ID: %d", id)
}