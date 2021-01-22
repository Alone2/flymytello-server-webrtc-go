# flymytello-server-webrtc-go
## Installation (Debug)
Have docker and openssl installed

```
# generate local cert
sh generateCerts.sh
```
You have to add this cert to Firefox for debugging
```
# build the app
sh build.sh
# start app
sh start.sh
# stop app
sh start.sh
# look at logs 
sh logs.sh

```
