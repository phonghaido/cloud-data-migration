# Data Migration Service from Amazon S3 to Google Cloud Storage

A project for migrating data from one Amazone S3 to Google Cloud Storage. The system is designed according to microservice architecture, each microservice is containerized by using Docker, and then be deployed to a container orchestration system (Kubernetes - K8s). The system allows scaling up (or down) the data migration service to spreate the load (or focus the load) to multiple applications (or one application).



## System Components
- **PubSub Emulator**: Message queue to coordinate and distribute message to each service to be consumed
- **Redis**: In-memory cache to store published messages and processed files information
- **Data Migration Service**: Consumes file info from message queue, then downloads it from Amazon S3 and uploads it to GCS
- **Publisher**: A HTTP Web server that publishes the file info to message queue. Also provides APIs for managing the cache

## Use Case Analysis

This section describes key functional and non-functional requirements of the system, and the corresponding technology decisions made to fulfill them effectively.



### ✅ Use Case 1: The service must be able to process multiple files concurrently

**Requirement:**
The system should handle multiple file transfers (from Amazon S3 to Google Cloud Storage) at the same time to maximize throughput and minimize the processing time

**Solution:**
Chose **Golang** for its lightweight **goroutines**, which allows the application to manage thousands of concurrent operations



### ✅ Use Case 2: The system should be horizontally scalable and avoid conflicts between pods in K8s

**Requirement:**
In a distributed deployment with multiple instances of consumers (the data migration services), the system must avoid duplicating work and should be scalable across pods and nodes.

**Solution:**
Chose **Google Cloud Pub/Sub** as a message queue to ensure message coordination and reliable distribution. Each file event is published once, and consumers independently subscribe and process messages.



### ✅ Use Case 3: Avoid reprocessing files that were already transferred

**Requirement:**
The system should not repeatedly process the same files, which could waste bandwidth and storage costs.

**Solution:**
Chose **Redis** as a fast, in-memory cache to **track the state of each file** that has been published or consumed. Before publishing or processing, the service checks Redis to verify if the file was already handled.



### ✅ Use Case 4: Detect changes in already-processed files

**Requirement:**
If a file has already been processed, but it was later updated in Amazon S3 (it has the same file name as before), the system should detect the change and process it again.

**Solution:**
Cache the file’s **ETag** (an identifier that changes when the file content changes) in Redis along with the file key. If the file has a **different ETag**, which means that the content was modified and the file should be reprocessed.



### ✅ Use Case 5: For local development and testing, the approach must minimize the costs from cloud services

**Requirement:**
Be able to test the full pipeline locally without needing access to cloud infrastructure (apart from Amazon S3 and GCS).

**Solution:**
- Use the **Google Cloud Pub/Sub Emulator** as the alternative for **Google Cloud Pub/Sub**
- Use **Docker Compose** to run local instances of services like Pub/Sub, Redis, Data Migration Service and Message Publisher service for development and testing
- Use **Minikube** to deploy all the instances to test the availability and the scalability


### ✅ Use Case 6: Modular and maintainable architecture

**Requirement:**
The system should be easy to extend, test, and maintain. Different components may need to be scaled independently.

**Solution:**
Use a **microservices architecture** to separate concerns:
- The **publisher** detects new or updated files and publishes events.
- The **consumer** handles downloading from S3 and uploading to GCS.
- Redis acts as the **shared state manager**.


## Demo


