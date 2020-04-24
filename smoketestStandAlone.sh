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

echo "Start the accountd service"
cd src/cmd/accountd; go build; ./accountd -configFile "testdata/config/config" -secretsDir "testdata/secrets" &
echo "Wait for the accountd service to start..."
sleep 2

###
### Success Tests
###
echo ""
echo "GET all users"
echo "    RUN curl http://"localhost:5000"/users"
echo "    EXPECT 5 users"
curl http://localhost:5000/users | jq "."

echo ""
echo "POST(Insert) new user Brian Wilson"
echo "    RUN curl -i -X POST http://"localhost:5000"/users/ -H "Content-Type: application/json" -d "{\"accountid\":1,\"name\":\"Brian Wilson\",\"email\":\"goodvibrations@gmail.com\",\"role\":1,\"password\":\"helpmerhonda\"}""
echo "    EXPECT an HTTP Status 201 Created"
curl -i -X POST http://"localhost:5000"/users/ -H "Content-Type: application/json" -d "{\"accountid\":1,\"name\":\"Brian Wilson\",\"email\":\"goodvibrations@gmail.com\",\"role\":1,\"password\":\"helpmerhonda\"}"

echo ""
echo "Get all users"
echo "    RUN curl http://"localhost:5000"/users"
echo "    EXPECT to see new user Brian Wilson with ID: 6"
curl http://localhost:5000/users | jq "."

echo ""
echo "PUT(Update) user Brian Wilson"
echo "    RUN curl -i -X PUT http://"localhost:5000"/users/6 -H "Content-Type: application/json" -d "{\"accountid\":1,\"id\":6,\"name\":\"BeachBoy Brian Wilson\",\"email\":\"goodvibrations@gmail.com\",\"role\":1,\"password\":\"helpmerhonda\"}""
echo "    EXPECT an HTTP Status 200 OK"
curl -i -X PUT http://"localhost:5000"/users/6 -H "Content-Type: application/json" -d "{\"accountid\":1,\"id\":6,\"name\":\"BeachBoy Brian Wilson\",\"email\":\"goodvibrations@gmail.com\",\"role\":1,\"password\":\"helpmerhonda\"}"

echo ""
echo "Get all users"
echo "    RUN curl http://"localhost:5000"/users"
echo "    EXPECT to see user Brian Wilson with new name BeachBoy Brian Wilson"
curl http://localhost:5000/users | jq "."

echo ""
echo "DELETE just added user Brian Wilson"
echo "    RUN curl -X DELETE http://"localhost:5000"/users/6"
echo "    EXPECT an HTTP Status 200 OK"
curl -i -X DELETE http://localhost:5000/users/6

echo ""
echo "Get all users"
echo "    RUN curl http://"localhost:5000"/users"
echo "    EXPECT to see 5 users without DELETE-d user Brian Wilson"
curl http://localhost:5000/users | jq "."

echo ""
echo "Bulk POST multiple users"
echo "    RUN curl -i -X POST http://"localhost:5000"/users/ -H "Bulk-Request: true" -H "Content-Type: application/json" -d "{\"users\":[{\"accountid\":1,\"name\":\"Brian Wilson\",\"email\":\"goodvibrations@gmail.com\",\"role\":1,\"password\":\"helpmerhonda\"},{\"accountid\":1,\"name\":\"Frank Zappa\",\"email\":\"donteatyellowsnow@gmail.com\",\"role\":1,\"password\":\"searsponcho\"}]}""
echo "    EXPECT an overall HTTP Status of 201 Created, and a 'results' JSON body showing the results of each individual request"
curl -i -X POST http://localhost:5000/users/ -H "Bulk-Request: true" -H "Content-Type: application/json" -d "{\"users\":[{\"accountid\":1,\"name\":\"Brian Wilson\",\"email\":\"goodvibrations@gmail.com\",\"role\":1,\"password\":\"helpmerhonda\"},{\"accountid\":1,\"name\":\"Frank Zappa\",\"email\":\"donteatyellowsnow@gmail.com\",\"role\":1,\"password\":\"searsponcho\"}]}"

echo ""
echo ""
echo "Get all users"
echo "    RUN curl http://"localhost:5000"/users"
echo "    EXPECT to see 7 users with Brian Wilson and Frank Zappa"
curl http://localhost:5000/users | jq "."


echo ""
echo "Bulk POST multiple users, adding only Neil Finn, but as a bulk operation."
echo "    RUN curl -i -X POST http://"localhost:5000"/users/ -H "Bulk-Request: true" -H "Content-Type: application/json" -d "{\"users\":[{\"accountid\":1,\"name\":\"Split Enz Neil Finn\",\"email\":\"splitenz@gmail.com\",\"role\":1,\"password\":\"sinner\"}]}""
echo "    EXPECT an overall HTTP Status of 201 Created, and a user ID of 8!!!"
echo "    EXPECT a single user is inserted via bulk POST so we can have a deterministic user ID of 8."
curl -i -X POST http://localhost:5000/users/ -H "Bulk-Request: true" -H "Content-Type: application/json" -d "{\"users\":[{\"accountid\":1,\"name\":\"Split Enz Neil Finn\",\"email\":\"splitenz@gmail.com\",\"role\":1,\"password\":\"sinner\"}]}"

echo ""
echo "Get all users"
echo "    RUN curl http://"localhost:5000"/users"
echo "    EXPECT to see 8 users with Split Enz Neil Finn"
curl http://localhost:5000/users | jq "."

echo ""
echo "Bulk PUT multiple users (really only 1, but 'Bulk-Request is true"
echo "    RUN curl -i -X PUT http://"localhost:5000"/users/ -H "Bulk-Request: true" -H "Content-Type: application/json" -d "{\"users\":[{\"accountid\":1,\"name\":\"Fleetwood Mac Neil Finn\",\"email\":\"splitenz@gmail.com\",\"role\":1,\"password\":\"sinner\"}]}""
echo "    EXPECT an overall HTTP Status of 200 OK, and a 'results' JSON body showing the Fleetwood Mac Neil Finn"
curl -i -X PUT http://localhost:5000/users/ -H "Bulk-Request: true" -H "Content-Type: application/json" -d "{\"users\":[{\"accountid\":1,\"id\":9,\"name\":\"Fleetwood Mac Neil Finn\",\"email\":\"splitenz@gmail.com\",\"role\":1,\"password\":\"helpmerhonda\"}]}"
echo ""

echo ""
echo "Get all users"
echo "    RUN curl http://"localhost:5000"/users"
echo "    EXPECT to see 8 users with Fleetwood Mac Neil Finn"
curl http://localhost:5000/users | jq "."


###
### Failure Tests
###
echo ""
echo "GET bad request"
echo "    RUN curl -i http://"localhost:5000"/xxx"
echo "    EXPECT an HTTP Status 400 Bad Request or 404 Not Found if running in k8s, /xxx is an unexpected request"
curl -i http://localhost:5000/xxx 
echo ""
echo ""

echo ""
echo "POST(Insert) new user Brian Wilson with a duplicate email address"
echo "    RUN curl -i -X POST http://"localhost:5000"/users/ -H "Content-Type: application/json" -d "{\"accountid\":1,\"name\":\"Brian Wilson\",\"email\":\"mickeyd@gmail.com\",\"role\":1,\"password\":\"helpmerhonda\"}""
echo "    EXPECT an HTTP Status 400 Bad Request since email addresses must be unique"
curl -i -X POST http://"localhost:5000"/users/ -H "Content-Type: application/json" -d "{\"accountid\":1,\"name\":\"Brian Wilson\",\"email\":\"mickeyd@gmail.com\",\"role\":1,\"password\":\"helpmerhonda\"}"
echo ""
echo ""

echo ""
echo "PUT(Update) non-existing user Brian Wilson"
echo "    RUN curl -i -X PUT http://"localhost:5000"/users/6 -H "Content-Type: application/json" -d "{\"accountid\":1,\"id\":6,\"name\":\"BeachBoy Brian Wilson\",\"email\":\"goodvibrations@gmail.com\",\"role\":1,\"password\":\"helpmerhonda\"}""
echo "    EXPECT an HTTP Status 400 Bad Request since you can't update a non-existent user"
curl -i -X PUT http://"localhost:5000"/users/6 -H "Content-Type: application/json" -d "{\"accountid\":1,\"id\":6,\"name\":\"BeachBoy Brian Wilson\",\"email\":\"goodvibrations@gmail.com\",\"role\":1,\"password\":\"helpmerhonda\"}"
echo ""
echo ""

echo ""
echo "DELETE a non-existing user"
echo "    RUN curl -X DELETE http://"localhost:5000"/users/6"
echo "    EXPECT an HTTP Status 200 OK (MySQL silently accepts deletes of non-existing rows)"
curl -i -X DELETE http://localhost:5000/users/6

sleep 2
echo "Stop MySQL"
docker rm -f mysql
sleep 1
echo "Stop accountd"
ps aux | grep accountd | grep testdata | awk '{printf("kill %s\n", $2) | "sh" }'