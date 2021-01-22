sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./cert/priv.key -out ./cert/cert.crt
sudo openssl pkcs12 -export -out ./cert/cert.p12 -inkey ./cert/priv.key -in ./cert/cert.crt -certfile ./cert/cert.crt

sudo chmod 755 cert/ -R
