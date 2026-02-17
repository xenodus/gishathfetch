# Gishath Fetch

Gishath Fetch is a high-performance web application designed for Magic: The Gathering players to search for card singles across multiple local game stores (LGS) concurrently. It streamlines the process of finding the best prices and availability by aggregating data from various sources in real-time.

## 🚀 Features

- **⚡ Concurrent Search**: Scrapes multiple LGS websites simultaneously using a highly parallel Go backend for near-instant results.
- **🎯 Precision Filtering**: Implements advanced result filtering and normalization to ensure high match accuracy.
- **💰 Price Sorting**: Automatically compiles and sorts results from all sources by price, making it easy to find the best deal.
- **🛒 Unified Shopping Cart**: Save cards from different stores into a single, persistent cart for easy tracking.
- **🔍 Smart Suggestions**: Real-time search suggestions with term highlighting to help you find the right card faster.

## 🏗️ Architecture

Gishath Fetch utilizes a modern decoupled architecture:

- **Frontend**: A responsive Single Page Application (SPA) built with **React 19** and **Vite**. It uses **Bootstrap 5** for a clean, mobile-friendly UI and custom React hooks for efficient state management of search results and the unified shopping cart.
- **Backend**: A robust API written in **Go 1.26.0**, deployed as **AWS Lambda** functions. The backend leverages Go's concurrency primitives (Goroutines) to perform multiple LGS scrapings in parallel, aggregating the results before returning them to the client.

## 🛠️ Tech Stack

### Frontend
- **Framework**: [React 19](https://react.dev/)
- **Build Tool**: [Vite](https://vitejs.dev/)
- **Styling**: [Bootstrap 5](https://getbootstrap.com/)
- **Icons**: [React Feather](https://feathericons.com/)

### Backend
- **Language**: [Go 1.26.0](https://go.dev/)
- **Infrastructure**: [AWS Lambda](https://aws.amazon.com/lambda/)
- **Scraping**: [Colly](http://go-colly.org/), [GoQuery](https://github.com/PuerkitoBio/goquery)
- **Environment**: [GoDotEnv](https://github.com/joho/godotenv)
- **ID Generation**: [Sonyflake](https://github.com/sony/sonyflake)

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](file:///Users/paya0969/Desktop/projects/gishathfetch/LICENSE) file for details.

---

*Gishath Fetch is not affiliated with Wizards of the Coast or any of the supported local game stores.*
