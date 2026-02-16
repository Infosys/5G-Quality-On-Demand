# 5G_Quality_on_Demand
# Free5GC v4.0.1 Custom Repository

## Description
The repository consists of free5gc-compose, source codes of various 5G NFs, solution guide. The source codes of NEF, PCF, SMF Network Functions (NFs) where we have done development and improvements are as per 3GPP R17 & R18 specifications. Rest of the NF source codes are default. Here, we can deploy the free5GC by creating custom images and optimize the source codes.Repository consist two free5gc-compose, use free5gc-compose-UERANSIM for intergrating with ueransim simulator and free5gc-compose-external_gNB for integrating with actual radio.

## Pre-Requisites

The solution is tested with below compute and storage resources:

1. CPU: 16vCPU
2. Memory: 16GB
3. Storage: 250GB

While integrating with UERANSIM make sure to use below versions:

1. OS: Ubuntu 22
2. Docker version: 27.5.0
3. GTP5G version: 0.9.11

While integrating with an actual radio make sure to use below versions:

1. OS: Ubuntu 22
2. Docker version: 27.5.0                
3. GTP5G version: 0.9.14

Make sure to pull the below images
1. docker pull free5gc/ueransim:latest
2. docker pull mongo:3.6.8
3. docker pull free5gc/webui:v4.0.1
4. docker pull free5gc/n3iwue:latest

Else you need not modify the above default image names in docker-compose.yaml file which can be also pulled while the executing the YAML file.


## Install GTP5G

Follow the steps below to install gtp5g
1. Clone the gtp5g repo

    `git clone --branch v0.9.11 --depth 1 https://github.com/free5gc/gtp5g.git`

2. Execute the below commands one by one inside gtp5g directory

    `sudo apt install gcc-12 g++-12`

    `make clean && make`

    `sudo make install`

3. To check gtp5g is running

    `lsmod | grep gtp`

4. Enable QoS

    `echo 1 >  /proc/gtp5g/qos `

5. Check the version of gtp5g

    `modinfo gtp5g`


## Build Source Code as Docker Images

1. Clone the repository and navigate to free5gc-compose directory where we have the config files for all the containers that is mentioned in the docker-compose.yaml file.
2. Inside this directory we have docker-compose.yaml file where we can mention the image names that we have build. For all the 5G NFs we can build the images from source code as mentioned in the below steps. 

    `image: <Image name built from source code>`

3. Install the latest version of Go Lang using snap or apt package manager, we have used go version go1.25.4 linux/amd64.
If its already installed check the version. 

    `go version`

4. Navigate to the source code directory to build an image of any NFs.
5. Then, navigate to the respective NF Folder where Makefile and Dockerfile are present

	`sudo make`

6. To build docker image give the below command from the respective NF Folder

	`docker build -t <give an image name of your choice> .`

7. Go to docker-compose.yaml file under freegc-compose-UERANSIM directory while using ueransim simulator and free5gc-compose-external_gNB directory

8. In the yaml file replace the image name with the custom image name that you build for the desired NF

9. Make sure to give the below commands to start/up all the containers where docker-compose.yaml file resides

    `docker compose -f docker-compose.yaml up -d`

10. To down all the conatainers

    `docker compose -f docker-compose.yaml down`


11. While integrating with an actual radio, make necessary changes in free5gc-compose-external_gNB/docker-compose.yaml file in n2,n3 and n6 IPs. A sample docker-compose.yaml file provided in free5gc-compose-external_gNB repo for integrating with an actual radio.

12. Once all the docker containers are up,
visit http://localhost:5000 in your browser to use the Free5GC GUI for provisioning SIM details and configurations. 

    username: admin
    password: free5gc

13. Follow Mutual TLS Authentication steps below for authenticating QoS APIs between AF and NEF

## Mutual TLS Authentication

Mutual TLS (mTLS) authentication is a security mechanism that ensures both the client (AF) and the server (NEF) authenticate each other using TLS certificates before establishing a secure connection.

Follow below steps for m-TLS Authentication in NEF after bringing the containers up so that we can check the NEF IP. Run the below command in the client (Application Function/Server) interacting with NEF.
1. Create CA certificate.

    `openssl req -x509 -new -nodes -keyout ca.key -out ca.pem -days 365 -subj "/CN=MyRootCA"`


2. Create Client(AF) key and certificate signed by CA.
    
    Key --> `openssl genrsa -out af_client.key 2048` 
    
    Certificate --> `openssl req -new -key af_client.key -out af_client.csr -subj "/CN=af_client"` 

    Sign the certificate by CA --> `openssl x509 -req -in af_client.csr -CA ca.pem -CAkey ca.key -CAcreateserial -out af_client.pem -days 365`

3. Create Server (NEF) key and certificate signed by CA. 
 
    Key --> `openssl genrsa -out nef.key 2048` 
    
    Certificate --> `openssl req -new -key nef.key -out nef.csr -subj "/CN=<provide the NEF IP Address>"` 

    To check NEF IP Address give command 

    `docker inspect nef`

    Sign the certificate by CA --> `openssl x509 -req -in nef.csr -CA ca.pem -CAkey ca.key -CAcreateserial -out nef.pem -days 365`

4. Add CA certificate and NEF key in free5gc-compose/cert directory.

5. Now go to free5gc-compose/config/nefcfg.yaml file. Here under configuration>sbi>scheme replace http to https. Also under configuration>sbi>tls add below lines one after another 

    `caPem: cert/ca.pem`

    `verifyClient: true`

6. Now down all the containers and start/up it by docker commands mentioned in the above section


## Attach UERANSIM to Free5GC

UERANSIM is a opensource simulator which simulates the actual User Equipment (UE) and Radio Access Network (RAN)

1. Enter into ueransim docker terminal using the below command from CLI

    `docker exec -it ueransim bash`

2. Activate the UERANSIM UE

    `./nr-ue -c config/uecfg.yaml`

3. Once a successfull PDU establishment occurs then the UE will get an IP which can be seen in the uesimtun0 interface by giving below command

    `ip a`

## Testing QoD in Free5GC

Testing of the QoD can be done using IPERF tool between client and server, and necessary Post/Patch requests can be send based on increased traffic requirements. Further details can be referred from 5G QoD solution guide.

## Contributing
Feel free to submit pull requests or open issues.
