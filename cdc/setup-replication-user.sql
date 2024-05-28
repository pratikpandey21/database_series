CREATE USER 'replicator'@'%' IDENTIFIED WITH mysql_native_password BY 'replicapass';
GRANT REPLICATION SLAVE ON *.* TO 'replicator'@'%';
FLUSH PRIVILEGES;
