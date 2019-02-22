# EventSource
docker build -t justin2997/knative-rabbitmq-event-source direct-event-source/.
docker push justin2997/knative-rabbitmq-event-source

# Channel EventSource
docker build -t justin2997/knative-rabbitmq-channel-event-source channel-event-source/.
docker push justin2997/knative-rabbitmq-channel-event-source

# Consummer
docker build -t justin2997/knative-rabbitmq-logger-pods pods/.
docker push justin2997/knative-rabbitmq-logger-pods

# Producer
docker build -t justin2997/knative-rabbitmq-producer producer/.
docker push justin2997/knative-rabbitmq-producer

# Logger
docker build -t justin2997/knative-rabbitmq-logger logger/.
docker push justin2997/knative-rabbitmq-logger