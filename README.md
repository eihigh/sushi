# Sushi - The next generation form of shell
絶賛開発中のため実用はしないでください。Here be dragons.

## Screenshot
[Screenshot](screenshot.png)

## Installation
Go1.16 or later required.

```
go install github.com/eihigh/sushi/cmd/sushi@master
```

## Concepts
上から下へログを垂れ流すだけのシェルとは一線を画した「次世代の形状のシェル」です。

- コマンドが終了するのを待たずに次のコマンドを実行できます
- 各コマンドの出力は別々に保存されます
- Key Up / Down で表示するコマンドの出力を切り替えます
