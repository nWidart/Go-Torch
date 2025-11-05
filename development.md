# GoTorch — Torchlight Infinite Tracker

## Dev tools

### append.sh
```shell
 ./append.sh -file mylog.log MapStart
 ./append.sh -file mylog.log BagInit --itemId 370201
 ./append.sh -file mylog.log ItemDrop 370201
 ./append.sh -file mylog.log MapEnd
```

### Update prices

```shell
go run ./cmd/updateprices
go run ./cmd/updateprices --file full_table.json --dry-run
```

## License

MIT (see your repository choice).
