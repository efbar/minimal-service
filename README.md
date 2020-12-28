Minimal Service
=======

## Aim of the project

The purpose is to create a simple HTTP service to be used in microservices environments.
This service can be deployed as raw binary or as a container and could be useful for testing service-to-service communication with focus on service discovery and service mesh scenarios.

## How it works

It accepts only two methods, `GET` and `POST`.

At every `GET` request, it will respond with the same sent body or none if empty or not present in request. It will add some useful headers like timing metrics related headers and who served the request.
Request `Content-Type` could be `application/json` and `text/plain`, it will respond accordingly.

At every `POST`, it will check that body must include a json string:

```json
{"rebound":"true","endpoint":"http://my.other.service"}
```

it will check the keys strictly, and then will create a `GET` request to the endpoint filled in `endpoint` key.

Every other body with any differences (unless the `endpoint` value) will be discarded and warned.


## Installation

Install binary:

```console
foo@bar:~/minimal-service$ go install
```

Build container, basic:

```console
foo@bar:~/minimal-service$ docker build . -t <tag>
```

with environment variables:

```console
foo@bar:~/minimal-service$ docker build --build-arg alt_port=8080 . -t <tag>
```

## Run

In a non-containerized way:

```console
foo@bar:~/minimal-service$ go run main.go
```

or, after installation

```console
foo@bar:~/minimal-service$ minimal-service
```

You can change the service port exporting like this:

```bash
export SERVICE_PORT=8080
```

With Docker, something like:

```console
foo@bar:~/minimal-service$ docker run --rm -p8080:8080 -e SERVICE_PORT=8080 --name simple <tag>
```

The default service port is `9090`.

### Health Status API

You can perform an health-check with a simple `GET` request at path `/health`.
If up, it will respond with a `200` status and with `Status OK` string in body.

