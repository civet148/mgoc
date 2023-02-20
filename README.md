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

# quick start

```go
package main
import (
    "github.com/civet148/log"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo/options"
)
func main() {
	e, err := NewEngine("mongodb://root:123456@192.168.20.108:27017/test?authSource=admin")
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	var students []*Student
	err = e.Model(&students).
            Table("student_info").
            Options(&options.FindOptions{}).
            Filter(bson.M{
                "name": "lory",
            }).
            Desc("created_time").
            Query()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	for _, s := range students {
		log.Infof("student %+v", s)
	}
}
```