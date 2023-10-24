package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-redis/redis"
	"github.com/gocolly/colly/v2"
)

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

// main is the entry point of the web crawler application.
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
func hourlyReset() {
	// The loop will iterate once every hour
	for range time.Tick(time.Hour) {
		// Recalculate the crawling limit based on the current settings
		limit = pagesPerHour * (numNonPayingWorker + numPayingWorker)
	}
}

// adminInitializer initializes and configures the system based on values stored in Redis.
func adminInitializer() {
	// Retrieve the number of paying workers from Redis
	setPayingWokers, err := client.Get("sendX-numPayingWorker").Result()
	if err != nil {
		// Handle the case where the value cannot be fetched from Redis
		fmt.Println("Can't Fetch paying workers using the default value ")
	}

	// Retrieve the number of non-paying workers from Redis
	setNonPayingWokers, _ := client.Get("sendX-numNonPayingWorker").Result()
	if err != nil {
		// Handle the case where the value cannot be fetched from Redis
		fmt.Println("Can't Fetch non-paying workers using the default value ")
	}

	// Convert the retrieved values to integers and update the global variables
	numPayingWorker, _ = strconv.Atoi(setPayingWokers)
	numNonPayingWorker, _ = strconv.Atoi(setNonPayingWokers)

	// Retrieve the crawling speed from Redis
	setSpeed, _ := client.Get("sendX-workerSpeed").Result()
	if err != nil {
		// Handle the case where the value cannot be fetched from Redis
		fmt.Println("Can't Fetch crawling limit using the default value ")
	}

	// Convert the retrieved value to an integer and update the global pages per hour variable
	pagesPerHour, _ = strconv.Atoi(setSpeed)

	// Recreate the worker channels based on the updated worker counts
	payingWorkers = make(chan string, numPayingWorker)
	nonPayingWorkers = make(chan string, numNonPayingWorker)

	// Recalculate the crawling limit based on the updated values
	limit = pagesPerHour * (numNonPayingWorker + numPayingWorker)
}

// crawler handles HTTP requests for crawling a URL and serving its content.
func crawler(w http.ResponseWriter, r *http.Request) {
	// Extract the URL and determine if the user is paying or non-paying.
	URL := r.URL.Query().Get("url")
	paying := r.URL.Query().Get("paying") == "true"

	// Check if the Redis client is available before making a Redis request.
	if client == nil {
		panic("Error: client is not available")
	}

	// Enter the critical section to safely access Redis.
	mu.Lock()
	val, err := client.Get(URL).Result()
	mu.Unlock()

	// If the URL is found in Redis cache, serve the cached content and return.
	if err == nil {
		io.WriteString(w, val)
		return
	}

	// If the hourly crawl limit is reached, log the limit exceeded message and return.
	if limit == 0 {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Println("Hourly crawl limit is exceeded")
		return
	}

	var workerChan chan string

	// Determine the worker channel based on whether the user is paying or non-paying.
	if paying {
		workerChan = payingWorkers
	} else {
		workerChan = nonPayingWorkers
	}

	workerChan <- URL // Acquire a worker from the respective channel.

	// Launch a goroutine to perform the crawling.
	go workers(URL, workerChan)

	// Attempt to find the result in the results map with a maximum retry count.
	for t := 0; t < 20; t++ {
		if results[URL] == true {
			// Re-enter the critical section to safely access Redis.
			mu.Lock()
			val, err := client.Get(URL).Result()
			mu.Unlock()
			// If the result is found, delete it from the results map, serve the content, and return.
			if err == nil {
				delete(results, URL)
				io.WriteString(w, val)
				return
			}
		}

		time.Sleep(time.Second * 2)
	}

	// If the target URL is not found within the retry limit, return a Not Found (404) response.
	w.WriteHeader(http.StatusNotFound)
}

// workers is responsible for processing a URL, performing crawling, and updating the cache.
func workers(URL string, workerChan chan string) {
	// Decrement the hourly crawl limit for this URL.
	limit--

	// Fetch the HTML content of the URL with a retry mechanism.
	html, errs := fetchWithRetryMechanism(URL)

	// Handle errors during fetching.
	if errs == "ERRORS" {
		fmt.Printf("Cannot Get : %s", URL)
		return
	}

	// Modify the HTML content to fix URLs.
	modifiedHtml, err := modifyHtml(html, URL)

	// Handle the case where the target page is not modified.
	if err != nil {
		fmt.Println("Target Page is not modified", err)
	}

	// Enter the critical section to safely update Redis with the modified HTML.
	mu.Lock()
	err = client.Set(URL, modifiedHtml, 3600*time.Second).Err()
	mu.Unlock()

	// Handle errors during updating Redis.
	if err != nil {
		fmt.Println("Not able to set value in Redis Database")
		log.Println(err)
		return
	}

	// Send a confirmation to the results queue that the URL has been successfully crawled.
	results[URL] = true

	// Release the worker by removing it from the worker channel.
	<-workerChan
}

// fetchWithRetryMechanism fetches the HTML content of a URL with a retry mechanism.
// It makes multiple GET requests with retries to handle possible network issues.
func fetchWithRetryMechanism(URL string) (string, string) {
	// Initialize the errors variable to indicate if there were errors during fetching.
	errs := "ERRORS"
	html := "" // Initialize the HTML content variable.

	// Perform up to 10 retries to fetch the HTML content.
	for i := 1; i <= 10; i++ {
		c := colly.NewCollector()

		// Define a callback function to handle the response and store the HTML content.
		c.OnResponse(func(r *colly.Response) {
			if string(r.Body) != "" {
				html = string(r.Body)
			}
		})

		// Make the GET request to the desired URL.
		err := c.Visit("https://" + URL)

		// Check for errors during the GET request.
		if err != nil {
			fmt.Printf("No. of Retry: %d\n", i)
			continue // Retry the request.
		}

		// If HTML content is successfully fetched, clear the errors and return the content.
		if html != "" {
			errs = ""
			return html, errs
		}

		// If no content is received, sleep for 2 seconds before retrying.
		fmt.Printf("No. of Retry: %d\n", i)
		time.Sleep(time.Second * 2)
	}

	// Return the HTML content and errors (if any) after all retries.
	return html, errs
}

// modifyHtml processes the provided HTML content, ensuring that URLs within specific HTML tags are correctly formatted.
// It takes the input 'html' as a string and a 'URL' string to create an absolute URL if necessary.
// It returns the modified HTML as a string and any error encountered during processing.
func modifyHtml(html string, URL string) (string, error) {
	// Parse the HTML content into a goquery document.
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		// If an error occurs during parsing, log the error.
		fmt.Println("ERROR : ", err)
	}

	// Apply the 'goqueryHandler' function to modify URLs within specific HTML tags.
	goqueryHandler(doc, "img", "src", URL)
	goqueryHandler(doc, "img", "srcset", URL)
	goqueryHandler(doc, "script", "src", URL)

	// Generate the modified HTML content after URL adjustments.
	modifiedHtml, err := doc.Html()
	if err != nil {
		// If an error occurs during generating the HTML, log the error.
		log.Println("Error: ", err)
	}

	// Return the modified HTML content and any encountered error.
	return modifiedHtml, err
}

// goqueryHandler manipulates an HTML document using the goquery library.
// It finds elements with the specified 'tag' and modifies the 'attr' attribute,
// ensuring that URLs are properly formatted relative to the given 'URL'.
func goqueryHandler(doc *goquery.Document, tag string, attr string, URL string) {
	// Find all elements with the specified 'tag' in the HTML document.
	doc.Find(tag).Each(func(index int, element *goquery.Selection) {
		// Retrieve the value of the 'attr' attribute for the current element.
		srcAttr, exists := element.Attr(attr)

		// Check if the 'attr' attribute exists for the current element.
		if exists {
			finalStr := ""

			// Extract the first five characters from the 'srcAttr' value.
			if len(srcAttr) >= 5 {
				for i := 0; i < 5; i++ {
					finalStr += string(srcAttr[i])
				}
			}

			// Check if the first five characters of 'srcAttr' are not "https."
			if finalStr != "https" {
				// Prepare the replacement URL based on the provided 'URL'.
				strTo := "https://" + URL

				// Check if 'srcAttr' does not start with a forward slash.
				if len(srcAttr) >= 1 && srcAttr[0] != '/' {
					strTo += "/"
				}

				// Modify the 'attr' attribute of the element to use the new URL.
				element.SetAttr(attr, strTo+srcAttr)
			}
		}
	})
}

// setWorkerFunc handles HTTP requests to update the number of worker threads for paying and non-paying customers.
func setWorkerFunc(w http.ResponseWriter, r *http.Request) {
	// Retrieve the values for the new number of paying and non-paying workers from the request's URL query parameters.
	setPayingWorker := r.URL.Query().Get("setPayingWorkerTo")
	setNonPayingWorker := r.URL.Query().Get("setNonPayingWorkerTo")

	// Convert the retrieved values to integers and update the global variables.
	numPayingWorker, _ = strconv.Atoi(setPayingWorker)
	numNonPayingWorker, _ = strconv.Atoi(setNonPayingWorker)

	// Lock the mutex to ensure safe access to shared resources during Redis operations.
	mu.Lock()

	// Update the number of paying workers in the Redis database.
	err := client.Set("sendX-numPayingWorker", setPayingWorker, 0).Err()

	// Unlock the mutex after Redis operation.
	mu.Unlock()

	// Check for errors during Redis operation.
	if err != nil {
		// Respond with an error message and log the error.
		w.Write([]byte("Error while Updating"))
		log.Fatal("Error: ", err)
	}

	// Lock the mutex again to ensure safe access to shared resources during Redis operations.
	mu.Lock()

	// Update the number of non-paying workers in the Redis database.
	err = client.Set("sendX-numNonPayingWorker", setNonPayingWorker, 0).Err()

	// Unlock the mutex after Redis operation.
	mu.Unlock()

	// Check for errors during Redis operation.
	if err != nil {
		// Respond with an error message and log the error.
		w.Write([]byte("Error while Updating"))
		log.Fatal("Error: ", err)
	}

	// Update the channels to match the new number of workers.
	payingWorkers = make(chan string, numPayingWorker)
	nonPayingWorkers = make(chan string, numNonPayingWorker)

	// Respond with a success message after updating the worker counts.
	w.Write([]byte("Number of workers updated."))
}

// setSpeedFunc handles HTTP requests to update the crawling speed per hour per worker.
func setSpeedFunc(w http.ResponseWriter, r *http.Request) {
	// Retrieve the new crawling speed per hour per worker from the request's URL query parameters.
	setWorkerSpeed := r.URL.Query().Get("setWorkerSpeedTo")
	fmt.Println(setWorkerSpeed)

	// Convert the retrieved value to an integer and update the 'pagesPerHour' global variable.
	pagesPerHour, _ = strconv.Atoi(setWorkerSpeed)

	// Lock the mutex to ensure safe access to shared resources during Redis operations.
	mu.Lock()

	// Update the crawling speed in the Redis database.
	err := client.Set("sendX-workerSpeed", setWorkerSpeed, 0).Err()

	// Unlock the mutex after the Redis operation.
	mu.Unlock()

	// Check for errors during Redis operation.
	if err != nil {
		println(err)
	}

	// Recalculate the crawling limit based on the updated crawling speed and the number of workers.
	limit = pagesPerHour * (numNonPayingWorker + numPayingWorker)

	// Respond with a success message after updating the crawling speed.
	w.Write([]byte("Crawling speed updated."))
}
