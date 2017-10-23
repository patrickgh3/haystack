Haystack
---

Efficiently browse Twitch VODs from the last few days. Live at [cwpat.me/haystack](https://cwpat.me/haystack).

![Screenshot](https://i.imgur.com/CAypCiU.png)

Building
---

Built using Go 1.6.2.

The config.yaml file is formatted as follows:
```
path:
    root: / ... /haystack
    images-relative: /images/t
    site-url: https://cwpat.me/haystack
twitch:
    client-key: aaaaaaaaa
    client-secret: bbbbbbbbbb
db:
    user: dbuser
    pass: cccccccccc
    dbname: dbname
timing:
    period-seconds: 300 # time between snapshots
    cutoff-seconds: 900 # cuttoff for new stream
    prune-days: 3 # how many days to store
webserver:
    ip: 127.0.0.1
    port: 1234
```
