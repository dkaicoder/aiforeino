# 用户玩法记录表
【核心用途】：记录用户在游戏内的玩法参与、礼物获取/消耗等行为数据，用于统计用户游戏行为、核算礼物相关收益与消耗、追溯抽奖批次明细。
数据库：lw_match
表名：user_player
表注释：用户玩法记录表
核心字段：id（主键自增）、uid（关联user表id）、gift_id（礼物ID）、gift_num（礼物数量）、gift_name（礼物名称）、gift_image（礼物图标）、player_id（玩法ID）、player_detail_id（玩法详情ID）、create_at（创建时间）、update_at（更新时间）、total_diamond（礼物总价）、gift_diamond（礼物单价）、once_diamond（单次消耗）、round（轮数）、room_id_int（房间ID）、detail_reward_id（奖励详情ID）、batch_id（抽奖批次id）
索引信息：主键索引（id）、普通索引（idx_player_id：player_id；idx_batch_id：batch_id）
关联关系：本表uid字段关联lw_match数据库user表的id字段

# 用户基础信息表
【核心用途】：存储平台用户的核心基础信息，作为全系统用户数据的关联核心，支撑各类业务模块的用户身份校验、信息展示与权限管控。
数据库：lw_match
表名：user
表注释：用户基础信息表
核心字段：id（主键自增，其他表uid字段均关联此字段）、phone（手机号）、account（账号）、accid（唯一标识）、ename_uid（用户别名）、room_id（房间ID）、room_id_int（房间数字ID）、nickname（用户昵称）、avatar（用户头像）、avatar_audit（审核中头像）、avatar_before（历史头像）、avatar_status（头像审核状态）、sex（性别）、state（用户账号状态）
索引信息：主键索引（id）、唯一索引（unique_accid：accid）、普通索引（idx_ename：ename_uid；idx_phone：phone）
关联关系：作为核心关联表，其他表的uid字段均关联本表的id字段

# 账单记录表
【核心用途】：记录用户的各类资金交易行为（收支、兑换、红包等），用于财务对账、用户消费明细统计、资金流向追溯与账务审计。
数据库：lw_match
表名：bill_record
表注释：账单记录表
核心字段：id（主键ID）、uid（用户ID）、type（交易类型）、only_id（唯一标识）、money_type（资金类型）、money（交易金额）、money_unit（收支类型）、update_at（更新时间）、create_at（创建时间）、remark（交易说明）、ext_id（扩展ID，关联礼物/游戏ID）、room_id_int（房间ID）、target_uid（对方用户ID）、origin_money（原始金额）、to_money（转化后金额）、target_uids（多方对方用户ID）
索引信息：主键索引（id）
关联关系：字段uid关联lw_match数据库user表的id字段；字段ext_id关联礼物/游戏相关表字段（gift_id、game_id_int）；字段room_id_int关联房间相关字段