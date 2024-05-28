#!/bin/bash
# 1. Connect to the master MYSQL and create the replication user
cat ./setup-replication-user.sql | docker exec -i mysql-primary mysql -uroot -prootpassword


# 2. Connect to the master MySQL and get the current binary log file and position
MASTER_STATUS=$(docker exec mysql-primary mysql -uroot -prootpassword -e "SHOW MASTER STATUS;")
CURRENT_LOG_FILE=$(echo $MASTER_STATUS | awk '{print $6}')  # Adjust the index based on output format
CURRENT_LOG_POS=$(echo $MASTER_STATUS | awk '{print $7}')   # Adjust the index based on output format

echo $CURRENT_LOG_FILE
echo $CURRENT_LOG_POS
# 3. Now configure the slave with the fetched log file and position
docker exec mysql-replica mysql -uroot -prootpassword -e \
"CHANGE MASTER TO
  MASTER_HOST='mysql-primary',
  MASTER_USER='replicator',
  MASTER_PASSWORD='replicapass',
  MASTER_LOG_FILE='$CURRENT_LOG_FILE',
  MASTER_LOG_POS=$CURRENT_LOG_POS;
START SLAVE;"

# 4. Create database and table on master
cat ./init.sql | docker exec -i mysql-primary mysql -uroot -prootpassword
