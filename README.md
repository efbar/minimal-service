Minimal Service
=======

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

In case of not `200` codes from the endpoint, the headers will be not modified and the body will cointain `500 Internal Server Error`.

### Features

This service can be started setting various environment variables, here a list:

* `SERVICE_PORT` is the port service is listen from, default is `9090`.
* `DELAY_MAX`, in seconds, just make service responding time to be delayed with a value between 0 and the `DELAY_MAX`, randomly at every request.
* `TRACING`, could be `1` or `0`, it will enable tracing to a Jaeger endpoint through the use of OpenTelemetry dependency.
* `JAEGER_URL` will set custom Jaeger URL. Default is `http://localhost:14268/api/traces`.
* `DISCARD_QUOTA` is a value from 0 to 100. Represents the percentage of discarded requests, it will only return 200 status without any body returned.
* `REJECT` could be `1` and `0`, it will reject the request returning a 500 status.
* `DEBUG` could be `1` and `0`, it will add DEBUG logging.

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


