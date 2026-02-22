## FK Master Ledger

This file is the master, append-only ledger for FKlausnir settlement events and audit entries.

Guidelines
- Do NOT store raw PayPal credentials or full transaction dumps in this repository.
- Only store sanitized, minimal audit records. Prefer storing cross-reference hashes and metadata.
- All entries should be appended; do not modify historical rows without an explicit amendment entry.

Format (Markdown table)
| Date (UTC) | TxID / Reference | Source | Amount | Currency | CrossRefSHA | Auditor | Notes |
|---|---|---:|---:|---:|---|---|---|

Example entry
| 2026-01-15T12:34:56Z | FK-TX-000001 | PayPal | 1000.00 | EUR | abcd1234... | auditor-1 | Initial settlement entry |

Automated append
- The ledger can be appended by automated audit tools (for example ledger/audit_logic.go) which must sanitize inputs and write with restricted file permissions.

Amendments
- If an entry must be changed, add a new line indicating the amendment and reference the original TxID.

Security
- Set file permissions to 0600 when written by automation.
- Use GitHub Secrets and private artifact storage for any sensitive artifacts.
