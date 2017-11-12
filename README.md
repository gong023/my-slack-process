![https://circleci.com/gh/gong023/my-slack-process](https://circleci.com/gh/gong023/my-slack-process.png?circle-token=9b19953963017037b8e38e63fee2239f2c9b43a9&style=shield)

```
PATH="/usr/local/go/bin:$PATH"
WTOKEN=XXXXX
WEATHER_WEBHOOK=XXXXX

00  16 * * * forecast -wtoken $WTOKEN 2>&1 | stdpost -webhook $WEATHER_WEBHOOK
00   0 * * * forecast -wtoken $WTOKEN 2>&1 | stdpost -webhook $WEATHER_WEBHOOK
*/20 * * * * stdpostd -wtoken $WTOKEN -messages <(inoreader -refresh_path XXX -client_id XXX -client_sec XXX -tags XXX)
```
