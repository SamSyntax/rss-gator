# Gator

**Gator** is an RSS feed aggregator written in Go. It provides a command-line interface to manage RSS feeds, user accounts, and feed subscriptions. The application uses PostgreSQL for data storage, requiring a PostgreSQL connection string for configuration.

---

## Features

- **User Management**:
  - Register and log in users.
  - Reset (delete) user accounts.
  - List all registered users.

- **Feed Management**:
  - Add new RSS feeds to the system.
  - View all available feeds.
  - Follow or unfollow feeds.
  - List feeds followed by the logged-in user.

- **Feed Aggregation**:
  - Fetch and aggregate RSS feed content.

---

## Prerequisites

- **Go**: Ensure you have Go installed. [Download Go](https://golang.org/dl/)
- **PostgreSQL**: Install PostgreSQL and have access to a running instance. [PostgreSQL Setup](https://www.postgresql.org/download/)

---

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/SamSyntax/gator.git
   cd gator
   go mod tidy
   go build -o gator .
   ```

