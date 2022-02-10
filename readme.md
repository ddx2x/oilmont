# ddx2x

## 注意
开发人员开发新的应用后使用 make gen-dockerfile生成dockerfile再提交

## 构建
make docker

## docker运行所有
make run


## run etcd on local
```
docker run --restart=always -tid -p2379:2379 --name=etcd dengzitong/etcd:latest etcd -name etcd1 \
-advertise-client-urls=http://0.0.0.0:2379 \
-listen-client-urls=http://0.0.0.0:2379 \
-initial-cluster-state \
new
```

## run mongodb on local
```
docker run --name=mongo -tid -p27017:27017 mongo --bind_ip=0.0.0.0 --port=27017 --replSet=rs0

docker exec -ti mongo mongo
rs.initiate(
   {
      _id: "rs0",
      members: [
         { _id: 0, host : "127.0.0.1:27017" },
      ]
   }
)
```