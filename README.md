# mgoc
mongodb ORM client

# deploy mongodb

- configuration

```shell

# make mongodb directories on the host machine
$ mkdir -p /data/mongodb/conf /data/mongodb/db

# view mongodb config file
$ cat /data/mongodb/conf/mongod.conf

# mongod.conf

# for documentation of all options, see:
#   http://docs.mongodb.org/manual/reference/configuration-options/

# Where and how to store data.
storage:
  dbPath: /var/lib/mongodb
  journal:
    enabled: true
#  engine:
#  mmapv1:
#  wiredTiger:

# where to write logging data.
systemLog:
  destination: file
  logAppend: true
  path: /var/log/mongodb/mongod.log

# network interfaces
net:
  port: 27017
  bindIp: 0.0.0.0


# how the process runs
processManagement:
  timeZoneInfo: /usr/share/zoneinfo

#security:

#operationProfiling:

#replication:

#sharding:

## Enterprise-Only Options:

#auditLog:

#snmp:

```
- set up mongodb docker container

```shell

# start mongodb by docker
$ docker run -p 27017:27017 --restart always  --log-opt max-size=500m -v /data/mongodb/conf/mongod.conf:/etc/mongod.conf -v /data/mongodb/db:/data/db --name mongodb -d mongo:4.4.10
```

- create account

```shell
$ docker exec -it mongodb mongo admin
> use admin
> db.createUser({user:"root", pwd: "123456", roles: ["root"]})
```