docker-compose up -d


# Connect to any node
docker-compose exec redis1 redis-cli -p 7001

# Check cluster info
cluster info

# Check cluster nodes
cluster nodes

# Test sharding
set ticket:{1} "data1"
set ticket:{2} "data2"
set ticket:{3} "data3"


docker exec -it redis1 redis-cli --cluster create redis1:7001 redis2:7002 redis3:7003 --cluster-replicas 0 --cluster-yes

docker exec -it redis1 redis-cli

docker exec -it redis1 redis-cli cluster nodes