# GoTorch — Torchlight Infinite Tracker

GoTorch is a desktop app (Wails v2 + React) and CLI that monitors Torchlight Infinite's `UE_game.log`, detects map runs, and tallies item drops during each run.

Download the latest release via the Actions tab. (_will be improved via released on stabilized_)

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
