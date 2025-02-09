package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Test calcPoints function
func TestCalcPoints(t *testing.T) {
	receipt := Receipt{Retailer: "M&M Corner Market", PurchaseDate: "2022-03-20", PurchaseTime: "14:33",
		Items: []Item{{"Gatorade", "2.25"}, {"Gatorade", "2.25"}, {"Gatorade", "2.25"}, {"Gatorade", "2.25"}}, Total: "9.00"}
	points := calcPoints(receipt)
	if points != 109 {
		t.Errorf("TestCalcPoints expected 109, got %d", points)
	}
	receipt2 := Receipt{Retailer: "Target", PurchaseDate: "2022-01-01", PurchaseTime: "13:01",
		Items: []Item{{"Mountain Dew 12PK", "6.49"}, {"Emils Cheese Pizza", "12.25"}, {"Knorr Creamy Chicken", "1.26"},
			{"Doritos Nacho Cheese", "3.35"}, {"   Klarbrunn 12-PK 12 FL OZ  ", "12.00"}}, Total: "35.35"}
	points2 := calcPoints(receipt2)
	if points2 != 28 {
		t.Errorf("TestCalcPoints expected 28, got %d", points)
	}
}
func TestGetReceiptPointsInvalidID(t *testing.T) {
	req, err := http.NewRequest("GET", "/receipts/invalid-id/points", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("TestGetReceiptPointsInvalidID expected status 404, got %d", status)
	}
}
func TestGetReceiptPointsHappyPath(t *testing.T) {
	receipt := Receipt{Retailer: "Store", PurchaseDate: "2022-01-01", PurchaseTime: "14:00",
		Items: []Item{{ShortDescription: "Mountain Dew 12PK", Price: "6.49"}}, Total: "10.00"}
	body, _ := json.Marshal(receipt)
	req, _ := http.NewRequest("POST", "/receipts/process", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("TestGetReceiptPointsHappyPath: expected status 200, got %d", rr.Code)
	}
	var id map[string]string
	json.NewDecoder(rr.Body).Decode(&id)
	idName := id["id"]
	getString := "/receipts/" + idName + "/points"
	req, err := http.NewRequest("GET", getString, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr2 := httptest.NewRecorder()
	handler(rr2, req)

	if rr2.Code != http.StatusOK {
		t.Errorf("TestGetReceiptPointsHappyPath: expected status 200, got %d", rr2.Code)
	}
}

func TestCalcPointsAllRules(t *testing.T) {
	receipt := Receipt{Retailer: "Store12", PurchaseDate: "2024-02-07", PurchaseTime: "14:15",
		Items: []Item{{"ABC", "3.00"}, {"ABC", "4.00"}}, Total: "10.00"}
	expected := 105
	result := calcPoints(receipt)
	if result != expected {
		t.Errorf("TestCalcPointsAllRules: expected %d, got %d", expected, result)
	}
}

func TestPostReceiptHappyPath(t *testing.T) {
	receipt := Receipt{Retailer: "Store", PurchaseDate: "2022-01-01", PurchaseTime: "14:00",
		Items: []Item{{ShortDescription: "Mountain Dew 12PK", Price: "6.49"}}, Total: "10.00"}
	body, _ := json.Marshal(receipt)
	req, _ := http.NewRequest("POST", "/receipts/process", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("TestPostReceiptHappyPath: expected status 200, got %d", rr.Code)
	}
}

func TestPostReceiptBlank(t *testing.T) {
	invalidReceipt := Receipt{}

	body, _ := json.Marshal(invalidReceipt)
	req, err := http.NewRequest("POST", "/receipts/process", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("TestPostReceiptBlank Expected status 400, got %d", status)
	}
}

func TestInvalidReceipts(t *testing.T) {
	invalidReceipts := []Receipt{
		{Retailer: "Invalid@", PurchaseDate: "2022-01-01", PurchaseTime: "14:00",
			Items: []Item{{ShortDescription: "Mountain Dew 12PK", Price: "6.49"}}, Total: "10.00"},
		{Retailer: "Store", PurchaseDate: "01-01-2022", PurchaseTime: "14:00",
			Items: []Item{{ShortDescription: "Mountain Dew 12PK", Price: "6.49"}}, Total: "10.00"},
		{Retailer: "Store", PurchaseDate: "2022-01-01", PurchaseTime: "1400",
			Items: []Item{{ShortDescription: "Mountain Dew 12PK", Price: "6.49"}}, Total: "10.00"},
		{Retailer: "Store", PurchaseDate: "2022-01-01", PurchaseTime: "14:00",
			Items: []Item{{ShortDescription: "Mountain@ Dew 12PK", Price: "6.49"}}, Total: "10.00"},
		{Retailer: "Store", PurchaseDate: "2022-01-01", PurchaseTime: "14:00",
			Items: []Item{{ShortDescription: "Mountain Dew 12PK", Price: "649"}}, Total: "10.00"},
		{Retailer: "Store", PurchaseDate: "2022-01-01", PurchaseTime: "14:00",
			Items: []Item{{ShortDescription: "Mountain Dew 12PK", Price: "6.49"}}, Total: "1000"},
	}

	for _, tt := range invalidReceipts {
		body, _ := json.Marshal(tt)
		req, _ := http.NewRequest("POST", "/receipts/process", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("TestInvalidReceipts status 400, got %d", status)
		}
	}
}
