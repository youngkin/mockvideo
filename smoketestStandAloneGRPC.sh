#!/bin/bash
# Smoke tests for MockVideo app

echo "Starting MySQL"
echo "    docker run -d --name mysql -p 6603:3306 -e MYSQL_ALLOW_EMPTY_PASSWORD=yes mysql:latest"
docker run -d --name mysql -p 6603:3306 -e MYSQL_ALLOW_EMPTY_PASSWORD=yes mysql:latest
echo "Waiting for MySQL to come up..."
sleep 15
echo "SETUP DATABASE:"
echo "    Create tables"
./infrastructure/sql/createTablesDocker.sh
echo "    Pre-populate tables with test data"
./infrastructure/sql/createTestDataDocker.sh

echo ""
echo ""
echo "Start the accountd service"
cd cmd/accountd; go build; ./accountd -configFile "testdata/config/config" -secretsDir "testdata/secrets" -protocol "grpc" &
echo "Wait for the accountd service to start..."
sleep 2

echo ""
echo ""
echo "Run tests"
cd grpc/users/testclient; go build; ./testclient


echo ""
echo ""
sleep 2
echo "Stop MySQL"
docker rm -f mysql
sleep 1
echo "Stop accountd"
ps aux | grep accountd | grep testdata | awk '{printf("kill %s\n", $2) | "sh" }'