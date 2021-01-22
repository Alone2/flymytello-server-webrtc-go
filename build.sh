docker build -t flymytello .
sh stop.sh
docker rm flymytellocont
docker run --network host -v $(pwd)/cert:$homeDocker/cert --env PUBLIC_CHAIN_CERT=cert/cert.crt --env PRIVATE_KEY_CERT=cert/priv.key --name flymytellocont -d -t flymytello
sleep 5
docker exec -it flymytellocont "./setup/setup"
sh stop.sh
