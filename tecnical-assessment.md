# Technical Assessment: URL Shortener Service

## 📌 Overview

Your task is to design and implement a scalable, highly available **URL Shortener Service** (similar to Bitly or TinyURL).

This project is designed to evaluate your software architecture skills, choice of data stores, clean code practices, performance optimizations, and API design.

---

## 🎯 High-Level Requirements

### Functional Requirements

1. **URL Shortening:**
* Given a valid long URL, the service must return a unique, short URL.


2. **Redirection:**
* Accessing the short URL must redirect the user to the original long URL.
* *Decision point:* You may decide whether to use HTTP `301` or `302` redirects based on your architectural trade-offs.


3. **Custom Alias (Optional/Bonus):**
* Users should be able to specify a custom short key (e.g., `my.service/custom-alias`) if available.


4. **Analytics (Optional/Bonus):**
* Track the number of clicks for each short URL.



### Non-Functional Requirements

1. **High Availability & Low Latency:**
* The redirection endpoint should be optimized for extremely fast read responses.


2. **Collision Prevention:**
* Shortened keys must be unique. No two long URLs should collide onto the same active key.


3. **Scalability:**
* Design the application and database layer with horizontal scaling in mind (high read-to-write ratio).



---

## 🏗️ Technical Constraints & Freedom

* **Programming Language / Framework:** Free choice (e.g., Go, Node.js, Java, Python, C#, Rust).
* **Database:** Free choice (Relational, NoSQL, Key-Value, or a combination).
* **Caching:** You may use in-memory caching mechanisms (e.g., Redis, Memcached) if deemed necessary for performance.
* **Architecture:** Monolith, Microservices, or Serverless — choose the architecture that best fits your design rationale.

---

## 📦 Deliverables

1. **Source Code:**
* Fully functional codebase pushed to a Git repository.
* Clean, readable code following standard conventions for your chosen language.


2. **Documentation (`README.md`):**
* Instructions on how to run and test the application locally (preferably using Docker / Docker Compose).
* **System Architecture Overview:** A brief explanation of your design choices, data flow, hash/ID generation strategy, and database schema.
* **Trade-offs & Rationale:** Explain why you chose your specific stack, database type, HTTP redirect code (`301` vs `302`), and caching strategy.


3. **API Documentation:**
* Clear specification of the endpoints (Swagger/OpenAPI spec, Postman collection, or Markdown format).



---

## 🧪 Evaluation Criteria

Your submission will be reviewed based on the following criteria:

* **Architecture & System Design:** Thoughtful separation of concerns, data modeling, and performance considerations.
* **Code Quality:** Readability, maintainability, error handling, and test coverage.
* **Problem Solving:** How effectively you handle edge cases (e.g., invalid URLs, collisions, expired links, non-existent keys).
* **Documentation:** Clarity in explaining architectural decisions and setup instructions.