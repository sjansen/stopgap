# DynamoDB Schema

## Primary Keys

| Entity  | PK            | SK            |
|---------|---------------|---------------|
| Channel | CHAN#slack_id | CHAN#slack_id |
| Mutex   | MUTEX#name    | MUTEX#name    |
| Role    | USER#slack_id | USER#slack_id |
| User    | ROLE#name     | ROLE#name     |

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
