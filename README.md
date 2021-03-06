[![Build Status](https://travis-ci.org/youngkin/mockvideo.svg?branch=master)](https://travis-ci.org/youngkin/mockvideo) [![Go Report Card](https://goreportcard.com/badge/github.com/youngkin/mockvideo)](https://goreportcard.com/report/github.com/youngkin/mockvideo)

This project is related to the blog series [Developing & Deploying Kubernetes Applications on a Raspberry Pi Cluster](https://medium.com/better-programming/develop-and-deploy-kubernetes-applications-on-a-raspberry-pi-cluster-fbd4d97a904c). Unlike the focus of that blog series, this application isn't meant to be deployed exclusively on a Raspberry Pi cluster. Rather, its intent is to showcase microservice development best practices with a relatively simple, but production-ready, application. The application is written in Go.

- [Overview](#overview)
- [API](#api)
  - [HTTP](#http)
    - [Resources](#resources)
    - [Common HTTP status codes](#common-http-status-codes)
  - [gRPC](#grpc)
- [Running and testing the application](#running-and-testing-the-application)
  - [Prerequisites](#prerequisites)
  - [Pre-commit check and smoke tests.](#pre-commit-check-and-smoke-tests)
  - [Running the application](#running-the-application)
    - [Local execution](#local-execution)
    - [Run in a Docker container](#run-in-a-docker-container)
    - [Run in Kubernetes](#run-in-kubernetes)
  - [Testing the application](#testing-the-application)
- [Docker Respositories](#docker-respositories)
  
# Overview

MockVideo, as its name implies, provides a mockup of a fictional cable TV company or [MSO (Multiple System Operator)](https://www.techopedia.com/definition/26084/multiple-system-operators-mso). Today most of these companies are evolving beyond simply providing TV service. Several provide not only TV service, but also Internet access and wireless telephony service. The intial focus of this will be on the video delivery aspects of a fictional MSO.

Currently, the following capabilities have been implemented:

1. Use of [Travis CI](https://travis-ci.org)
2. Creating Docker images
3. Helm deployments to a Kubernetes cluster
4. Implementation of an Account microservice. Currently the implemention supports the ability to perform CRUD operations on users associated with any account from a MySQL database. The primary purpose of this initial capability is to demonstrate:

    1.  Application (Go) [package design](https://www.ardanlabs.com/blog/2017/02/package-oriented-design.html)  as promoted by Bill Kennedy at [Arden Labs](https://www.ardanlabs.com), [Dave Cheney's SOLID blog](https://dave.cheney.net/2016/08/20/solid-go-design), and indirectly by ["Uncle Bob's" Clean architecture article](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html).
    2.  Application configuration
    3.  Logging
    4.  Database (MySQL) access
    5.  Implementation of HTTP and gRPC endpoints
    6.  Unit testing, including HTTP and gRPC server testing, and the use [advanced Go testing techniques](https://povilasv.me/go-advanced-testing-tips-tricks/) including the use of 'golden files'.
    7.  Integration testing, including starting, initializing, and stopping required services like MySQL.
    8.  Use of Helm/Kubernetes including:

        1. Helm deployments and upgrades
        2. Kubernetes Ingress
        3. Kubernetes Secrets and ConfigMaps
        4. Kubernetes volumes
    9.  Metrics publishing via Prometheus with associated Grafana dashboard(s)
    10. CI/CD via Jenkins, this includes deploying building a local Docker image. If testing is successful then the docker image will be tagged and pushed to Docker Hub, the application will be deployed to a Kubernetes cluster, and the current git commit will be tagged with the docker image's tag. The application's GitHub repo is linked to [Travis CI](https://travis-ci.org/github/youngkin/mockvideo) where CI including integration testing is performed. There is no automated CD on Travis, the Kubernetes cluster isn't accessible.
    11. Best practices for all of the above.

The primary code for the Account service is located at [cmd/accountd](https://github.com/youngkin/mockvideo/tree/master/cmd/accountd) with supporting code in the [internal package](https://github.com/youngkin/mockvideo/tree/master/internal), and gRPC interfaces and structs in the [pkg package](https://github.com/youngkin/mockvideo/tree/master/pkg). [Helm](https://github.com/youngkin/mockvideo/tree/master/infrastructure/helm), [kubernetes](https://github.com/youngkin/mockvideo/tree/master/infrastructure/kubernetes), [sql](https://github.com/youngkin/mockvideo/tree/master/infrastructure/sql), and [Grafana dashboards](https://github.com/youngkin/mockvideo/tree/master/infrastructure/dashboards) are located in the [infrastructure directory](https://github.com/youngkin/mockvideo/tree/master/infrastructure).

The next phase will focus on the initial development of a second microservice. The purpose of this will be to demonstrate the use of common code and to also demonstrate the use of Helm in a second microservice. These two goals will validate the package design and the ability to continue to develop and deploy multiple microservices on independent schedules. I'll also be working on finding, or developing, an integration test capability in Go.

Beyond this, other candidate features include:

1. Full support for user accounts including what services they've subscribed to.
2. Support for obtaining program guide information
3. Support for recordings, including scheduling recordings (from the program guide) and querying scheduled recordings.

There is no intent to provide access to any video, video recording, or playback of video recording. Implementing these types of features is beyond the scope of this effort and would interfere with the stated purpose of creating a "template" application that demonstrates best-practices of microservice applications deployed to a Kubernetes cluster.

This README will be regularly updated as progress continues. I welcome contributions, PRs, Issues, and comments.

# API

## HTTP

A User is represented in JSON as follows:

```
{
  accountid: {int}      //  The identifier of the user's account
  href: {string}        //  Resource URL, e.g., /users/1. Returned on GET. Don't populate for POST/PUT
  id: {int}             //  Resource identifier, don't populate on POST
  name: {string}        //  The user's name
  email: {string}       //  The user's email address
  role: {int}           //  The user's role relative to the account. They can be the owner (0), admin 
                            capabilities (1), or restricted (2)
  password: {string}    //  This field is provided on POST or PUT. It will never be returned by a GET request.
}
```

Example:

``` JSON
{
  "accountid": 42,
  "href": "/users/42",
  "id": 101,
  "name": "Mickey Dolenz",
  "email": "mdolenz@themonkees.com",
  "role": 0,
  "password": "heyheywerethemonkees"
}
```

The JSON representation for a set of Users is:

``` 
{
  "Users": [
    {
      accountid: {int}
      href: {string}
      id: {int}
      name: {string}
      email: {string}
      role: {int}
      password: {string}
    },
    ...
  ]
}
```

Example:

```
{
  "Users": [
    {
       "accountid": 42,
       "href": "/users/42",
       "id": 101,
       "name": "Mickey Dolenz",
       "email": "mdolenz@themonkees.com",
       "role": 0,
        "password": "heyheywerethemonkees"
    },
    {
       "accountid": 42,
       "href": "/users/52",
       "id": 105,
       "name": "Cass Elliot",
       "email": "cass@mama.com",
       "role": 1,
        "password": "mondaymonday"
    },
    ...
  ]
}
```

For Bulk requests a JSON body is returned that details the results of each sub-request in the bulk operation. With a bulk POST operation the results would be formatted as shown below:

```
{
  "overallstatus":409,
  "results": [
    {
      "httpstatus": 201,
      "errmsg": "",
      "user": {
        "accountid": 1,
        "href": "",
        "id": 6,
        "name": "Brian Wilson",
        "email": "goodvibrations@gmail.com",
        "role": 1,
        "password": "helpmerhonda"
      }
    },
     {
      "httpstatus": 400,
      "errmsg": "attempt to insert duplicate user",
      "user": {
        "accountid": 1,
        "href": "",
        "id": -1,
        "name": "Frank Zappa",
        "email": "donteatyellowsnow@gmail.com",
        "role": 1,
        "password": "searsponcho"
      }    }
  ]
}
```

The `results` above shows the first user was successfully created. The second request failed with an HTTP status of 400. The `errmsg` indicates that the request was an attempt to create a duplicate user. `overallstatus` is a **409** indicating that the entire request did not complete successfully. Said another way, the overall request was at best partially successful.

### Resources

|Verb   | Resource | Description  | Status  | Status Description |
|:------|:---------|:-------------|--------:|:-------------------|
|GET    |/accountdhealth   |Health check, returns `I'm Healthy!` if all's OK  | 200| Service healthy |
|GET    |/users            |Get all users                                     | 200| All users returned |
|GET    |/users/{id}       |Get the user identified by `{id}`                   | 200| user returned |
|       |                  |                                     | 404| user not found|
|POST   |/users     |Create a new user, do not include `id` in JSON body. Returns `Location` header containing self reference|201|user successfully created|
|       |           |If request includes the HTTP header `"Bulk-Request: true"` multiple users will be created in a single request. The `Location` header will not be present. The HTTP response body will contain the results of each sub-request.|201|All users successfully created|
|       |           |                          |409| One or more of the sub-requests failed. Details will be in the body of the response.|
|PUT    |/users/{id}|Update an existing user identified by `{id}`, pass complete JSON in body|200|user updated|
|       |          |                                                                       |404| user not found|
|       |           |If request includes the HTTP header `"Bulk-Request: true"` multiple users will be updated in a single request. The HTTP response body will contain the results of each sub-request.|200|All users successfully created|
|       |           |                          |409| One or more of the sub-requests failed. Details will be in the body of the response.|
|DELETE |/users/{id}|Deletes the referenced resource|200|user was deleted|
|       |          |                                |200|user was not found|

### Common HTTP status codes

|Status|Action|
|-----:|:-----|
|400|Bad request, don't retry|
|429|Server busy, can retry after `Retry-After` time has expired (in seconds)|
|500|Internal server error, can retry, subsequent request _might_ succeed|

## gRPC

gRPC access is also supported. You must import the [github.com/youngkin/mockvideo/pkg/accountd](https://github.com/youngkin/mockvideo/tree/master/pkg/accountd) package to use it. Currently only Golang(Go) clients are supported. The following interface is available:

```go
type UserServerClient interface {
    GetUser(ctx context.Context, in *UserID, opts ...grpc.CallOption) (*User, error)
    GetUsers(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*Users, error)
    CreateUser(ctx context.Context, in *User, opts ...grpc.CallOption) (*UserID, error
    CreateUsers(ctx context.Context, in *Users, opts ...grpc.CallOption) (*BulkResponse, error)
    UpdateUser(ctx context.Context, in *User, opts ...grpc.CallOption) (*empty.Empty, error)
    UpdateUsers(ctx context.Context, in *Users, opts ...grpc.CallOption) (*BulkResponse, error)
    DeleteUser(ctx context.Context, in *UserID, opts ...grpc.CallOption) (*empty.Empty, error)
    Health(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*HealthMsg, error)
}
```

See [pkg](https://github.com/youngkin/mockvideo/tree/master/pkg) for details regarding the API

# Running and testing the application

This section covers how to run the application as a standalone executable, a Docker container, and in a Kubernetes cluster. 

## Prerequisites

A Go development environment must be installed. See the [Go installation page](https://golang.org/doc/install) for details.

A MySQL database, or a feature equivalent such as MariaDB, is required. From here on, the term MySQL will mean either MySQL or its feature equivalent.

MySQL can either be run in a Docker container or from the command line as a standalone service (e.g., installed via Homebrew and started with `brew services start mysql`). Perhaps the easiest way to run MySQL is in a Docker container. This is easily accomplished by running:

```
docker run -d --name mysql -p 6603:3306 -e MYSQL_ALLOW_EMPTY_PASSWORD=yes mysql:latest
```

Things to note about this:

1. `--name mysql` creates a named Docker container. This prevents more than one MySQL instance from running at any given time. If you want to run any tests over with a new copy of the database (e.g., `smoketest.sh`) you'll need to remove the container. To remove both the container and the image at the same time run `docker rm -f mysql`.
2. `-p 6603:3306` creates an external port mapping, 6603, to the internal MySQL port of 3306. 
3. `-e MYSQL_ALLOW_EMPTY_PASSWORD=yes` starts the instance without a `root` password. This is needed by the  `smoketest.sh` script
4. `mysql:latest` will retrieve the mysql image with the tag `latest` which is likely to be the most recent version of the MySQL image.

 The username and password are assumed to be `admin/admin`. These, and other configuration variables, can be changed as described in the sections below on Helm Secrets and `Running the application/ Local execution`. The database can be set up by running `<path-to-project>/mockvideo/infrastructure/sql/createTables.sh` (or `createTablesDocker.sh` if MySQL is running in Docker). You will need to change the host IP address in these files for MySQL to an appropriate IP address (i.e, the `-h10.0.0.100`). `createTestData.sh` or `createTestDataDocker.sh` can be used to populate the the tables with some test data. The IP address specified in these files will likely also have to be changed. As an alternative, setting up the database and populating the tables is done automatically by the `smoketest` scripts described below.

The `accountd` executable will need to be built as well. For local execution:  
```
cd <projectroot>/mockvideo/src/cmd/accountd
go build
```

To build a docker executable run:
```
env GOOS=linux GOARCH=arm GOARM=7 go build
```

This will build an executable that will run in Minikube, Docker Desktop (Mac), and on a Raspberry Pi. You may need to tweak this command, and possibly the Dockerfile, to get an image appropriate for your deployment environment.

To build a Docker image, after building the docker executable above, from the same directory as above run:

```
docker build -t local/accountd .
```

`local` can be anything. I chose `local` to differentiate it from my Docker Hub image. Likewise `accountd` can be anything, but this is the name of the executable so it seems reasonable. I chose not to tag the image relying instead on `latest` since I don't plan on keeping around several versions of this local image.

[Helm](https://helm.sh/) and [Helm Secrets](https://github.com/zendesk/helm-secrets) are required to install to Kubernetes. [Helm secrets – a missing piece in Kubernetes](https://lab.getbase.com/helm-secrets-a-missing-piece-in-kubernetes/) is a good starting point for learning how to use Helm Secrets.

Regarding secrets, if you want to override or even view the MySQL user and password you'll likely need to recreate the `<path-to-project>/mockvideo/src/cmd/accountd/helm/accountd/secrets.values.yaml` file. It's encrypted with my OpenPGP password and is (hopefully!) inaccessible. Prior to encryption with Helm Secrets the file looks like this:

``` YAML
secrets:
    dbuser: admin
    dbpassword: admin
```

## Pre-commit check and smoke tests.

This check is handy when making changes to ensure there is no obvious breakage from any changes.

From the project root directory (`mockvideo`) run:

``` 
./build.sh test
```

This runs `go vet ./...`, `go fmt ./...`, `golint ./...`, and `go test -race ./...` against both the HTTP and gRPC endpoints.

Running `smoketestStandalone.sh` is a good way to see the application in operation. This script will:

1. build the application
2. start MySQL
3. initialize the database
4. start the application
5. run a series of simple tests against the HTTP interface
6. finish by stopping MySQL and the application

Run the script as follows:
```
./smoketestStandalone.sh
```

`smoketestStandaloneGRPC.sh` performs the same operations against gRPC endpoints

## Running the application

Please ensure that all prerequisites have been met as described above.

### Local execution

Run the following commands from  `<path-to-project>/mockvideo/src/cmd/accountd`:

```
go build
./accountd -configFile "testdata/config/config" -secretsDir "testdata/secrets" -protocol ["http" | "grpc"]
```

Per the configuration, the application will listen on port 5000. This, as well as the MySQL location, username, and password can all be configured using configuration and secrets files referred to by the `-configFile` and `-secretsDir` flags in the command line. `smoketest.sh` provides a good example of this command in action. The `-protocol` flag is used to direct the service to start HTTP or gRPC endpoints. They are mutually exclusive. `"http"` is the default if `-protocol` isn't specified.

### Run in a Docker container

See `Prerequisites` above for instructions on how to build the docker container.

[This link provides a good overview of the Docker `run` command](https://rollout.io/blog/the-basics-of-the-docker-run-command/).

`docker run -d   -p 5001:5000 -v <path-to-project>/mockvideo/src/cmd/accountd/testdata:/opt/mockvideo/accountd local/accountd:latest`

`-p 5001:5000` forwards local port 5001 to the application listening port 5000. `-v ...` provides the file system mapping for the container to the local project directory containing the configuration and secrets files.

`docker stop <container-id>` will halt the docker container hosting the application.

### Run in Kubernetes

This requires a Kubernetes environment. Ensure that your context is correctly set. The application also requires an Ingress controller configured as described in [How to Install Kubernetes Ingress on a Raspberry Pi Cluster](https://medium.com/better-programming/install-kubernetes-ingress-on-a-raspberry-pi-cluster-e8d5086c5009). See that article for details. While a Raspberry Pi Kubernetes environment isn't required, the Ingress Controller is.

The first thing to do is to create the namespace to deploy the application into. From `<path-to-project>/mockvideo/infrastructure/kubernetes` enter:

```
kubectl create -f namespace.json
```

The next step is to set up the service and endpoint needed for the application to access MySQL. To start, `cd` to `<path-to-project>/mockvideo/infrastructure/helm`. Before setting up the service and endpoint verify that the configuration is appropriate for your deployment. Check the values specified in the `mockvideo/values.yaml` file, in particular the `ip:` entry at the bottom of the file. It is likely set to `10.0.0.100`. Change this to the IP address of the host where MySQL is running. Then, again from `<path-to-project>/mockvideo/infrastructure/helm` enter:

```
helm install --namespace video --name mysql mockvideo --debug
```

Here are the pertinent parts of the above command:

1. `--namespace video` directs which Kubernetes namespace is the target of the upgrade. 
2. `--name mysql` is the name of the associated Helm release, `mysql` in this case
3. `mockvideo` is the location of the Helm charts
4. `--debug` is a personal preference  of mine as it shows exactly what Helm is doing. This can be useful if something goes wrong.


The next set of commands are all run from `<path-to-project>/mockvideo/src/cmd/accountd`. We're now ready to install the application. Before starting we once again need to verify the configuration matches the deployment environment. The file `helm/values.yaml` should be consulted to ensure the values provided will be acceptable. They should all be fine. If you don't know what you're doing, don't change them. That said, you should look at the `image:`, `respository:` value. The current value, `ryoungkin/accountd`, points to my Docker Hub repository. You'll need to change this if you want to use your own docker image. As built, the docker image will expose the HTTP endpoints. The `cmd/accountd/Dockerfile` will have to be modified to expose the gRPC endpoints, i.e., `ENTRYPOINT ["/bin/accountd -protocol grpc"]`

As discussed in the `Prerequisites` section above, you may need to modify the secrets file. This will only be needed if your MySQL database does not have the user `admin`, with the password `admin` defined. See the section on Secrets above for more details if you need to modify this.

To install the application run:

``` 
helm secrets install --namespace video --name accountd helm/accountd --set image.tag=<desiredImageTag> --values helm/accountd/secrets.values.yaml --debug
```

Here are the pertinent parts of the above command:

1. `--namespace video` directs which Kubernetes namespace is the target of the upgrade. 
2. --name `accountd` is the name of the associated Helm release
3. `helm/accountd` is the location of the Helm charts
4. `--set image.tag=<desiredImageTag>` - The `--set` flag directs Helm to use the value specfied here instead of the matching element in the Helm values file. In this example the values file is located at `<path-to-project>/mockvideo/src/cmd/accountd/helm/values.yaml`. Using `--set` avoids having to update the values file to direct Helm to use the appropriate image. This is helpful since I'm using Jenkins to tag and push the Docker image on a successful build against `master`. If this approach wasn't used, the values file would have to be updated to reflect the new image tag, committed and pushed to master, causing Jenkins to once again build, tag, and push a new Docker image, and so on... 
5. `--values ...` is used to direct Helm Secrets to the location of the secrets file
6. `--debug` as above, this is a personal preference  of mine

To upgrade the application (after the initial installation):

```
helm secrets upgrade --namespace video --set image.tag=<desiredImageTag> --values helm/accountd/secrets.values.yaml accountd helm/accountd --debug
```

For example:

```
helm secrets upgrade --namespace video --set image.tag=0.1.22 --values helm/accountd/secrets.values.yaml accountd helm/accountd --debug
```

## Testing the application

As mentioned above, `.build.sh test`  can be run from the command line at the project root that will run unit tests and things like `go vet`.

There is another shell script at the project root called `smoketest.sh`. This will initialize a running MySQL database and run various `curl` commands to exercise the application. The difference between this script and `smoketestStandAlone.sh` is that running `smoketest.sh` requires a running MySQL database and a running application. It takes 3 parameters, MySQL address, MySQL port, and the service address. I intend to merge `smoketest.sh` and `smoketestStandalone.sh` in the future.

`smoketest.sh` can be run as follows:

```
./smoketest.sh help 
usage:
    smoketest <dbaddr> <dbport> <svcaddr>
Example:
    smoketest localhost 3306 accountd.kube
```

The above example assumes a Kubernetes deployment with an Ingress controller configured as described in [How to Install Kubernetes Ingress on a Raspberry Pi Cluster](https://medium.com/better-programming/install-kubernetes-ingress-on-a-raspberry-pi-cluster-e8d5086c5009). `accountd.kube` is the host name defined by the application's ingress specification in Kubernetes. This assumes a working Kubernetes cluster. It doesn't matter if this cluster is on a Raspberry Pi or not.

Running smoketest against a locally executing application, assuming MySQL is running in a Docker container as described above, could be done by running:

```
./smoketest.sh localhost 6603 localhost:5000
```

To run against an application running in a Docker container, assuming MySQL is running in a Docker container as described above, you could run:

``` sh
./smoketest.sh localhost 6603 localhost:5001
```

This assumes that MySQL is setup as described above and that the application is run from the command line or via `docker` as described above. 

# Docker Respositories

- Account/User services - [ryoungkin/accountd](https://hub.docker.com/repository/docker/ryoungkin/accountd)
