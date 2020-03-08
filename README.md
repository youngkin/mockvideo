[![Build Status](https://travis-ci.org/youngkin/mockvideo.svg?branch=master)](https://travis-ci.org/youngkin/mockvideo)

# MockVideo

This project is related to the blog series [Developing & Deploying Kubernetes Applications on a Raspberry Pi Cluster](https://medium.com/@RichYoungkin/develop-and-deploy-kubernetes-applications-on-a-raspberry-pi-cluster-fbd4d97a904c). Unlike the focus of that blog series, this application isn't meant to be deployed exclusively on a Raspberry Pi cluster. Rather, its intent is to showcase microservice development best practices with a relatively simple, but production-ready, application.

MockVideo, as its name implies, provides a mockup of a fictional cable TV company or [MSO (Multiple System Operator)](https://www.techopedia.com/definition/26084/multiple-system-operators-mso). Today most of these companies are evolving beyond simply providing TV service. Several provide not only TV service, but also Internet access and wireless telephony service. The intial focus of this will be on the video delivery aspects of a fictional MSO.

Current development is focused on:

1. Use of [Travis CI](https://travis-ci.org)
2. Creating Docker images
3. Helm deployments to a Kubernetes cluster
4. Implementation of a Customer microservice. Currently the implemention supports the ability to get all customers from a MySQL database. The primary purpose of this initial capability is to demonstrate:

    1.  Application (Go) [package design](https://www.ardanlabs.com/blog/2017/02/package-oriented-design.html)  as promoted by Bill Kennedy at [Arden Labs](https://www.ardanlabs.com).
    2.  Application configuration
    3.  Logging
    4.  Database (MySQL) access
    6.  Unit testing, including HTTP server testing
    7.  Use of Helm/Kubernetes including:

        1. Helm deployments and upgrades
        2. Kubernetes Ingress
        3. Kubernetes Secrets and ConfigMaps
        4. Kubernetes volumes
    5.  Metrics publishing via Prometheus with associated Grafana dashboard(s)
  
    8. Best practices for all of the above.

The project is still in the early phases of development. The initial and current goal is to complete all of the above aspects for the Customer microservice, specifically the ability to get customer information. Achieving this will provide the groundwork necessary to continue to the next phase.

The next phase will focus on the initial development of a second microservice. The purpose of this is to demonstrate the use of common code and to also demonstrate the use of Helm in a second microservice. These two goals will validate the package design and the ability to continue to develop and deploy multiple microservices on independent schedules.

Beyond this, other candidate features include:

1. Full support for customer accounts including what services they've subscribed to.
2. Support for obtaining program guide information
3. Support for recordings, including scheduling recordings (from the program guide) and querying scheduled recordings.

There is no intent to provide access to any video, video recording, or playback of video recording. Implementing these types of features is beyond the scope of this effort and would interfere with the stated purpose of creating a "template" application that demonstrates best-practices of microservice applications deployed to a Kubernetes cluster.

This README will be regularly updated as progress continues. I welcome contributors, PRs, Issues, and comments.

# Docker Respositories

* Customer service - [ryoungkin/customerd](https://hub.docker.com/repository/docker/ryoungkin/customerd)