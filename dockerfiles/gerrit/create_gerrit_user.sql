CREATE USER 'gerrit2'@'%' IDENTIFIED BY 'secret';
CREATE DATABASE reviewdb;
ALTER DATABASE reviewdb charset=latin1;
GRANT ALL PRIVILEGES ON *.* TO 'gerrit2'@'%' IDENTIFIED BY 'secret';
FLUSH PRIVILEGES;
