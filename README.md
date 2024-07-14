# discorss

A tool to send updated feed items to Discord server via webhook.

## how to use

1. Write a config file (should be named as `config.yml`).
2. Prepare cron job.
3. Build the binary and set cron.

### config

```
# config.yml
hook: https://discord.com/api/webhooks/XXXXXX.../
frequency: 3 # should be one of the divisors of 24
feeds:
  - http://example.com
  - ...
```

- hook: Discord Webhook URL.
- frequency: How many numbers you want to receive the update in a day. For example, if this is set to `3`, you'll receive the update every 8 hours.
- feeds: Feed URLs.

### cron job

This must match `frequency` in `config.yml`.  
e.g. `0 4,12,20 * * * cd /path/to/binary && ./discorss`

### build

```
go build
```

Place the binary and `config.yml` in the same directory, and add the cron job to finish the setup.
