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
 ### Required Features
- Web crawling on-demand.
- Caching mechanism for recently crawled pages using Redis.
- Real-time crawling for uncached pages.
- Retrying unavailable pages for temporary unavailability.
- Prioritization of paying customers.
### Good-to-Have Features
- Concurrent crawling with multiple workers and a worker pool pattern.
### Great-to-Have Features
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
### Variables 
 ```go
 var (
	// mu is a Mutex used for protecting concurrent access to shared resources.
	mu sync.Mutex

	// client is a pointer to a Redis client for caching.
	client *redis.Client

	// numNonPayingWorker represents the number of non-paying worker threads.
	numNonPayingWorker = 2

	// numPayingWorker represents the number of paying worker threads.
	numPayingWorker = 5

	// payingWorkers is a channel for paying customers to acquire worker threads.
	payingWorkers = make(chan string, numPayingWorker)

	// nonPayingWorkers is a channel for non-paying customers to acquire worker threads.
	nonPayingWorkers = make(chan string, numNonPayingWorker)

	// results is a map used for tracking completed tasks or pages.
	results = make(map[string]bool)

	// pagesPerHour defines the default maximum pages per hour per worker.
	pagesPerHour = 10

	// port represents the network port the application listens on.
	port = ":3000"

	// limit calculates the overall crawling limit based on worker counts and pages per hour.
	limit = pagesPerHour * (numNonPayingWorker + numPayingWorker)
)
```
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
- `connectRedis` : 
   - checks for an existing Redis client to prevent redundant connections.
   - It defines a Redis connection URL with connection details.
   - The function parses the Redis URL, handling potential parsing errors.
   - If URL parsing fails, it prints an error message and terminates the program.
   - On successful parsing, it creates a Redis client using the parsed options.
   - It confirms the successful Redis connection with a printed message.
   - The function may call an additional initialization function, such as adminInitializer.
- `hourlyReset` :
   - It sets up a loop using a time.Tick that triggers once every hour.
   - Within the loop, it recalculates the crawling limit based on the current settings, taking into account the number of non-paying and paying workers.
- `adminInitializer` : 
   - Retrieves configuration settings from a Redis database to update the crawling system.
   - It fetches the number of paying workers from Redis and handles errors in case of retrieval issues.
   - It similarly retrieves the number of non-paying workers and handles any errors that may arise during retrieval.
   - The function converts the retrieved values to integers and updates global variables `numPayingWorker` and `numNonPayingWorker`.
   - It also retrieves the crawling speed from Redis, handling retrieval errors.
   - The function converts the speed value to an integer and updates the global `pagesPerHour` variable.
   - It recreates the worker channels (`payingWorkers` and `nonPayingWorkers`) based on the updated worker counts.
   - Lastly, it recalculates the crawling limit based on the updated values, ensuring the system reflects the most recent configurations.
- `crawler` :
   - This function is responsible for handling user requests to crawl a URL and serve its content.
   - It extracts the requested URL and determines whether the user is a paying customer.
   - It checks the availability of the Redis client for caching.
   - Inside a critical section, it attempts to retrieve the URL's content from the Redis cache.
   - If the content is found in the cache, it serves the cached content and exits.
   - It checks if the hourly crawl limit has been exceeded and returns an appropriate response if it has.
   - The `workerChan` is chosen based on the user type, paying or non-paying.
   - A worker is acquired from the appropriate worker channel.
   - A goroutine is launched to perform the crawling in parallel.
   - It repeatedly checks the `results` map for the target URL, with a maximum of 20 retries.
   - When the target URL is found in `results`, it re-enters a critical section and retrieves the content from Redis.
   - If the content is found, it is deleted from `results`, served, and the function returns.
   - If the target URL is not found within the retry limit, a Not Found (404) response is returned.
- `workers` :
   - This function is responsible for handling the crawling of a specific URL.
   - It decrements the hourly crawl limit for the provided URL.
   - The HTML content of the URL is fetched with a retry mechanism using the `fetchWithRetryMechanism` function.
   - It handles errors that may occur during the fetching process.
   - If the fetching process results in errors, it logs an error message and returns, effectively terminating the crawling for that URL.
   - The fetched HTML content is then modified using the `modifyHtml` function to fix URLs.
   - It checks for errors during the modification process, but it does not handle them in a specific way.
   - Inside a critical section, it updates the Redis cache with the modified HTML content for the URL.
   - It handles errors that might occur during the Redis update and logs any encountered errors.
   - A confirmation is sent to the `results` queue, indicating that the URL has been successfully crawled.
   - The worker is released by removing it from the worker channel, making it available for future crawling tasks.
- `fetchWithRetryMechanism` :
   - This function is responsible for attempting to fetch the HTML content of a URL with a retry mechanism.
   - It takes the target URL as input and returns the fetched HTML content and an errors indicator.
   - It initializes the `errs` variable with the string "ERRORS" to indicate that errors have occurred during fetching. It also initializes the `html` variable to an empty string to store the fetched HTML content.
   - The function attempts up to 15 retries to fetch the HTML content from the provided URL.
   - For each retry, it creates a new Colly collector (`c`) to make the GET request to the URL.
   - It defines a callback function to handle the response and store the HTML content if it's not an empty string.
   - After making the GET request, it checks for errors. If an error occurs, it logs the number of the retry and continues to the next iteration to retry the request.
   - If the HTML content is successfully fetched (i.e., it's not an empty string), the `errs` variable is cleared (set to an empty string), and the function returns the fetched HTML content along with the cleared errors indicator.
   - If no content is received after a retry, it logs the number of the retry and sleeps for 2 seconds before retrying.
   - After all retries are exhausted, the function returns the HTML content and the "ERRORS" indicator to signify that errors occurred during the retries.
- `modifyHtml` :
   - This function is responsible for modifying the HTML content to fix URLs within specific HTML tags.
   - It accepts two parameters:
   - `html` (string): The input HTML content that needs to be modified.
   - `URL` (string): The base URL used to adjust relative URLs found in the HTML.
   - The function starts by attempting to parse the input HTML content into a goquery document using `goquery.NewDocumentFromReader`. If there's an error during parsing, it logs the error using `fmt.Println`.
   - The `goqueryHandler` function is called three times to modify URLs within specific HTML tags: "img" tags with "src" and "srcset" attributes, and "script" tags with "src" attributes.
   - After applying these URL adjustments, the function generates the modified HTML content using the `doc.Html()` method. If an error occurs during this process, it logs the error using `log.Println`.
   - Finally, the function returns the modified HTML content (or the original HTML content if no modifications were made) and any encountered error.
- `goqueryHandler` :
   - It takes four parameters: `doc` (the goquery document to be modified), `tag` (the HTML tag of elements to be targeted for URL modification), `attr` (the name of the attribute within the elements to be modified), and `URL` (the base URL used to adjust relative URLs in the HTML).
   - The function begins by finding all elements with the specified HTML tag within the provided goquery document using `doc.Find(tag)`.
   - For each matching element, it retrieves the value of the specified attribute (`attr`) and checks if the attribute exists for the element.
   - If the attribute exists, it proceeds to modify the URL within the attribute:
   - It extracts the first five characters of the attribute value.
   - It checks whether the first five characters of the attribute value are not equal to "https," indicating that the URL is not an absolute URL.
   - If the URL is not absolute, it prepares a new URL based on the provided base URL.
   - It checks whether the attribute value does not start with a forward slash, and if not, it appends a forward slash to the new URL.
   - Finally, it modifies the attribute of the element to use the new URL, effectively converting relative URLs to absolute URLs.
   - This function ensures that URLs within specific HTML elements are correctly formatted and adjusted based on the provided base URL, making them absolute URLs when necessary.
- `setWorkerFunc` :
   - THis function is an HTTP request handler function that processes requests to update the number of paying and non-paying workers.
   - It retrieves the values for the new number of paying and non-paying workers from the request's URL query parameters: `setPayingWorkerTo` and `setNonPayingWorkerTo`.
   - The retrieved values are converted to integers and used to update the global variables `numPayingWorker` and `numNonPayingWorker`.
   - To ensure safe access to shared resources during Redis operations, the function locks the mutex (`mu`) before proceeding.
   - It updates the number of paying workers in the Redis database using the `client.Set` function, and then unlocks the mutex.
   - After each Redis operation, the function checks for errors and responds with an error message if an error occurs. Additionally, it logs the error for debugging purposes.
   - The function then locks the mutex again to update the number of non-paying workers in the Redis database, following a similar pattern of Redis interaction and error handling.
   - Once both the paying and non-paying worker counts are updated in Redis, it adjusts the channels (`payingWorkers` and `nonPayingWorkers`) to match the new worker counts.
   - Finally, it responds with a success message indicating that the number of workers has been updated.
- `setSpeedFunc` 
   - This is an HTTP request handler function that processes requests to update the crawling speed per hour per worker.
   - It retrieves the new crawling speed value from the request's URL query parameters, specifically from the parameter named `setWorkerSpeedTo`.
   - The retrieved value is converted to an integer and used to update the global variable `pagesPerHour`.
   - To ensure safe access to shared resources during Redis operations, the function locks the mutex (`mu`) before proceeding.
   - It updates the crawling speed in the Redis database using the `client.Set` function and then unlocks the mutex.
   - After the Redis operation, the function checks for errors, and if an error occurs, it prints the error.
   - The function then recalculates the crawling limit based on the updated crawling speed and the number of workers.
   - Finally, it responds with a success message indicating that the crawling speed has been updated.