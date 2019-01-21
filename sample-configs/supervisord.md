```
[program:journal2srv]
command=/opt/jrnl_server --config=/etc/journal2srv.ini
autostart = true
startsec = 1
user = root
redirect_stderr = true
stdout_logfile_maxbytes = 200MB
stdout_logfile_backups = 10
stdout_logfile = /var/log/supervisor/journal2srv.log
autorestart=true
startretries=1000
stopasgroup=true
killasgroup=true
stdout_events_enabled=true
stderr_events_enabled=true
exitcodes=0,2
```
