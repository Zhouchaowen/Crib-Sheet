# MySQL面试题

## 基础概念
1. MySQL 中的数据排序是怎么实现的？
2. MySQL 的 Change Buffer 是什么？它有什么作用？
3. 详细描述一条 SQL 语句在 MySQL 中的执行过程
4. MySQL 的存储引擎有哪些？它们之间有什么区别？
5. MySQL 的索引类型有哪些？
6. MySQL InnoDB 引擎中的聚簇索引和非聚簇索引有什么区别？
7. MySQL 中的回表是什么？
8. MySQL 索引的最左前缀匹配原则是什么？
9. MySQL 的覆盖索引是什么？
10. MySQL 的索引下推是什么？

## 索引优化
11. 在 MySQL 中建索引时需要注意哪些事项？
12. MySQL 中使用索引一定有效吗？如何排查索引效果？
13. MySQL 中的索引数量是否越多越好？为什么？
14. 请详细描述 MySQL 的 B+ 树中查询数据的全过程
15. 为什么 MySQL 选择使用 B+ 树作为索引结构？
16. MySQL 中 VARCHAR(100) 和 VARCHAR(10) 的区别是什么？
17. 在什么情况下，不推荐为数据库建立索引？
18. MySQL 中 EXISTS 和 IN 的区别是什么？

## 事务和锁
19. 什么是 Write-Ahead Logging (WAL) 技术？它的优点是什么？MySQL 中是否用到了 WAL？
20. 你们生产环境的 MySQL 中使用了什么事务隔离级别？为什么？
21. MySQL 是如何实现事务的？
22. MySQL 中长事务可能会导致哪些问题？
23. MySQL 中的 MVCC 是什么？
24. 如果 MySQL 中没有 MVCC，会有什么影响？
25. MySQL 中的事务隔离级别有哪些？
26. MySQL 默认的事务隔离级别是什么？为什么选择这个级别？
27. 数据库的脏读、不可重复读和幻读分别是什么？
28. MySQL 中有哪些锁类型？
29. MySQL 的乐观锁和悲观锁是什么？
30. MySQL 中如果发生死锁应该如何解决？

## 性能优化
31. 如何使用 MySQL 的 EXPLAIN 语句进行查询分析？
32. MySQL 中 count(*)、count(1) 和 count(字段名) 有什么区别？
33. MySQL 中 int(11) 的 11 表示什么？
34. MySQL 中 varchar 和 char 有什么区别？
35. MySQL 中如何进行 SQL 调优？
36. MySQL 中如何解决深度分页的问题？
37. 如何在 MySQL 中监控和优化慢 SQL？
38. MySQL 中 DELETE、DROP 和 TRUNCATE 的区别是什么？
39. MySQL 中 INNER JOIN、LEFT JOIN 和 RIGHT JOIN 的区别是什么？
40. MySQL 中 `LIMIT 100000000, 10` 和 `LIMIT 10` 的执行速度是否相同？

## 数据类型和存储
41. MySQL 中 DATETIME 和 TIMESTAMP 类型的区别是什么？
42. 数据库的三大范式是什么？
43. 在 MySQL 中，你使用过哪些函数？
44. MySQL 中 TEXT 类型最大可以存储多长的文本？
45. MySQL 中 AUTO_INCREMENT 列达到最大值时会发生什么？
46. 在 MySQL 中存储金额数据，应该使用什么数据类型？
47. 什么是数据库的视图？
48. 什么是数据库的游标？
49. 为什么不推荐在 MySQL 中直接存储图片、音频、视频等大容量内容？

## 架构和高可用
50. 相比于 Oracle，MySQL 的优势有哪些？
51. 为什么阿里巴巴的 Java 手册不推荐使用存储过程？
52. 如何实现数据库的不停服迁移？
53. MySQL 数据库的性能优化方法有哪些？
54. MySQL 中 InnoDB 存储引擎与 MyISAM 存储引擎的区别是什么？
55. MySQL 的查询优化器如何选择执行计划？
56. 什么是数据库的逻辑删除？数据库的物理删除和逻辑删除有什么区别？
57. 什么是数据库的逻辑外键？数据库的物理外键和逻辑外键各有什么优缺点？
58. MySQL 事务的二阶段提交是什么？
59. MySQL 三层 B+ 树能存多少数据？
60. MySQL 在设计表（建表）时需要注意什么？
61. MySQL 插入一条 SQL 语句，redo log 记录的是什么？
62. SQL 中 select、from、join、where、group by、having、order by、limit 的执行顺序是什么？