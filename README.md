# ReceiptService
 receipt server can be run with
 go run receiptServer.go

Confirm working with 
curl -X POST http://localhost:8080/receipts/process -H "Content-Type: application/json" -d '{
  "retailer": "M&M Corner Market",
  "purchaseDate": "2022-03-20",
  "purchaseTime": "14:33",
  "items": [
    {
      "shortDescription": "Gatorade",
      "price": "2.25"
    },{
      "shortDescription": "Gatorade",
      "price": "2.25"
    },{
      "shortDescription": "Gatorade",
      "price": "2.25"
    },{
      "shortDescription": "Gatorade",
      "price": "2.25"
    }
  ],
  "total": "9.00"
}'

retuns id such as:
{"id":"2adfe6a5-1b0d-4a26-880a-bfece566118d"}

confirm working with
curl -X GET http://localhost:8080/receipts/<id>/points

returns
{"points":109}


 tests can be run with
go test ./...