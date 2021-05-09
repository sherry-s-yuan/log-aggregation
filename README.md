# Log Aggregation

## Setup

### Docker Setup
Run the following command to setup docker development environment

```
curl -sSL https://raw.githubusercontent.com/bitnami/bitnami-docker-kafka/master/docker-compose.yml > docker-compose.yml

docker-compose -f docker-compose.yml up
```

Note the kafka environment should have those variable during development

```
    - KAFKA_LISTENERS=PLAINTEXT://:9092
    - KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://127.0.0.1:9092
    - KAFKA_ZOOKEEPER_CONNECT=zookeeper:2181
    - ALLOW_PLAINTEXT_LISTENER=yes
```

### Go Module Setup

```
go mod init log-aggregation
go mod tidy
```
### Elasticsearch Setup
To change vm.max_map_count
```
wsl -d docker-desktop
sysctl -w vm.max_map_count=262144
exit
```

### To Access docker backend files
```
docker run -it --privileged --pid=host debian nsenter -t 1 -m -u -i sh
```

## Commands

### Send data to producer
To send new logs to system on Windows:
```
curl -X POST http://127.0.0.1:8082/api/v1/logs -H "Content-type:application/json" -d "{ \"levelname\":\"ERROR\", \"message\":\"Docker starting...\", \"timestamp\":\"2021-05-09T14:48:00.000Z\", \"filename\":\"transport.go\" }"
```

To send new logs to system on Linux:
```
curl --location --request POST '127.0.0.1:3000/api/v1/comments' 
--header 'Content-Type: application/json' \
--data-raw '{ "level":"info", "message":"Docker starting..." }'
```

### Kubernetes deployment
```
kompose convert
```

Change "imagePullPolicy" option in deployment.yaml to correspond to "Never" when the container doesn't rely on pulling image from docker hub.

Change "apiVersion" option in networkpolicy.yaml to "networking.k8s.io/v1".

Change "creationTimestamp" in persistentvolumeclaim.yaml to "2021-05-08T10:09:47Z" (can be any specific time)

Change "_" to "-" in metadata->name under networkpolicy

Then run the following to create pods

```
kubectl apply -f .
```

To check the pod status and service, run the following

```
kubectl get po
kubectl get svc
```

## Useful Links
https://github.com/bitnami/bitnami-docker-kafka/blob/master/README.md

https://pkg.go.dev/github.com/shopify/sarama

