# Start server

```sh
$ go mod download
```

```sh
$ go run main.go

   ____    __
  / __/___/ /  ___
 / _// __/ _ \/ _ \
/___/\__/_//_/\___/ v4.6.3
High performance, minimalist Go web framework
https://echo.labstack.com
____________________________________O/_______
                                    O\
⇨ http server started on [::]:8080
```

websockets should be available at `localhost:8080/sequence/:sequence/delay/:delay`
where `sequence` can be `success`, `fail` or `random` and `delay` is the duration of time to wait between messages in the form of `200ms`, `10s` or `1m`
