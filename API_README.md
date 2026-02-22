# PayPal CSV Processing & Liquidity Calculator API

This backend service provides endpoints to upload PayPal CSV files and calculate total liquidity based on transaction data.

## Features

- **CSV Upload**: Upload and parse PayPal CSV files
- **Liquidity Calculation**: Calculate total liquidity, fees, balances, and transaction counts
- **Multi-Currency Support**: Track liquidity by currency
- **Status Tracking**: Monitor completed, pending, and refunded transactions
- **Comprehensive Error Handling**: Validation of CSV format and data integrity

## API Endpoints

### POST /api/upload

Upload a PayPal CSV file for processing.

**Request:**
- Method: `POST`
- Content-Type: `multipart/form-data`
- Field: `file` (CSV file)

**Response:**
```json
{
  "message": "CSV uploaded and parsed successfully",
  "transaction_count": 150
}
```

**Error Responses:**
- `400 Bad Request`: Invalid CSV format or missing required fields
- `405 Method Not Allowed`: Wrong HTTP method

### GET /api/liquidity

Retrieve calculated liquidity information from the last uploaded CSV.

**Request:**
- Method: `GET`

**Response:**
```json
{
  "total_gross": 15000.50,
  "total_fees": -450.25,
  "total_net": 14550.25,
  "final_balance": 20000.00,
  "transaction_count": 150,
  "by_currency": {
    "USD": 12000.00,
    "EUR": 2550.25
  },
  "completed_count": 145,
  "pending_count": 3,
  "refunded_count": 2
}
```

**Error Responses:**
- `404 Not Found`: No CSV has been uploaded yet
- `405 Method Not Allowed`: Wrong HTTP method

### GET /health

Health check endpoint.

**Request:**
- Method: `GET`

**Response:**
```json
{
  "status": "healthy"
}
```

## PayPal CSV Format

The API expects PayPal CSV files with the following columns (header names are case-insensitive):

### Required Fields:
- `Transaction ID` (or `transaction id`)

### Optional Fields:
- `Date` (or `Timestamp`)
- `Name` (or `From Name`, `To Name`)
- `Type` (or `Transaction Type`)
- `Status` (or `Transaction Status`)
- `Currency` (or `Currency Code`)
- `Gross` (or `Amount`, `Gross Amount`)
- `Fee` (or `Fee Amount`)
- `Net` (or `Net Amount`)
- `Balance` (or `Account Balance`)
- `Note` (or `Message`, `Item Title`)

### Example CSV:
```csv
Transaction ID,Date,Name,Type,Status,Currency,Gross,Fee,Net,Balance
TX123,2024-01-15,John Doe,Payment,Completed,USD,100.00,-2.90,97.10,500.00
TX124,2024-01-16,Jane Smith,Payment,Completed,USD,50.00,-1.75,48.25,548.25
```

## Building and Running

### Build the server:
```bash
go build -o bin/server ./cmd/server
```

### Run the server:
```bash
./bin/server -port 8080
```

The server will start on the specified port (default: 8080).

## Testing

### Run all tests:
```bash
go test ./...
```

### Run tests for a specific package:
```bash
go test ./internal/parser -v
go test ./internal/liquidity -v
go test ./api -v
```

### Test coverage:
```bash
go test -cover ./...
```

## Architecture

```
finnur-fk/
├── api/                      # HTTP API server
│   ├── server.go            # Main server implementation
│   └── server_test.go       # API endpoint tests
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── internal/
│   ├── parser/              # CSV parsing module
│   │   ├── paypal_parser.go
│   │   └── paypal_parser_test.go
│   └── liquidity/           # Liquidity calculation module
│       ├── calculator.go
│       └── calculator_test.go
└── go.mod                   # Go module definition
```

## Error Handling

The API provides detailed error messages for various failure scenarios:

- **Empty CSV**: "CSV file is empty"
- **Missing Transaction ID**: "CSV missing required field: Transaction ID"
- **Invalid Amount**: "cannot parse amount 'value': error details"
- **Empty Transaction ID**: "transaction ID is empty"
- **No Data Available**: "No transactions available. Please upload a CSV first."

## Security Considerations

- Maximum upload size: 10MB
- CSV validation before processing
- Input sanitization for all fields
- No sensitive data stored in logs
- In-memory storage (data cleared on server restart)

## Future Enhancements

- Persistent storage (database)
- Authentication and authorization
- Rate limiting
- CSV export functionality
- Historical data tracking
- Advanced analytics and reporting
