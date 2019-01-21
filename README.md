this is MVP

command line notes collector.

```
./build.sh
```

note - journal client

jrnl_server - journal server

```
http://journal2serveraddr:port
http://journal2serveraddr:port/?query=CaseSensitivePrefixToFilter&limit=10000
```

edit /etc/journal2srv.ini 

```
bind = ip:port // :port
db = /path/were/db/lives.db

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
