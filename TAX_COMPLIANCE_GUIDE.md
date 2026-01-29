# Tax-Compliance View Mode - Implementation Guide

## Overview
This implementation provides a secure Tax-Compliance view mode for the FKlausnir Enterprise Dashboard that filters out Private/Crypto transactions, showing only Public transactions suitable for official audits.

## Key Features

### 🔒 Tax-Compliance Mode (Default)
- **Automatically filters** to show only transactions flagged as 'Public'
- **Professional UI** designed for official tax audits
- **Audit-ready formatting** with clear transaction categorization
- **No authentication required** - safe for external auditors

### 🔓 Full View Mode (Architect-Key Protected)
- **Requires Architect-Key authentication** for access
- Shows all transactions including Private/Crypto assets
- Encrypted data markers for sensitive transactions
- Protected by environment variable or default key

## Architecture

### Backend (Go)
- **main.go**: RESTful API server with view mode filtering
- `/api/dashboard?view=tax-compliance`: Returns only Public transactions
- `/api/dashboard?view=full`: Requires `X-Architect-Key` header, returns all transactions
- **Encryption markers**: Private/Crypto transactions include encrypted data indicators

### Frontend (HTML/JS)
- **dashboard.html**: Professional single-page dashboard
- Toggle between Tax-Compliance and Full View modes
- Real-time statistics and transaction ledger
- Responsive design suitable for audit presentations

## Running the Application

### Prerequisites
- Go 1.21 or higher

### Quick Start
```bash
# Build and run
go run main.go

# Or build binary
go build -o finnur-fk main.go
./finnur-fk
```

The server will start on `http://localhost:8080`

### Configuration
**IMPORTANT**: Set the Architect Key via environment variable (required for Full View):
```bash
export ARCHITECT_KEY="your-strong-random-key-here"
go run main.go
```

The system will NOT provide a default key for security reasons. Full view mode will be unavailable without setting this environment variable.

## Usage

### Tax-Compliance View (Default)
1. Open browser to `http://localhost:8080`
2. Dashboard loads in Tax-Compliance mode automatically
3. Only Public transactions are visible
4. ✓ AUDIT APPROVED stamp displayed
5. Suitable for sharing with tax authorities

### Full View (Secure)
1. Click "🔓 Full View (Requires Key)" button
2. Enter Architect Key in the password field
3. Click "Unlock"
4. All transactions (including Private/Crypto) are now visible

## Transaction Types

### Public Transactions
- Salary payments
- Business expenses
- Consulting services
- Equipment purchases
- Utility bills
- Standard business income/expenses

### Private/Crypto Transactions (Hidden in Tax-Compliance View)
- Bitcoin purchases
- Ethereum staking
- Crypto exchange transfers
- NFT purchases
- Other crypto-related activities
- Marked with `EncryptedData` field

## Security Features

1. **View Mode Filtering**: Server-side enforcement of transaction visibility
2. **Architect-Key Protection**: Private/Crypto data requires authentication via environment variable
3. **No Default Key**: System requires explicit key configuration for security
4. **Encrypted Data Markers**: Sensitive information stored with encryption indicators
5. **Privacy by Design**: Default view is audit-safe without exposing sensitive data
6. **No Data Leakage**: Private transactions completely absent from Tax-Compliance API response
7. **Generic Error Messages**: Authentication errors don't reveal implementation details

### Production Security Requirements
⚠️ **IMPORTANT**: For production deployments:
- Always use HTTPS to protect the Architect-Key in transit
- Generate a strong, random Architect-Key (e.g., using `openssl rand -base64 32`)
- Never commit the Architect-Key to version control
- Rotate keys periodically
- Use environment variables or secure secret management systems

## Sample Data
The implementation includes 10 sample transactions:
- 6 Public transactions (suitable for tax compliance)
- 4 Private/Crypto transactions (hidden in tax-compliance view)

## API Endpoints

### GET /api/dashboard
Query parameters:
- `view`: "tax-compliance" (default) or "full"

Headers (for full view):
- `X-Architect-Key`: Authentication key for accessing Private/Crypto data

Response:
```json
{
  "transactions": [...],
  "view_mode": "tax-compliance",
  "total_amount": 10350.00,
  "count": 6
}
```

### GET /api/transactions?id={transactionId}
Retrieve specific transaction details. Requires `X-Architect-Key` header for Private/Crypto transactions.

## Testing

### Test Tax-Compliance View
```bash
curl http://localhost:8080/api/dashboard?view=tax-compliance
```
Should return only 6 Public transactions.

### Test Full View (without key)
```bash
curl http://localhost:8080/api/dashboard?view=full
```
Should return 401 Unauthorized.

### Test Full View (with key)
```bash
export ARCHITECT_KEY="your-strong-random-key"
curl -H "X-Architect-Key: $ARCHITECT_KEY" \
     http://localhost:8080/api/dashboard?view=full
```
Should return all 10 transactions.

## Professional Audit Presentation
The Tax-Compliance view provides:
- Clean, professional layout suitable for official presentations
- Color-coded transaction types and amounts
- Clear categorization (Income, Expense, Investment, Transfer)
- Summary statistics (Total Transactions, Net Amount)
- Audit approval stamp for compliance assurance
- Print-friendly format

## Future Enhancements
- Database integration for persistent storage
- Multi-user authentication system
- Transaction import from financial institutions
- PDF export for audit reports
- Advanced filtering and date range selection
- Real-time encryption/decryption of Private/Crypto data

---

**"Johnson Style: We don't just count the nuts, we certify the forest."** 🟢🚀📈
