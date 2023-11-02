# finviz insider parsing

`go build -o finviz_parser`

## crontab

`crontab -e`

add new row (we need `cd` to change working dir to find out .env file):

`0 9 * * * cd /root/finviz_parser && ./finviz_parser > /root/finviz_parser/log 2>&1`

save and verify:

`crontab -l`

P.S. Check your system's logs (typically /var/log/syslog or /var/log/cron) for any issues related to running the cron jobs.