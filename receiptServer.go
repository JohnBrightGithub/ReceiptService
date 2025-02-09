package main

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type Item struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}
type Receipt struct {
	Retailer     string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Items        []Item `json:"items"`
	Total        string `json:"total"`
}

var receipts = make(map[string]Receipt)
var pointsAwarded = make(map[string]int)
var curID = 0

// Get users
func calcPoints(receipt Receipt) int {
	points := 0

	//one point for every alphanumeric character in the retailer name
	idLen := len(receipt.Retailer)
	for i := 0; i < idLen; i++ {
		if unicode.IsNumber(rune(receipt.Retailer[i])) {
			points++
		} else if unicode.IsLetter(rune(receipt.Retailer[i])) {
			points++
		}
	}

	//50 points if total is a round dollar amount with no cents
	// r, _ := regexp.Compile(`^\\d+\\.\\d{2}$`)
	// fmt.Println(r.MatchString(receipt.Total))
	total, _ := strconv.ParseFloat(receipt.Total, 32)
	if math.Mod(total, 1.0) == 0 {
		points += 50
	}
	if math.Mod(total, 0.25) == 0 {
		points += 25
	}

	//5 points for every two items on the receipt
	twoItems := int(len(receipt.Items) / 2)
	points += (twoItems * 5)

	//If the trimmed length of the item description is a multiple of 3,
	//multiply the price by 0.2 and round up to the nearest integer.
	//The result is the number of points earned.
	for _, item := range receipt.Items {
		trimmedLength := len(strings.TrimSpace(item.ShortDescription))
		if trimmedLength%3 == 0 {
			price, _ := strconv.ParseFloat(item.Price, 32)
			nearestInt := int(math.Ceil(0.2 * price))
			points += nearestInt
		}
	}

	//6 points if the day in the purchase date is odd
	dashIndexLast := strings.LastIndex(receipt.PurchaseDate, "-")
	dayStr := receipt.PurchaseDate[dashIndexLast+1:]
	day, _ := strconv.Atoi(dayStr)
	if day%2 == 1 {
		points += 6
	}

	//10 points if the time of purchase is after 2:00pm and before 4:00pm.
	colonIndex := strings.Index(receipt.PurchaseTime, ":")
	hourStr := receipt.PurchaseTime[:colonIndex]
	minuteStr := receipt.PurchaseTime[colonIndex+1:]
	hour, _ := strconv.Atoi(hourStr)
	minute, _ := strconv.Atoi(minuteStr)
	//assuming 2:01pm means +10 points and 4:00pm means no points
	if (hour > 14 && hour < 16) || (hour == 14 && minute > 0) {
		points += 10
	}
	return points
}

// patterns for parameters
var retailerPattern = "^[\\w\\s\\-&]+$"
var totalPattern = "^\\d+\\.\\d{2}$"
var shortDescPattern = "^[\\w\\s\\-]+$"
var pricePattern = "^\\d+\\.\\d{2}$"
var idPattern = "^\\S+$"

func validateString(parameter string, pattern string) bool {
	pat := regexp.MustCompile(pattern)
	return pat.MatchString(parameter)
}

func validateTime(timeStr string) error {
	_, err := time.Parse("15:04", timeStr)
	return err
}

func validateDate(dateStr string) error {
	_, err := time.Parse(time.DateOnly, dateStr)
	return err
}

func postReceipt(w http.ResponseWriter, r *http.Request) {
	var receipt Receipt

	//invalid because error reading json
	err := json.NewDecoder(r.Body).Decode(&receipt)
	if err != nil {
		http.Error(w, "The receipt is invalid.", http.StatusBadRequest)
		return
	}

	//invalid because required fields empty
	if receipt.Retailer == "" || receipt.PurchaseDate == "" || receipt.PurchaseTime == "" || receipt.Total == "" || len(receipt.Items) <= 0 {
		http.Error(w, "The receipt is invalid.", http.StatusBadRequest)
		return
	}

	//invalid because retailer doesn't fit pattern
	if !validateString(receipt.Retailer, retailerPattern) {
		http.Error(w, "The retailer name is invalid", http.StatusBadRequest)
		return
	}

	//invalid because total doesn't fit pattern
	if !validateString(receipt.Total, totalPattern) {
		http.Error(w, "The total is invalid", http.StatusBadRequest)
		return
	}

	//invalid because an item doesn't fit pattern
	numItems := len(receipt.Items)
	for i := 0; i < numItems; i++ {
		if !validateString(receipt.Items[i].ShortDescription, shortDescPattern) {
			http.Error(w, "The short description for an item is invalid", http.StatusBadRequest)
			return
		}
		if !validateString(receipt.Items[i].Price, pricePattern) {
			http.Error(w, "The price for an item is invalid", http.StatusBadRequest)
			return
		}
	}

	//invalid because purchase time is invalid
	err = validateTime(receipt.PurchaseTime)
	if err != nil {
		http.Error(w, "The purchase time is invalid", http.StatusBadRequest)
		return
	}

	//invalid because purchase date is invalid
	err = validateDate(receipt.PurchaseDate)
	if err != nil {
		http.Error(w, "The purchase date is invalid", http.StatusBadRequest)
		return
	}

	id := strconv.Itoa(curID)
	receipts[id] = receipt
	points := calcPoints(receipt)
	pointsAwarded[id] = points

	curID++

	//sanity check our id
	if !validateString(id, idPattern) {
		http.Error(w, "The short description for an item is invalid", http.StatusBadRequest)
		return
	}

	//Returns the ID assigned to the receipt.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": id})

}

func getReceiptPoints(w http.ResponseWriter, r *http.Request) {

	//get id from URL
	id := r.URL.Path[len("/receipts/"):]
	slashIndex := strings.Index(id, "/")
	id = id[:slashIndex]

	if !validateString(id, idPattern) {
		http.Error(w, "invalid ID", http.StatusNotFound)
		return
	}
	pointTotal, contains := pointsAwarded[id]

	if !contains {
		http.Error(w, "Receipt not found for given ID ", http.StatusNotFound)
		return
	}

	//Returns the points awarded to the receipt.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"points": pointTotal})
}

func handler(w http.ResponseWriter, r *http.Request) {
	//if process, call post method
	if r.URL.Path == "/receipts/process" && r.Method == http.MethodPost {
		postReceipt(w, r)
		return
	}

	//if getMethod, call get Receipt points
	if strings.HasPrefix(r.URL.Path, "/receipts/") && r.Method == http.MethodGet {
		getReceiptPoints(w, r)
		return
	}

	http.Error(w, "Not Found", http.StatusNotFound)
}

func main() {
	log.Println("Server running on :8080")
	http.ListenAndServe(":8080", http.HandlerFunc(handler))
}
