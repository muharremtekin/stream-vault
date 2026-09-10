# The local single-server stack can remain stopped for long periods. Keep its
# persisted state usable without deleting the Consul volume after the default
# seven-day stale-server safety window has elapsed.
server_rejoin_age_max = "87600h"
