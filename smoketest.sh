#!/bin/bash
# Smoke tests for MockVideo app
#
if [  $1 = "help" ]
then
    echo "usage:"
    echo "    smoketest <dbaddr> <dbport> <svcaddr>"
    echo "Example:"
    echo "    smoketest localhost 3306 accountd.kube"
    exit
fi

DBADDR=$1
DBPORT=$2
VIDEOADDR=$3
DBPASSWORD="admin"

###
### Setup database
###
echo "SETUP DATABASE:"
echo "    Create tables"
mysql -h$DBADDR -uadmin -p$DBPASSWORD -P $DBPORT < ./infrastructure/sql/create.sql
echo "    Pre-populate tables with test data"
mysql -h$DBADDR -uadmin -p$DBPASSWORD -P $DBPORT < ./infrastructure/sql/testdata.sql

###
### Success Tests
###
echo ""
echo "GET all users"
echo "    RUN curl http://"$VIDEOADDR"/users"
echo "    EXPECT 5 users"
curl http://$VIDEOADDR/users | jq "."

echo ""
echo "POST(Insert) new user Brian Wilson"
echo "    RUN curl -i -X POST http://"$VIDEOADDR"/users/ -H "Content-Type: application/json" -d "{\"accountid\":1,\"name\":\"Brian Wilson\",\"email\":\"goodvibrations@gmail.com\",\"role\":1,\"password\":\"helpmerhonda\"}""
echo "    EXPECT an HTTP Status 201 Created"
curl -i -X POST http://"$VIDEOADDR"/users/ -H "Content-Type: application/json" -d "{\"accountid\":1,\"name\":\"Brian Wilson\",\"email\":\"goodvibrations@gmail.com\",\"role\":1,\"password\":\"helpmerhonda\"}"

echo ""
echo "Get all users"
echo "    RUN curl http://"$VIDEOADDR"/users"
echo "    EXPECT to see new user Brian Wilson with ID: 6"
curl http://$VIDEOADDR/users | jq "."

echo ""
echo "PUT(Update) user Brian Wilson"
echo "    RUN curl -i -X PUT http://"$VIDEOADDR"/users/6 -H "Content-Type: application/json" -d "{\"accountid\":1,\"id\":6,\"name\":\"BeachBoy Brian Wilson\",\"email\":\"goodvibrations@gmail.com\",\"role\":1,\"password\":\"helpmerhonda\"}""
echo "    EXPECT an HTTP Status 200 OK"
curl -i -X PUT http://"$VIDEOADDR"/users/6 -H "Content-Type: application/json" -d "{\"accountid\":1,\"id\":6,\"name\":\"BeachBoy Brian Wilson\",\"email\":\"goodvibrations@gmail.com\",\"role\":1,\"password\":\"helpmerhonda\"}"

echo ""
echo "Get all users"
echo "    RUN curl http://"$VIDEOADDR"/users"
echo "    EXPECT to see user Brian Wilson with new name BeachBoy Brian Wilson"
curl http://$VIDEOADDR/users | jq "."

echo ""
echo "DELETE just added user Brian Wilson"
echo "    RUN curl -X DELETE http://"$VIDEOADDR"/users/6"
echo "    EXPECT an HTTP Status 200 OK"
curl -X DELETE http://$VIDEOADDR/users/6

echo ""
echo "Get all users"
echo "    RUN curl http://"$VIDEOADDR"/users"
echo "    EXPECT to see 5 users without DELETE-d user Brian Wilson"
curl http://$VIDEOADDR/users | jq "."


###
### Failure Tests
###
echo ""
echo "GET bad request"
echo "    RUN curl -i http://"$VIDEOADDR"/xxx"
echo "    EXPECT an HTTP Status 400 Bad Request, /xxx is an unexpected request"
curl -i http://$VIDEOADDR/xxx 

echo ""
echo "POST(Insert) new user Brian Wilson with a duplicate email address"
echo "    RUN curl -i -X POST http://"$VIDEOADDR"/users/ -H "Content-Type: application/json" -d "{\"accountid\":1,\"name\":\"Brian Wilson\",\"email\":\"mickeyd@gmail.com\",\"role\":1,\"password\":\"helpmerhonda\"}""
echo "    EXPECT an HTTP Status 400 Bad Request since email addresses must be unique"
curl -i -X POST http://"$VIDEOADDR"/users/ -H "Content-Type: application/json" -d "{\"accountid\":1,\"name\":\"Brian Wilson\",\"email\":\"mickeyd@gmail.com\",\"role\":1,\"password\":\"helpmerhonda\"}"

echo ""
echo "PUT(Update) non-existing user Brian Wilson"
echo "    RUN curl -i -X PUT http://"$VIDEOADDR"/users/6 -H "Content-Type: application/json" -d "{\"accountid\":1,\"id\":6,\"name\":\"BeachBoy Brian Wilson\",\"email\":\"goodvibrations@gmail.com\",\"role\":1,\"password\":\"helpmerhonda\"}""
echo "    EXPECT an HTTP Status 400 Bad Request since you can't update a non-existent user"
curl -i -X PUT http://"$VIDEOADDR"/users/6 -H "Content-Type: application/json" -d "{\"accountid\":1,\"id\":6,\"name\":\"BeachBoy Brian Wilson\",\"email\":\"goodvibrations@gmail.com\",\"role\":1,\"password\":\"helpmerhonda\"}"

echo ""
echo "DELETE a non-existing user"
echo "    RUN curl -X DELETE http://"$VIDEOADDR"/users/6"
echo "    EXPECT an HTTP Status 200 OK (MySQL silently accepts deletes of non-existing rows)"
curl -i -X DELETE http://$VIDEOADDR/users/6

