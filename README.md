# Chat Messenger Using Redis Pub/Sub

## Description
This project is a chat messenger application that utilizes Redis Pub/Sub for message passing between clients and the server. It's designed to showcase how to implement a real-time chat system using Go and Redis. The application is divided into two main components: the client and the server, which communicate over Redis Pub/Sub channels.

## Features
- Real-time messaging between clients through a server
- Use of Redis Pub/Sub for efficient message broadcasting
- Separate client and server modules for modular development
- Easy to setup and run with Docker or manually

## Installation
### Prerequisites
- Go (1.15 or later)
- Redis server

### Setting Up
1. Clone the repository:
```bash
git clone <repository-url>
```
2. Install dependencies:
```bash
cd <project-directory>
go mod tidy
```

## Usage
### Running the Server
1. Navigate to the server directory:
```bash
cd server
```
2. Run the server:
```bash
go run main.go
```
### Running the Clients
1. Open a new terminals and navigate to the client directory:
```bash
cd client
```
2. Run the client:
```bash
go run main.go
```
## Dependencies
- Redis Go client (for details, refer to `go.mod` and `go.sum` files)
- Other Go packages as specified in `go.mod`

