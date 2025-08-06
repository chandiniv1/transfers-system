# 🔐 Transfers System (Go + PostgreSQL)

This project implements a basic internal transfers backend API in Go, supporting:

- Account creation
- Balance tracking
- Transaction logging
- Transfers between accounts

---

## 📁 Project Structure

```
.
├── main.go             # Entry point of the application
├── go.mod / go.sum     # Go modules
├── server/             # HTTP server and routing logic
├── db/                 # DB interaction logic using sqlc
├── utils/              # Helper utilities (error handling, validation)

```

---

## 🛠️ 1. Installation & Setup

### Clone the repository

```bash
git clone https://github.com/chandiniv1/transfers-system.git
cd transfers-system
```

### Install dependencies

```bash
go mod tidy
```

### Setup PostgreSQL

```bash
make network
make postgres
```

### Create DataBase

make sure that postgres is running

```bash
make createdb
```

### Run DataBase Migrations

```bash
make migrateup
```

To rollback all migrations:

```bash
make migratedown
```

To create a new migration:

```bash
make new_migration name=<name>
```

### Generate SQL code

```bash
make sqlc
```

### Run the server

```bash
make server
```

Server will start on: `http://localhost:8080`

---

## 📬 2. API Endpoints

### ✅ Create Account

- **POST** `/accounts`
```json
{
  "account_id": 2,
  "balance": 1000,
  "currency": "USD"
}
```

### 🔍 Get Account

- **GET** `/accounts/{id}`

### 🔁 Transfer Between Accounts

- **POST** `/transactions`
```json
{
  "from_account_id": 1,
  "to_account_id": 2,
  "amount": 300,
  "currency": "USD"
}
```


### 📄 List Accounts

- **GET** `/accounts?page_id=1&page_size=5`

---

## 🧪 3. Testing

Run tests with:

```bash
make test
```

---

## 📖 4. Tech Stack

- [Golang](https://golang.org/)
- [PostgreSQL](https://www.postgresql.org/)
- [sqlc](https://docs.sqlc.dev/)
- [gin](https://gin-gonic.com/en/docs/)

---
