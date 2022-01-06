CREATE TABLE `cars` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `location_x` double DEFAULT NULL,
  `location_y` double DEFAULT NULL,
  `status` varchar(20) DEFAULT NULL,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) 
);

create table orders (
  order_id int(11) not null AUTO_INCREMENT,
  car_id int(11) default 0,
  from_x double default 0,
  from_y double default 0,
  to_x double default 0,
  to_y double default 0,
  status varchar(20),
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`order_id`) 
);

create table nearcars(
  id int(11) not null AUTO_INCREMENT,
  order_id int(11),
  cars varchar(1000),
  consumed int,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  key(consumed, order_id),
  PRIMARY KEY(`id`)
);

create table locations(
  id int(11) not null AUTO_INCREMENT,
  `x` double DEFAULT NULL,
  `y` double DEFAULT NULL,
  PRIMARY KEY(`id`)
);