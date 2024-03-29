# Demo 说明

## Demo 涉及表格说明

### cars,记录当前车辆地址，状态信息
- id: 唯一标识,BIGINT
- location-x: double
- location-y:double
- status: idle/running/off,varchar
- createtime: DATETIME
- updatetime: DATETIME

### orders，记录乘客订单信息
- id:唯一标识 BIGINT
- car_id: 当前用车 BIGINT
- from_x:double
- from_y:double
- to_x:double
- to_y:double
- status:running/finished/waiting
- createtime: DATETIME
- updatetime: DATETIME


### nearcars, 记录 flink 计算出来的最近 N 辆车信息
- order_id: 订单ID
- cars:string, json, 表示推荐车辆 ID ，如 [1,2,3,4]
- consumed: int, 是否被消费过，默认 0 未消费过。1 表示该数据已经被处理过。
- create_time: Datetime
  
## 脚本监听事件说明
Demo 主要通过模拟实现以下几个事件，来模拟本项目在网约车实时调度。

### 司机
#### 司机上班
- 将司机信息注册/更新到 cars 表格中(insert/update）, 车辆状态设置为 on，位置信息更新

 ``` TIDB
  insert into cars set id=?,location, for duplicate.. set status=on,location..
  ``` 

TODO: 该脚本可以模拟初始化大量司机。

#### 司机下班
1. 检查是否正在订单中，若否，拒绝
2. 司机下班：将当前车辆状态设置为 off

``` TIDB
begin()
 select * from cars where id=xx for update
 // check status
 update cars set status=off where id=xxx
commit
```

### 乘客
#### 乘客下单
1. 将当前乘客的订单信息入库，订单状态为 waiting.
   后端异步线程：
   Flink 实时计算合适 车辆，并将下单数据更新到 TiDB
    - 更新当前车辆状态为 running
    - 更新订单信息 order 到 running 

Flink 实时推荐适合 车辆，并将下单数据更新到 TiDB:
```
-- 设置下时区
SET 'table.local-time-zone' = 'Asia/Shanghai';
SET execution.result-mode=tableau;

-- 源表
CREATE TABLE tidb_cars ( 
    `id` INT NOT NULL, 
     location_x DOUBLE,
     location_y DOUBLE,
     status STRING,
     create_time TIMESTAMP,
     update_time TIMESTAMP,
     PRIMARY KEY (`id`) NOT ENFORCED) 
WITH ( 
    'connector' = 'tidb-cdc',
    'hostname' = '127.0.0.1',
    'port' = '4000',
    'tikv.grpc.timeout_in_ms' = '1000',
    'tikv.pd.addresses' = '127.0.0.1:2379', 
    'username' = 'flinkuser',
    'password' = 'flinkpwd',
    'database-name' = 'test',
    'table-name' = 'cars');

-- 源表
CREATE TABLE tidb_orders ( 
     order_id INT NOT NULL, 
     car_id INT,
     from_x DOUBLE,
     from_y DOUBLE,
     to_x DOUBLE,
     to_y DOUBLE,
     status STRING,
     create_time TIMESTAMP,
     update_time TIMESTAMP,
     PRIMARY KEY (`order_id`) NOT ENFORCED) 
WITH ( 
    'connector' = 'tidb-cdc',
    'hostname' = '127.0.0.1',
    'port' = '4000',
    'tikv.grpc.timeout_in_ms' = '1000',
    'tikv.pd.addresses' = '127.0.0.1:2379', 
    'username' = 'flinkuser',
    'password' = 'flinkpwd',
    'database-name' = 'test',
    'table-name' = 'orders');
 
 -- 结果表
 CREATE TABLE tidb_nearcars ( 
     order_id INT,
     cars STRING,
     consumed INT,
     create_time TIMESTAMP,
     PRIMARY KEY (`order_id`) NOT ENFORCED) 
WITH ( 
    'connector' = 'jdbc',
    'url' = 'jdbc:mysql://localhost:4000/test',
    'table-name' = 'nearcars',
     'username' = 'root',
     'password' = ''
);   

// 提交 Streaming SQL作业，提交后作业会一直running, 不管是订单状态更新了，还是车辆信息更新了，都会输出最新的推荐结果
INSERT INTO tidb_nearcars
SELECT order_id, concat('[', LISTAGG(car_id), ']'),  CAST(0 as INT) as consumed, LOCALTIMESTAMP as create_time
FROM 
(SELECT order_id, CAST(car_id as STRING) AS car_id, distance, rownum
FROM ( -- 按距离最短 取 TOP 10 
   SELECT order_id, car_id, distance,
         ROW_NUMBER() OVER (PARTITION BY order_id ORDER BY distance) AS rownum 
   FROM (
      SELECT o.order_id, c.id as car_id,
      ROUND(SQRT((POWER(o.from_x - c.location_x, 2) + POWER(o.from_y - c.location_y, 2) ) ), 2) as distance
      FROM 
         tidb_orders o LEFT JOIN 
         tidb_cars c ON o.status = 'waiting' AND c.status = 'idle' ) t
   ) 
WHERE rownum <= 10) top_t
GROUP BY order_id;

```

2. 将下单数据更新到 TiDB
    - 更新当前车辆状态为 running
    - 创建订单信息 order

``` TIDB 
 begin()
 // check and update car's status
 update cars set status=running where status=on and id=xx
 // init order
 insert into order set ...

```

#### 乘客下车
更新 TiDB
- 更新订单状态
- 更新司机状态及车辆位置

``` TIDB
begin()
// update order
update orders set status=closed
// update car's status & location
update cars..
commit()
```



## 数据流向架构
![./img/data.jpg](./img/data.jpg)

1. 所有数据更新都写入到 TiDB,这里 TIDB 提供 OLTP.
2. TiDB 数据通过 Flink CDC 到 Pravega 和 DataLake
   1. 实时分析通过 Pravega 提供给 Flink 进行计算
      1. Flink 计算完以后在将结果放入 pravega
      2. 后端有个线程专门处理 pravega 里面的车辆推荐结果，去 tidb 下单。
   2. 离线分析通过 Data Lake.
      1. WEB 前端数据从 DataLake 获取（离线分析）

TiDB CDC 数据入湖 SQL 
```
SET execution.checkpointing.interval=3s;
-- 源表
CREATE TABLE tidb_cars ( 
    `id` INT NOT NULL, 
     location_x DOUBLE,
     location_y DOUBLE,
     status STRING,
     create_time TIMESTAMP(3),
     update_time TIMESTAMP(3),
     PRIMARY KEY (`id`) NOT ENFORCED) 
WITH ( 
    'connector' = 'tidb-cdc',
    'hostname' = '127.0.0.1',
    'port' = '4000',
    'tikv.grpc.timeout_in_ms' = '1000',
    'tikv.pd.addresses' = '127.0.0.1:2379', 
    'username' = 'flinkuser',
    'password' = 'flinkpwd',
    'database-name' = 'test',
    'table-name' = 'cars');

-- 源表
CREATE TABLE tidb_orders ( 
     order_id INT NOT NULL, 
     car_id INT,
     from_x DOUBLE,
     from_y DOUBLE,
     to_x DOUBLE,
     to_y DOUBLE,
     status STRING,
     create_time TIMESTAMP(3),
     update_time TIMESTAMP(3),
     PRIMARY KEY (`order_id`) NOT ENFORCED) 
WITH ( 
    'connector' = 'tidb-cdc',
    'hostname' = '127.0.0.1',
    'port' = '4000',
    'tikv.grpc.timeout_in_ms' = '1000',
    'tikv.pd.addresses' = '127.0.0.1:2379', 
    'username' = 'flinkuser',
    'password' = 'flinkpwd',
    'database-name' = 'test',
    'table-name' = 'orders');
 
 -- 源表
 CREATE TABLE tidb_nearcars ( 
     order_id INT,
     cars STRING,
     consumed INT,
     create_time TIMESTAMP(3),
     PRIMARY KEY (`order_id`) NOT ENFORCED) 
WITH ( 
   'connector' = 'tidb-cdc',
    'hostname' = '127.0.0.1',
    'port' = '4000',
    'tikv.grpc.timeout_in_ms' = '1000',
    'tikv.pd.addresses' = '127.0.0.1:2379', 
    'username' = 'flinkuser',
    'password' = 'flinkpwd',
    'database-name' = 'test',
    'table-name' = 'nearcars');


CREATE TABLE hudi_orders(
     order_id INT NOT NULL, 
     car_id INT,
     from_x DOUBLE,
     from_y DOUBLE,
     to_x DOUBLE,
     to_y DOUBLE,
     status STRING,
     create_time TIMESTAMP(3),
     update_time TIMESTAMP(3),
     PRIMARY KEY (`order_id`) NOT ENFORCED
)
WITH (
  'connector' = 'hudi',
  'path' = '///home/ec2-user/data/orders',
    'write.tasks' = '1',
  'compaction.tasks' = '1',
  'table.type' = 'MERGE_ON_READ'
 );
 
 CREATE TABLE hudi_cars(
     `id` INT NOT NULL, 
     location_x DOUBLE,
     location_y DOUBLE,
     status STRING,
     create_time TIMESTAMP(3),
     update_time TIMESTAMP(3),
     PRIMARY KEY (`id`) NOT ENFORCED
)
WITH (
  'connector' = 'hudi',
  'path' = '///home/ec2-user/data/cars',
    'write.tasks' = '1',
  'compaction.tasks' = '1',
  'table.type' = 'MERGE_ON_READ'
 );
 
 CREATE TABLE hudi_nearcars(
     order_id INT,
     cars STRING,
     consumed INT,
     create_time TIMESTAMP(3),
     PRIMARY KEY (`order_id`) NOT ENFORCED
)
WITH (
  'connector' = 'hudi',
  'path' = '///home/ec2-user/data/nearcars',
  'table.type' = 'MERGE_ON_READ',
  'write.tasks' = '1',
  'compaction.tasks' = '1',
  'read.streaming.enabled' = 'true'
 );
 
 -- 实时同步作业，将TiDB中的表全增量地同步到 hudi 中
 begin statement set;
  insert into hudi_orders select * from tidb_orders;
  insert into hudi_cars select * from tidb_cars;
  insert into hudi_nearcars select * from tidb_nearcars;
 end;

```


# 目前 Demo 走 HTTP 协议
## Web 启动条件：
数据库条件 ：
- 目前默认连接方式  "root:@tcp(127.0.0.1:4000)/test?charset=utf8"，可根据需要修改 main 函数对应内容。
- 创建表：./data/create_table.sql
- 启动方式：
```
cd ../web
go build
./web 
```
## 接口说明
### 司机上班
司机状态变为 running.
上班条件为：当前司机状态处于 offline.
```
// PUT /car/{id}?x=${x}&y=${y}
// create a car or online a car
// Example: curl -X PUT "http://localhost:7998/car/1?x=12&y=13"
```

### 司机下班
司机状态变为 offline. 
下班条件为：当前司机状态为 idle. 

```
// DELETE /car/{id}
// offline a car
// Example: curl -X DELETE "http://localhost:7998/car/1"
```

### 乘客下单
下单后，当前订单状态为 waiting, 等待接单

```
// PUT /order?fromx=?&fromy=?&tox=?toy=?
// a new order
// Example: curl -X PUT "http://localhost:7998/order?fromx=1&fromy=2&tox=12&toy=13"
// {"Id":3,"CarID":0,"FromX":1,"FromY":2,"ToX":12.2,"ToY":13.3,"Status":"waiting","CreateTime":"2022-01-06 21:40:23","UpdateTime":"2022-01-06 21:40:23"}
```


### 结束订单
结束订单后，订单状态变为 finished, 司机状态变为 idle.
结束条件为：当前订单状态为 running.
```
// DELETE /order/{orderID}
// Finish an order
// Example: curl -X DELETE "http://localhost:7998/order/1"
```
### 随机获取可用 location(调试用)

```
// GET /location
// get a location randomly
// Example: curl -X GET "http://localhost:7998/location"
// {"Id":20424,"X":40.25,"Y":115.53}
```
