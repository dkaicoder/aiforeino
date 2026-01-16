# 表结构：user_player
**表注释**：用户玩法记录表  
**所属数据库**：lw_match

## 字段列表
| 字段名 | 类型 | 主键 | 自增 | 非空 | 默认值 | 注释 |
|--------|------|------|------|------|--------|------|
| id | bigint | ✅ | ✅ | ✅ | - | - |
| uid | int | ❌ | ❌ | ❌ | 0 | uid（关联user表的id字段） |
| gift_id | int | ❌ | ❌ | ❌ | 0 | 礼物ID |
| gift_num | int | ❌ | ❌ | ❌ | 0 | 礼物数量 |
| gift_name | varchar(255) | ❌ | ❌ | ❌ | NULL | - |
| gift_image | varchar(255) | ❌ | ❌ | ❌ | 0 | 礼物图标 |
| player_id | int | ❌ | ❌ | ❌ | 0 | 玩法ID |
| player_detail_id | int | ❌ | ❌ | ❌ | 0 | 详情 |
| create_at | timestamp | ❌ | ❌ | ✅ | CURRENT_TIMESTAMP | - |
| update_at | timestamp | ❌ | ❌ | ✅ | CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | - |
| total_diamond | int | ❌ | ❌ | ❌ | NULL | 礼物总价 |
| gift_diamond | int | ❌ | ❌ | ❌ | NULL | 礼物单价 |
| once_diamond | int | ❌ | ❌ | ❌ | 0 | 单次消耗 |
| round | int | ❌ | ❌ | ✅ | 0 | 轮数 |
| room_id_int | bigint | ❌ | ❌ | ✅ | 0 | - |
| detail_reward_id | int | ❌ | ❌ | ❌ | NULL | - |
| batch_id | char(50) | ❌ | ❌ | ❌ | NULL | 抽奖批次id |

## 索引信息
- PRIMARY KEY：id（BTREE）
- idx_player_id：player_id（普通索引）
- idx_batch_id：batch_id（BTREE）

## 关联关系
- 关联user表：uid（本表）→ id（关联表）


# 表结构：user
**表注释**：用户基础信息表  
**所属数据库**：lw_match

## 字段列表
| 字段名 | 类型 | 主键 | 自增 | 非空 | 默认值 | 注释 |
|--------|------|------|------|------|--------|------|
| id | int | ✅ | ✅ | ✅ | - | 主键ID（其他表的uid字段均关联此字段） |
| phone | varchar(32) | ❌ | ❌ | ✅ | '' | 手机号 |
| account | varchar(32) | ❌ | ❌ | ❌ | '' | 账号 |
| accid | varchar(32) | ❌ | ❌ | ❌ | '' | - |
| ename_uid | bigint | ❌ | ❌ | ❌ | 0 | 别名 |
| room_id | varchar(32) | ❌ | ❌ | ❌ | '' | 房间ID |
| room_id_int | bigint | ❌ | ❌ | ❌ | 0 | - |
| nickname | varchar(64) | ❌ | ❌ | ❌ | '' | 用户昵称 |
| avatar | varchar(255) | ❌ | ❌ | ✅ | '' | 头像 |
| avatar_audit | varchar(255) | ❌ | ❌ | ❌ | NULL | 审核头像 |
| avatar_before | varchar(255) | ❌ | ❌ | ✅ | '' | - |
| avatar_status | tinyint(1) | ❌ | ❌ | ❌ | 0 | 1-待审核 2-审核失败 3-审核通过 |
| sex | tinyint(1) | ❌ | ❌ | ✅ | 0 | 1-男 2-女 |
| state | int | ❌ | ❌ | ❌ | 1 | 用户状态 1-正常 2-封禁 3-注销 |

## 索引信息
- PRIMARY KEY：id（BTREE）
- UNIQUE KEY：unique_accid（accid，BTREE）
- 普通索引：idx_ename（ename_uid，BTREE）、idx_phone（phone）

## 关联关系
- 核心关联：其他表的uid字段均关联本表的id字段