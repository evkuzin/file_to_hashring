# Hash Ring lab
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

- [ ] distributed transaction to handle file chunks in cases:
  - [ ] during upload
  - [ ]  during rebalance
- [ ] handle faults of any part of the solution with retries/evict bad node etc.
- [ ] implement weights
- [x] implement virtual nodes to rebalance ring in a better way.
```shell
server 1: keys 19.230001%
server 2: keys 19.959999%
server 3: keys 20.990000%
server 4: keys 19.559999%
server 5: keys 20.260000%
# adding one server
Rebalancing took 380.401417ms. 
Total keys before/after: 10000/10000. 
Moved keys: 1659
hashring_test.go:96: server 1: keys 16.370001%
hashring_test.go:96: server 2: keys 16.590000%
hashring_test.go:96: server 3: keys 17.120001%
hashring_test.go:96: server 4: keys 16.300001%
hashring_test.go:96: server 5: keys 17.030001%
hashring_test.go:96: server 6: keys 16.590000%
```
