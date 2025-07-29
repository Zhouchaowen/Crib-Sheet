# Transaction

## ACID
- 原子性：
事务是不可分割的最小操作单位，要么全部成功，要么全部失败回滚。
```
假设A账户向B账户转账1000元：
1. 从A账户扣除1000元
2. 向B账户增加1000元

这两个操作必须是原子的，要么都成功，要么都失败。如果只扣款成功但转入失败，
就会导致资金丢失。通过事务的原子性，可以确保这种情况不会发生。
```
- 一致性：
事务执行前后，数据库从一个一致性状态变到另一个一致性状态
```
商品库存和订单处理：
1. 库存数量为100
2. 创建一个10件商品的订单
3. 扣减库存至90

一致性确保了无论事务成功与否，库存数量和订单数量始终保持逻辑一致，
不会出现订单创建了但库存没减少，或库存减少了但订单没创建的情况。
```
- 隔离性：
多个事务并发执行时，事务之间互不干扰
```
多人同时购买同一场电影票：
1. 用户A查看座位5号是空的
2. 用户B同时查看座位5号是空的
3. 用户A选择购买5号座位
4. 用户B也想购买5号座位

通过适当的隔离级别（如REPEATABLE READ），可以防止多个用户同时购买同一个座位。
```
- 持久性：
事务一旦提交，其修改就永久保存在数据库中

## 隔离级别
1. 读未提交（READ UNCOMMITTED）
最低的隔离级别
特点：一个事务可以读取另一个事务未提交的数据
存在问题：脏读、不可重复读、幻读
应用场景示例：
```
事务A:
BEGIN;
UPDATE accounts SET balance = balance - 100 WHERE id = 1;
-- 还未提交

事务B:
BEGIN;
SELECT balance FROM accounts WHERE id = 1;  
-- 此时能看到事务A未提交的修改，如果事务A回滚，
-- 事务B读取到的数据就是脏数据
```
2. 读已提交（READ COMMITTED）
特点：一个事务只能读取另一个事务已经提交的数据
解决了：脏读
仍存在：不可重复读、幻读
应用场景示例：
```
事务A:
BEGIN;
SELECT balance FROM accounts WHERE id = 1;  -- 读取余额为1000
-- 此时事务B执行更新并提交
SELECT balance FROM accounts WHERE id = 1;  -- 再次读取余额为900
COMMIT;

事务B:
BEGIN;
UPDATE accounts SET balance = 900 WHERE id = 1;
COMMIT;
```
3. 可重复读（REPEATABLE READ）
MySQL的默认隔离级别
特点：在同一事务中多次读取同样记录的结果是一致的
解决了：脏读、不可重复读
仍存在：幻读
应用场景示例：
```
事务A:
BEGIN;
SELECT * FROM accounts WHERE balance > 1000;  -- 返回2条记录
-- 此时事务B插入了一条余额为1500的记录
SELECT * FROM accounts WHERE balance > 1000;  -- 仍然返回2条记录
-- 但如果插入新记录，可能会遇到幻读问题
COMMIT;

事务B:
BEGIN;
INSERT INTO accounts(id, balance) VALUES(3, 1500);
COMMIT;
```
4. 串行化（SERIALIZABLE）
最高的隔离级别
特点：事务串行执行，完全避免并发问题
解决了：脏读、不可重复读、幻读
性能最差
应用场景示例：
```
事务A:
BEGIN;
SELECT * FROM accounts WHERE id = 1;
-- 此时事务B试图更新id=1的记录会被阻塞，直到事务A提交或回滚
COMMIT;

事务B:
BEGIN;
UPDATE accounts SET balance = 900 WHERE id = 1;  -- 被阻塞
COMMIT;
```