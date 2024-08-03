# API Metrics Tracker

## Overview

API Metrics Tracker is a web application that allows you to measure and track various metrics of API endpoints. It provides a web interface to view the metrics and an API endpoint to fetch the metrics in JSON format.

## Features

- Measure API endpoint metrics such as request size, response size, response time, timestamp, and energy consumption.
- View metrics in a web interface.
- Fetch metrics via an API endpoint in JSON format.

## Endpoints

### Web Interface

- `/metrics` - Displays the metrics page.
- `/measure` - Processes API endpoint form submission and measures API metrics.

### API

- `/api/metrics` - Returns the metrics in JSON format.

## Installation

1. Clone the repository:
    ```sh
    git clone https://github.com/yourusername/apimetrics.git
    cd apimetrics
    ```

2. Install dependencies:
    ```sh
    go mod tidy
    ```

3. Set up the database:
    ```sh
    # Assuming you have a PostgreSQL database
    psql -U yourusername -d yourdatabase -f schema.sql
    ```

4. Run the application:
    ```sh
    go run main.go
    ```

5. Open your browser and navigate to `http://localhost:8080/metrics`.

## Usage

1. Open the `/metrics` page to view the metrics.
2. Use the `/measure` endpoint to submit an API endpoint for measurement.
3. Use the `/api/metrics` endpoint to fetch the metrics in JSON format.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## License

This project is licensed under the MIT License.