# EventSource Worker
docker build -t justin2997/knative-rabbitmq-event-source-worker event-source-worker/.
docker push justin2997/knative-rabbitmq-event-source-worker

# Producer
docker build -t justin2997/knative-rabbitmq-producer producer/.
docker push justin2997/knative-rabbitmq-producer

# Logger
docker build -t justin2997/knative-rabbitmq-logger logger/.
docker push justin2997/knative-rabbitmq-logger

# Setup Kube
echo "Start RabbitMQ Brocker"
k apply -f rabbitmq-config/.
sleep 10

echo "Start Event Source and Logger"
k apply -f event-source-worker/.
sleep 10

echo "Start Producer"
k apply -f producer/.
sleep 10

echo "Job Done !"