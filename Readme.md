# Web Crawler Project

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Dependencies](#dependencies)
  - [Installation](#installation)
- [Usage](#usage)
- [Paying and Non-Paying Customers](#paying-and-non-paying-customers)
- [Concurrent Crawling](#concurrent-crawling)
- [Administrative Control](#administrative-control)
- [Caching with Redis](#caching-with-redis)
- [Concurrency Patterns](#concurrency-patterns)
- [Contributing](#contributing)
- [License](#license)

## Overview

The Web Crawler Project is designed to provide a web crawling solution with a user-friendly web interface. Users can request the crawling of specific web pages, and the server efficiently handles these requests, prioritizing paying customers and providing reliable real-time or cached results.

## Features

- Web crawling on-demand.
- Caching mechanism for recently crawled pages using Redis.
- Real-time crawling for uncached pages.
- Retrying unavailable pages for temporary unavailability.
- Prioritization of paying customers.
- Concurrent crawling with multiple workers and a worker pool pattern.
- Administrative control over the number of workers and crawling speed.

## Getting Started

### Prerequisites

Before you begin, make sure you have the following prerequisites:

- Go (1.13+) installed on your machine.
### Dependencies

The code relies on the following external libraries:

- [github.com/PuerkitoBio/goquery](https://pkg.go.dev/github.com/PuerkitoBio/goquery): Used for parsing and manipulating HTML documents.

- [github.com/go-redis/redis](https://pkg.go.dev/github.com/go-redis/redis): Used for connecting to and interacting with a Redis database.

- [github.com/gocolly/colly/v2](https://pkg.go.dev/github.com/gocolly/colly/v2): Used for web scraping and crawling tasks.

### Installation

1. Clone this repository:

   ```sh
   git clone https://github.com/shin-igami/sendx-backend-iec2020039.git
2. Change to the project directory:

   ```sh
    cd web-crawler
4. Install Dependencies
    ```go
      go get "github.com/PuerkitoBio/goquery"
      go get "github.com/go-redis/redis"
      go get "github.com/gocolly/colly/v2"
3. Build the project:
    ```go
     go build
4. Run the Project
   ```go
      go run main.go
## Usage

To use the web crawler, access the provided web interface, enter the desired URL, and click the "Crawl" button. The server will handle the crawling process, providing either cached or real-time results based on the URL's recent crawling history.

## Paying and Non-Paying Customers

The system differentiates between paying and non-paying customers using query parameters passed to the backend through the frontend API call. Paying customers receive priority in the crawling queue.

## Concurrent Crawling

Concurrent crawling is supported, with a worker pool pattern and multiple workers available for paying customers and non-paying customers. This ensures maximum throughput and responsiveness.

## Administrative Control

Administrators have the ability to manage the crawling infrastructure through the provided API endpoints:

- **Number of Crawler Workers**: Administrators can configure the number of workers available for crawling to meet demand dynamically.

- **Crawling Speed Limit**: Administrators can set the crawling speed per hour per worker, and the system actively enforces these limits. When the hourly crawl limit is exceeded, an error is returned.

## Caching with Redis

This project utilizes Redis as a caching mechanism for storing and retrieving recently crawled pages. This enhances performance and reduces the load on the server.

## Concurrency Patterns

The code leverages various concurrency patterns to improve performance:

- **Go Routines**: Used to handle concurrent tasks, such as multiple web crawling tasks.

- **Channels**: Used for communication between different parts of the code.

- **Worker Pool Pattern**: Implemented to control the number of concurrent workers.
## Documentation

### Code Overview

The code is organized into several parts:

- **Main Function**: The entry point of the program where the HTTP server is initialized and handlers are defined.

- **Connect to Redis**: Establishes a connection to a Redis server for caching.

- **Admin Initializer**: Reads and sets configuration parameters from Redis for the number of workers and crawling speed.

- **Hourly Reset**: Periodically resets the hourly crawl limit.

- **Worker Functions**: Functions for actual web crawling tasks.

- **Handler Functions**: HTTP request handlers for different routes.

### Functions and Handlers
 - `main` : This main function serves as the entry point of the web crawler application. It registers HTTP request handlers for different endpoints, initializes necessary components, such as Redis and the hourly reset goroutine, and starts the HTTP server to listen on the specified port.
```go
func main() {
	// Register HTTP request handlers for different endpoints.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Serve the HTML file for the web interface.
		http.ServeFile(w, r, "./static/index.html")
	})
	http.HandleFunc("/results", crawler)
	http.HandleFunc("/set-workers", setWorkerFunc)
	http.HandleFunc("/set-speed", setSpeedFunc)

	// Perform the necessary setup and initialization steps.

	// Connect to the Redis database.
	connectRedis()

	// Start a goroutine to reset the hourly crawling limit.
	go hourlyReset()

	// Set up and start the HTTP server to listen on the specified port.
	fmt.Printf("Connected on port %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Println("Error occurred while connecting to Port")
		log.Fatal(err)
	}
}
```
 - `connectRedis` : This function, connectRedis, is responsible for establishing a connection to a Redis server for caching. It checks whether a connection already exists, and if not, it initializes a new connection using the provided Redis URL.
```go
func connectRedis() {
	// Check if a Redis client is already initialized
	if client != nil {
		return
	}

	// Define the Redis connection URL
	redisURL := "redis://default:glJ8STTn7PUyYjqjc6h1imIh89pFQHA0@redis-15188.c212.ap-south-1-1.ec2.cloud.redislabs.com:15188"

	// Parse the Redis URL and handle any potential errors
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		// If there's an error while parsing the URL, panic (terminate the program) with the error message
		fmt.Println("ERROR: Cannot connect to Database")
		panic(err)
	}

	// Create a new Redis client using the parsed options
	client = redis.NewClient(opt)

	// Print a message indicating that the connection to Redis was successfully established
	println("Connected to Redis")

	adminInitializer()
}
```
-`hourlyReset` : 
```go

