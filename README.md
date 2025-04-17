# Data Migration Service from Amazon S3 to Google Cloud Storage

A project for migrating data from one Amazone S3 to Google Cloud Storage. The system is designed according to microservice architecture, each microservice is containerized by using Docker, and then be deployed to a container orchestration system (Kubernetes - K8s). The system allows scaling up (or down) the data migration service to spreate the load (or focus the load) to multiple applications (or one application).



## System Components
- **PubSub Emulator**: Message queue to coordinate and distribute message to each service to be consumed
- **Redis**: In-memory cache to store published messages and processed files information
- **Data Migration Service**: Consumes file info from message queue, then downloads it from Amazon S3 and uploads it to GCS
- **Publisher**: A HTTP Web server that publishes the file info to message queue. It has a job running in background for every 10 minutes to publish all the file in Amazon S3 and also at the same time, can handle the incoming request for cache management.

## Use Case Analysis

This section describes key functional and non-functional requirements of the system, and the corresponding technology decisions made to fulfill them effectively.



### ✅ Use Case 1: The service must be able to process multiple files concurrently

- **Requirement:**
The system should handle multiple file transfers (from Amazon S3 to Google Cloud Storage) at the same time to maximize throughput and minimize the processing time

- **Solution:**
Chose **Golang** for its lightweight **goroutines**, which allows the application to manage thousands of concurrent operations



### ✅ Use Case 2: The system should be horizontally scalable and avoid conflicts between pods in K8s

- **Requirement:**
In a distributed deployment with multiple instances of consumers (the data migration services), the system must avoid duplicating work and should be scalable across pods and nodes.

- **Solution:**
Chose **Google Cloud Pub/Sub** as a message queue to ensure message coordination and reliable distribution. Each file event is published once, and consumers independently subscribe and process messages.



### ✅ Use Case 3: Avoid reprocessing files that were already transferred

- **Requirement:**
The system should not repeatedly process the same files, which could waste bandwidth and storage costs.

- **Solution:**
Chose **Redis** as a fast, in-memory cache to **track the state of each file** that has been published or consumed. Before publishing or processing, the service checks Redis to verify if the file was already handled.



### ✅ Use Case 4: Detect changes in already-processed files

- **Requirement:**
If a file has already been processed, but it was later updated in Amazon S3 (it has the same file name as before), the system should detect the change and process it again.

- **Solution:**
Cache the file’s **ETag** (an identifier that changes when the file content changes) in Redis along with the file key. If the file has a **different ETag**, which means that the content was modified and the file should be reprocessed.



### ✅ Use Case 5: For local development and testing, the approach must minimize the costs from cloud services

- **Requirement:**
Be able to test the full pipeline locally without needing access to cloud infrastructure (apart from Amazon S3 and GCS).

- **Solution:**
    - Use the **Google Cloud Pub/Sub Emulator** as the alternative for **Google Cloud Pub/Sub**
    - Use **Docker Compose** to run local instances of services like Pub/Sub, Redis, Data Migration Service and Message Publisher service for development and testing
    - Use **Minikube** to deploy all the instances to test the availability and the scalability


### ✅ Use Case 6: Modular and maintainable architecture

- **Requirement:**
The system should be easy to extend, test, and maintain. Different components may need to be scaled independently.

- **Solution:**
    - Use a **microservices architecture** to separate concerns:
    - The **publisher** detects new or updated files and publishes events.
    - The **consumer** handles downloading from S3 and uploading to GCS.
    - Redis acts as the **shared state manager**.


## Local Deployment

### Preparing Cloud Plaform
#### Amazon Web Service
**Create new Access Key**  
    1. Go to https://us-east-1.console.aws.amazon.com/iam/home#/users  
    2. Select an existing user or create a new one if it is needed  
    3. Create an Access Key for ***Local Code***  
    4. Save the Access Key to local storage  

![alt text](https://github.com/phonghaido/cloud-data-migration/blob/main/demo/aws-iam.gif?raw=true)

**Create Amazon S3 Bucket**  
    1. Go to https://eu-central-1.console.aws.amazon.com/s3/home  
    2. Select "Create bucket"  
    3. Type the name of the bucket and click "Create Bucket"  
  
  
  
#### Google Cloud Plaform
**Create new Service Account**  
    1. Go to https://console.cloud.google.com/iam-admin/serviceaccounts?project={your-project-id}  
    2. Create a service account with this roles (*Pub/Sub Admin, Pub/Sub Editor, Pub/Sub Publisher, Pub/Sub Subscriber, Storage Admin, Storage Object Admin, Storage Object Creator, Storage Object User*)  
    3. Create a new Key and store it in JSON format  

![alt text](https://github.com/phonghaido/cloud-data-migration/blob/main/demo/gcp-svc-key.gif?raw=true)

### Run Application Locally With Docker Compose
All the commands to run and deploy the system can be found in ***Makefile***

#### Local DEV
Run (in WSL2 or Linux base OS)
```console
make compose-up
```
Stop
```console
make compose-down
```

#### Setting Up Local K8s Cluster with Minikube
##### Install ***Minikube***
```
# Download the latest Minikube
curl -Lo minikube https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64

# Make it executable
chmod +x ./minikube

# Move it to your user's executable PATH
sudo mv ./minikube /usr/local/bin/

#Set the driver version to Docker
minikube config set driver docker
```

##### Instsall ***kubectl***
```
# Download the latest Kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"

# Make it executable
chmod +x ./kubectl

# Move it to your user's executable PATH
sudo mv ./kubectl /usr/local/bin/
```

##### Run ***Minikube***
```
minikube start
```

##### Use minikube config in ***kubectl***
```
kubectl config use-context minikube
minikube start
```

##### Point your shell to minikube's docker-daemon
```
eval $(minikube -p minikube docker-env)
```

##### Deploy the system to ***minikube***
Before deployment, please add all the environment variables in to `kustomization.yaml` and add the `svc_account.json` to `/secret` directory
```
# Deploy Google PubSub Emulator
make prod-pubsub

# Deploy Redis
make prod-redis

# Deploy Data Migration Service
make prod-migration-mk

# Deploy Publisher
make prod-publisher-mk
```


## Demo

#### Publisher Logs
The first 3 logs are for the requests to retrieve (and delete) the cache, which are the keys and the eTags of all the files that have been published to PubSub message queue
and has been consumed by the data migration service. The `DELETE` request makes sure that all the caches are removed so we can have a fresh start

![alt text](https://github.com/phonghaido/cloud-data-migration/blob/main/demo/publisher.png?raw=true)

#### Data Migration Services Logs
The ***Data-Migration-Service*** has been scaled to 2 services to test the ability to handle multiple requests at the same time without conflicting with each other

**Pod 1**
![alt text](https://github.com/phonghaido/cloud-data-migration/blob/main/demo/pod1.png?raw=true)

**Pod 2**
![alt text](https://github.com/phonghaido/cloud-data-migration/blob/main/demo/pod2.png?raw=true)


## Usage
This part is for using APIs to manage cache data. The Publisher is the HTTP server that handles the incoming requests.
If you deploy the system to ***Minikube*** and don't have any service running on port `:8080`, then you can forward the port that the publisher of ***Minikube*** is running on in to your local machine

```
kubectl port-forward -n data-migration pod/publisher-588b54f974-b67lg 8080:8080 2>&1 >/dev/null &
```

#### APIs
```
# Get all the cache
GET /cache
curl http://localhost:8080/cache -u {admin_username}:{admin_password}

# Get caches by its type (published or consumed)
GET /cache/type?value={published/consumed}
curl http://localhost:8080/cache/type?value=published -u {admin_username}:{admin_password}

# Get cache by its name
GET /cache/name?value={type:Amazon_S3_file_key}
curl http://localhost:8080/cache/name?value=published:Docker Desktop Installer.exe -u {admin_username}:{admin_password}

# Delete all the cache
DELETE /cache
curl -X DELETE http://localhost:8080/cache -u {admin_username}:{admin_password}

# Delete caches by its type (published or consumed)
DELETE /cache/type?value={published/consumed}
curl -X DELETE http://localhost:8080/cache/type?value=published -u {admin_username}:{admin_password}

# Delete cache by its name
DELETE /cache/name?value={type:Amazon_S3_file_key}
curl -X DELETE http://localhost:8080/cache/name?value=published:Docker Desktop Installer.exe -u {admin_username}:{admin_password}