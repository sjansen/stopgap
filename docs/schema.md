# DynamoDB Schema

## Primary Keys

| Entity  | PK            | SK            |
|---------|---------------|---------------|
| Channel | CHAN#slack_id |               |
| Mutex   | MUTEX#name    |               |
| Role    | ROLE#name     |               |
| User    | USER#slack_id |               |

## Secondary Keys

| Entity  | GSI1PK        | GSI1SK     | Notes     |
|---------|---------------|------------|-----------|
| Mutex   | USER#slack_id | MUTEX#name | Locked By |

## Use Cases

| Access Pattern | Index | Parameters | Notes   |
|----------------|--------|-----------|---------|
| Create Mutex   |        |           |         |
| Delete Mutex   |        |           |         |
| Get Mutex      |        |           |         |
| List Mutexes   |        |           |         |
| Lock Mutex     |        |           |         |
| Unlock Mutex   |        |           |         |
