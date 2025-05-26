# 数据库表设计

数据库：`cn_udm_dbp`

```sql
-- CREATE SCHEMA cn_udm_dbp CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

## 表设计

`dbp_field` 元数据字段表：
```sql
create table cn_udm_dbp.dbp_fields
(
	id bigint auto_increment primary key comment '主键id',
	created_at datetime(3) null comment '创建时间',
	updated_at datetime(3) null comment '修改时间',
	deleted_at datetime(3) null comment '删除时间',
	field varchar(512) null comment '字段',
	json_path varchar(512) null comment 'json path',
	type varchar(512) null comment '字段类型（bool、float、int、string、enum、json）',
	length int unsigned null comment '字段长度',
	name varchar(512) null comment '字段长度',
	nullable tinyint(1) null
)ENGINE=InnoDB  comment '行为日志字段表';
create index idx_dbp_fields_deleted_at
	on cn_udm_dbp.dbp_fields (deleted_at);
```

```sql
insert into dbp_fields(created_at, updated_at, field, json_path, type, length, name, nullable)
values (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'distinct_id', 'distinct_id', 'string', 256, '用户id', false),
       (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '_track_id', '_track_id', 'long', 256, '_track_id', false),
       (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'time', 'time', 'long', 64, '时间', false),
       (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'user_type', 'properties.user_type', 'enum', 32, '用户类型', false),
       (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'event', 'event', 'string', 64, '事件', false),
       (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'project_id', 'properties.project_id', 'string', 128, '项目名', false),
       (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'platform', 'properties.platform', 'enum', 128, '平台', false),
       (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'page_id', 'properties.page_id', 'string', 128, '页面名称', false),
       (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'current_page_url', 'properties.current_page_url', 'string', 1024, '事件当前页面地址', true),
       (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'refer_page_id', 'properties.refer_page_id', 'string', 128, '来源页面名称', true),
       (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'referrer_material', 'properties.referrer_material', 'string', 128, '来源元素名称', true),
       (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'app_id', 'properties.app_id', 'string', 512, '应用唯一标识', true),
       (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'app_name', 'properties.app_name', 'string', 128, '应用名称', false),
       (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'app_version', 'properties.app_version', 'string', 256, '应用版本', true),
       (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'latitude', 'properties.latitude', 'float', 128, '经度', true),
       (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'longitude', 'properties.longitude', 'float', 128, '纬度', true),
       (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'carrier', 'properties.carrier', 'string', 128, '运营商', true),
       (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'os', 'properties.os', 'string', 128, '操作系统', true),
       (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'os_version', 'properties.os_version', 'string', 128, '系统版本', true),
       (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'network_type', 'properties.network_type', 'string', 128, '网络类型', true),
       (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'material_id', 'properties.material_id', 'string', 512, '素材id', false),
       (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'material_list', 'properties.material_list', 'string', 512, '素材列表', true),
       (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'material_position', 'properties.material_position', 'int', 512, '素材位置', true),
       (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'material_attr', 'properties.material_attr', 'json', 512, '素材类型', true);
       
```

`dbp_events` 事件表：
```sql
create table cn_udm_dbp.dbp_events
(
	id bigint unsigned auto_increment primary key comment '主键id',
	created_at datetime(3) null comment '创建时间',
	updated_at datetime(3) null comment '修改时间',
	deleted_at datetime(3) null comment '删除时间',
	event varchar(512) null comment '事件',
	description varchar(512) null comment '描述'
)ENGINE=InnoDB  CHARACTER SET utf8mb4 comment '行为日志事件表';

create index idx_dbp_events_deleted_at
	on cn_udm_dbp.dbp_events (deleted_at);
```

```sql
-- 初始化事件
insert into dbp_events(event, created_at, updated_at, description)
values ('page_view', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '页面浏览'),
       ('element_click', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '页面浏览'),
       ('scan', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '扫码'),
       ('element_expose', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '元素曝光'),
       ('app_uninstall', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'app卸载'),
       ('app_start', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'app启动'),
       ('app_end', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'app退出'),
       ('pay_order', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '支付订单成功'),
       ('signup', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '用户登录') ;
```

`dbp_event_fields` 事件字段表：
```sql
create table cn_udm_dbp.dbp_event_fields
(
	id bigint unsigned auto_increment primary key,
	created_at datetime(3) null,
	updated_at datetime(3) null,
	deleted_at datetime(3) null,
	event longtext null,
	field longtext null,
	nullable tinyint(1) null
)ENGINE=InnoDB  CHARACTER SET utf8mb4 comment '行为日志事件字段配置表';;

create index idx_dbp_event_fields_deleted_at
	on cn_udm_dbp.dbp_event_fields (deleted_at);
```

`dbp_field_enum_values`字段枚举值表：
```sql
create table cn_udm_dbp.dbp_field_enum_values
(
    id bigint auto_increment
        primary key,
    field varchar(128) not null comment '字段',
    enum_value varchar(128) not null,
    value_name varchar(256) not null comment '枚举值中文名',
    created_at datetime default CURRENT_TIMESTAMP not null comment '创建时间',
    updated_at datetime default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '修改时间',
    deleted_at datetime null comment '删除时间'
)ENGINE=InnoDB  CHARACTER SET utf8mb4 comment '枚举字典表';

create unique index uidx_field_value
    on cn_udm_dbp.dbp_field_enum_values (field, enum_value);
```

```sql
-- 新增user_type枚举值
insert into dbp_field_enum_values(field, enum_value, value_name, created_at, updated_at)
values 
       ('platform', 'APP', '应用', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
       ('platform', 'Wechat_Public', '微信公众号', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
       ('platform', 'Wechat_Applet', '微信小程序', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
       ('platform', 'Alipay_Applet', '阿里小程序', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
       ('platform', 'h5', '单链', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
       ('platform', 'Web', '系统服务', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
       ('platform', 'Other', '其他', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
```