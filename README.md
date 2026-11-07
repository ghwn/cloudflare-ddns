# cloudflare-ddns

Updates Cloudflare A records to this machine's current public IP.

## Setup

1. Install Go 1.27.2 or later.
2. In the Cloudflare dashboard, create an A record for each domain you want to update. Existing records are updated; missing ones are not created.
3. Create an API token with the **Zone → DNS → Edit** permission for your zone.
4. Copy your zone ID from the zone's **Overview** page.

## Run

Install the binary:

```
go build
install -D cloudflare-ddns ~/.local/bin/cloudflare-ddns
```

Then run `crontab -e` and add this line to run it every 5 minutes:

```
*/5 * * * * CLOUDFLARE_API_TOKEN=<your API token> CLOUDFLARE_ZONE_ID=<your zone ID> DOMAINS=example.com,www.example.com $HOME/.local/bin/cloudflare-ddns
```

List full record names in `DOMAINS`, separated by commas without spaces.
