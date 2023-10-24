# Web Crawler Code Documentation

This document provides an overview of the code for the Web Crawler project. The code is written in Go and uses various libraries and patterns to achieve web crawling functionality.

## Table of Contents

- [Introduction](#introduction)
- [Code Overview](#code-overview)
- [Dependencies](#dependencies)
- [Concurrency Patterns](#concurrency-patterns)
- [Functions and Handlers](#functions-and-handlers)
- [How to Use](#how-to-use)
- [License](#license)

## Introduction

The Web Crawler code is designed to crawl web pages, handle different types of customers (paying and non-paying), and provide efficient real-time or cached results. It utilizes Go and various libraries to achieve its goals.

## Code Overview

The code is organized into several parts:

- **Main Function**: The entry point of the program where the HTTP server is initialized and handlers are defined.

- **Connect to Redis**: Establishes a connection to a Redis server for caching.

- **Admin Initializer**: Reads and sets configuration parameters from Redis for the number of workers and crawling speed.

- **Hourly Reset**: Periodically resets the hourly crawl limit.

- **Worker Functions**: Functions for actual web crawling tasks.

- **Handler Functions**: HTTP request handlers for different routes.

## Dependencies

The code relies on the following external libraries:

- [github.com/PuerkitoBio/goquery](https://pkg.go.dev/github.com/PuerkitoBio/goquery): Used for parsing and manipulating HTML documents.

- [github.com/go-redis/redis](https://pkg.go.dev/github.com/go-redis/redis): Used for connecting to and interacting with a Redis database.

- [github.com/gocolly/colly/v2](https://pkg.go.dev/github.com/gocolly/colly/v2): Used for web scraping and crawling tasks.

## Concurrency Patterns

The code leverages various concurrency patterns to improve performance:

- **Go Routines**: Used to handle concurrent tasks, such as multiple web crawling tasks.

- **Channels**: Used for communication between different parts of the code.

- **Worker Pool Pattern**: Implemented to control the number of concurrent workers.

## Functions and Handlers

The code defines several important functions and HTTP request handlers, such as:

- `hourlyReset`: A function that resets the hourly crawl limit.

- `connectRedis`: Establishes a connection to a Redis server for caching.

- `adminInitializer`: Reads and sets configuration parameters from Redis for the number of workers and crawling speed.

- `goqueryHandler`: A function for manipulating HTML documents with Goquery.

- `modifyHtml`: Modifies HTML content by handling URLs in the document.

- `setWorkerFunc`: An HTTP handler to set the number of worker threads for crawling.

- `setSpeedFunc`: An HTTP handler to set the crawling speed.

- `workers`: A function to perform web crawling tasks.

- `crawler`: An HTTP handler for processing web crawling requests.

## How to Use

To use this code, follow these steps:

1. [Prerequisites](#prerequisites): Ensure that you have the necessary prerequisites, including Go and required libraries.

2. [Installation](#installation): Clone the repository and build the project.

3. [Configuration](#configuration): Configure the code by setting the number of workers and crawling speed through the provided API endpoints.

4. [Run](#run): Start the program, and it will listen on a specified port for incoming requests.

5. [Usage](#usage): Use the provided web interface to enter URLs and initiate web crawling tasks.

## License

This project is licensed under the MIT License. You are free to use, modify, and distribute the code in compliance with the license terms.

For detailed code explanations, please refer to the source code in the Go file.

