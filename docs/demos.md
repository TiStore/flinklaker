# Demo 说明

## Demo 涉及表格说明

### cars,记录当前车辆地址，状态信息
- id: 唯一标识,BIGINT
- location-x: int
- location-y:int
- status: idle/running/off,varchar
- createtime: DATETIME
- updatetime: DATETIME

### orders，记录乘客订单信息
- id:唯一标识 BIGINT
- car_id: 当前用车 BIGINT
- from_x:int
- from_y:int
- to_x:int
- to_y:int
- status:on/closed
- createtime: DATETIME
- updatetime: DATETIME


## 脚本监听事件说明
Demo 主要通过模拟实现以下几个事件，来模拟本项目在网约车实时调度。

### 司机
#### 司机上班
- 将司机信息注册/更新到 cars 表格中(insert/update）, 车辆状态设置为 on，位置信息更新

 ``` TIDB
  insert into cars set id=?,location, for duplicate.. set status=on,location..
  ``` 


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
1. 根据当前乘客的位置信息，使用 flink 计算出适合他的车辆信息
```
TODO: Flink SQL
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

