docker build -t justin2997/rabbit .
sleep 30

IMAGE="justin2997/rabbitmq"
TAG="latest"
CONTAINER="rabbitmq-server"

docker run -d \
--hostname $CONTAINER \
--name $CONTAINER \
-p 15672:15672 \
${IMAGE}:${TAG}