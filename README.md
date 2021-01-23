Minimal Service
=======

[![CircleCI](https://circleci.com/gh/efbar/minimal-service/tree/main.svg?style=shield)](https://circleci.com/gh/efbar/minimal-service/tree/main)

##### Docker Images

[https://hub.docker.com/r/efbar/minimal-service](https://hub.docker.com/r/efbar/minimal-service)

## Aim of the project

The purpose is to create a simple HTTP service to be used in microservices environments.
This service can be deployed as raw binary or as a container and could be useful for testing service-to-service communication with focus on service discovery and service mesh scenarios.

## How it works

It accepts only two methods, `GET` and `POST`.

At every `GET` request, it will respond with the same sent body or none if empty or not present in request. It will add some useful headers like timing metrics related headers and who served the request.
Request `Content-Type` could be `application/json` and `text/plain`, it will respond accordingly.

A `POST` request at path `/bounce` is accepted. Then it will check that body must include a json string:

```json
{"rebound":"true","endpoint":"http://my.other.service"}
```

it will strictly check json keys and then it will do a `GET` request to the endpoint filled in `endpoint` key.

Every other body with any differences (unless the `endpoint` value) will be discarded and warned.

If the endpoint value has a valid and alive url and the next `POST` request ends positively, the response will have as Headers the ones received from the endpoint response, plus some others like `Response-time`, `Request-time` and `Duration`. 
The body will contain the endpoint's response HTTP status (`OK 200`).

In case of not `200` codes from the endpoint, the headers will be not modified and the body will contain `500 Internal Server Error`.

### Features

The service has some features and you can set them with environment variables.

- Choose the port where the service is listening to
- Set a maximum delay time, so at every request received it will wait for a number of seconds in the range from `0` to the chosen value.
- It will enable tracing to a Jaeger endpoint through the use of OpenTelemetry dependency. It traces status codes and headers.
- Discard request entirely without any feedback or reject them with 500 http status response.
- Consul Connect integration, setting some env variables, you can register this service to a Consul catalog.

This service can be started setting various environment variables, here the list:

| Env variable | default |value range|
| ------------------- |:-----:|----|
|`SERVICE_PORT` | `9090`||
|`DELAY_MAX`|`0`|
|`TRACING`|`0`| `0` or `1`| 
|`JAEGER_URL`| `http://localhost:14268/api/traces`|`URI in form scheme://host:port/api/traces`|
|`DISCARD_QUOTA`|`0`|from `0` to `100`|
|`REJECT`|`0`| `0` or `1`| 
|`DEBUG`|`0`| `0` or `1`| 
|`CONNECT`|`0`| `0` or `1`| 
|`CONSUL_SERVER`|`http://127.0.0.1:8500`| `URI in form scheme://host:port` | 

### Health Status API

You can perform an health-check with a simple `GET` request at path `/health`.
If up, it will respond with a `200` status and with `Status OK` string in body.

## Build, Test, Installation, Run and so on...

You can use Makefile:

```bash
Choose a command from list below
 usage: make <command>

  build
  run
  install
  test
  clean
  build-linux
  build-mac
  build-windows
  build-docker
```


