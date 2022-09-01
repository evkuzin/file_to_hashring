# Hashring lab
## HOWTO

prerequisites:
- docker
- docker-compose
- golang

```shell
make test
make docker-compose-up
make upload
make rebalance
make download
make check
```

## TODO

- implement virtual nodes to rebalance ring in a better way. ATM keys distribution is nice, but +1 node is a mess:
```shell
server 1: keys 19.230001%
server 2: keys 19.959999%
server 3: keys 20.990000%
server 4: keys 19.559999%
server 5: keys 20.260000%
# adding one server
Total keys: 10000. Moved keys: 8347
server 1: keys 16.859999%
server 2: keys 16.820000%
server 3: keys 16.520000%
server 4: keys 16.900000%
server 5: keys 16.290001%
server 6: keys 16.609999%
```
