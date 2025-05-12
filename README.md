# sandbox-http-server-go

## Environment variables

<https://github.com/pollenjp/sandbox-http-server-go/blob/9f89cdfaa569f1a89c62d1cb2418c16704cb9b09/docker-compose.yml#L24-L26>

| Required         | required | example                                                             |
| :--------------- | :------- | :------------------------------------------------------------------ |
| `DATABASE_URL`   | no       | `postgres://testuser:password@postgres:5432/testdb?sslmode=disable` |
| `SERVER_ADDRESS` | no       | `127.0.0.1`                                                         |
| `SERVER_PORT`    | no       | `8080`                                                              |

## Accessible paths

| path  | description                                  |
| :---- | :------------------------------------------- |
| `/`   | return simple text.                          |
| `/db` | You can access when you set `DATABASE_URL` . |

## Development

### Run

```sh
go run .
```

docker-compose

```sh
make docker-rerun
```

Open <http://localhost:80/>
