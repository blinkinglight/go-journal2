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

jrnl_server - journal server

```
http://journal2serveraddr:port
http://journal2serveraddr:port/?query=CaseSensitivePrefixToFilter&limit=10000
```

edit /etc/journal2srv.ini 

```
bind = ip:port
db = /path/were/db/lives.db
indexdb = /path/to/index.db

[users]
user = password 
user1 = password1
user2 = password2
```

edit /etc/journal2cli.ini

```
name = Your[nick]Name
url = http://youusername:yourpassword@journal2serveraddress:port
```
