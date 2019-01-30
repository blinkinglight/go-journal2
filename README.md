# journal2 - command line admin journal

MVP


command line notes collector.

```
mkdir ~/gh
cd ~/gh
git clone https://github.com/blinkinglight/go-journal2
cd go-journal2
./build.sh
```

note - journal client

usage:
```
~/bin/note this is long long "text :)"
~/bin/note -q thi*
~/bin/note -n 10
```

jrnl_server - journal server. if SLACK_TOKEN env var is set, starts slack bot.
```
SLACK_TOKEN=xxx-xxxx-xxxx-xxx ./jrnl_server
```

```
http://journal2serveraddr:port
http://journal2serveraddr:port/?query=CaseSensitivePrefixToFilter&limit=10000
```

edit /etc/journal2srv.ini 

```
bind = localhost:18888
db = /path/were/db/lives.db
indexdb = /path/to/index.db

[users]
user:password = r
user1:password1 = w
user2:password2 = rw
```

edit /etc/journal2cli.ini

```
name = Your[nick]Name
url = http://user2:password2@localhost:18888
```

slack commands:
```
@bot last N - shows last N notes
@bot today 
@bot yesterday
@bot date 2019-01-02
```
