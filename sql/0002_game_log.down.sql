# DELIMITER $$
# CREATE PROCEDURE myFunction() # 创建mysql存储过程
#   BEGIN
#     DECLARE i INT DEFAULT 1;
#
#     WHILE (i <= 128) DO
#       INSERT INTO ascii_chart VALUES (i, CHAR(i));
#       SET i = i + 1;
#     END WHILE;
#
#     SELECT *
#     FROM ascii_chart;
#
#     DROP TABLE ascii_chart;
#
#   END$$