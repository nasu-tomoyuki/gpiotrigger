gpiotrigger
==========

# これはなに？
Raspberry Pi の GPIO を監視して、指定コマンドを実行します。
想定している動作はこんな感じです。

## 例:RPi にシャットダウンボタンを付ける
よくある使用例です。
RPi には電源スイッチがないので、シャットダウンに手間がかかります。本体のボタンでシャットダウン出来たら直感的で楽ですよね。

[PC 用のリセットスイッチ](http://www.ainex.jp/products/pa-045.htm)がちょうどいいので、これを GPIO 27 と GND に繋いでシャットダウンボタンとします。

接続後、次のように実行します。

```
sudo gpiotrigger -pin=27 -time=5 -command="shutdown -h now" &
```

これで、リセットスイッチを５秒間押し続けると shutdown コマンドが走ります。

## 例:シャットダウンボタンの監視を自動起動する
### gpiotrigger をコピー
```
$cp gpiotrigger /usr/local/bin/
```

### 起動させたいスクリプトを作成
```
$cat /usr/local/bin/gpio_shutdown.sh
#!/bin/bash
aplay /usr/local/bin/gpio_shutdown.wav
shutdown -h now
```

### 自動起動するように設定

/etc/rc.local に以下を追加

```
su -c "/usr/local/bin/gpiotrigger -time=5 -pin=27 -command=/usr/local/bin/gpio_shutdown.sh &"
```

以上で、起動するたびに GPIO 27 を監視するスクリプトが自動で起動され、５秒間スイッチを押すと wav を再生してからシャットダウンするようになります。


# インストール方法
```
$go get github.com/nasu-tomoyuki/gpiotrigger
```

これで github からソースを取得し、$GOPATH/bin にビルド結果をコピーします。

# 実行方法

GPIO の使用を始めるのに管理者権限が必要なので、sudo を付けて実行してください。

```
$sudo gpiotrigger
```

起動オプションについては -h で確認出来ます。

目的の GPIO がすでに使われている場合は起動出来ません。CTRL-C や kill で強制終了した場合は GPIO が使用しっぱなしのため、次回起動に失敗します。その場合はまず手動で GPIO を閉じてください。

```
$sudo echo 27 >/sys/class/gpio/unexport
```

# 技術的な説明

ターゲットの GPIO を仮想ファイルで開きます。そのファイルデスクリプタを epoll で監視します。メインスレッドはスリープして、監視スレッドは GPIO に変化があるまで epoll で停止します。

# 参考

* [Raspberry Piのメモ](http://www.siio.jp/index.php?How2RaspberryPi)
* [Raspberry Pi にシャットダウンボタンをつける](http://d.hatena.ne.jp/penkoba/20130925/1380129824)
* [RASPBERRY PIとGO言語でLチカさせてみました！](http://panda.holy.jp/2014/01/135/)
* [Basic epoll usage using Go](https://gist.github.com/gcmurphy/4174057)

... 他多数



