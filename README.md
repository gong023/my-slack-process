![](https://circleci.com/gh/gong023/my-slack-process.png?circle-token=9b19953963017037b8e38e63fee2239f2c9b43a9&style=shield)

```
PATH="/usr/local/go/bin:$PATH"
WTOKEN=XXXXX
WEATHER_WEBHOOK=XXXXX

* 16 * * * forecast -wtoken $WTOKEN -webhook $WEATHER_WEBHOOK 2>&1 1>/dev/null | stdpost -webhook $WEATHER_WEBHOOK
* 0 * * * forecast -wtoken $WTOKEN -webhook $WEATHER_WEBHOOK 2>&1 1>/dev/null | stdpost -webhook $WEATHER_WEBHOOK
```
