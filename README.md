# finviz insider parsing

`go build -o finviz_parser`

## crontab

`crontab -e`

add new row:

`0 9 * * * /root/finviz_parser/finviz_parser`

save and verify:

`crontab -l`

P.S. Check your system's logs (typically /var/log/syslog or /var/log/cron) for any issues related to running the cron jobs.