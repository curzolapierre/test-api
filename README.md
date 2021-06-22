# hook-manager
Hook manager is a HTTP server listening for different services like Github, Docker hub, ...

It connects to a redis database on `REDIS_URL`. This redis must be started independently.

# Traefik Credentials

To generate basic auth credentials:

Use `htpasswd` to generate them with:
```bash
echo $(htpasswd -nbB <USER> "<PASS>")
```

To be able to use them directly from docker-compose file (not from environment variables)
we must escape `$` with the command:

```bash
echo $(htpasswd -nbB <USER> "<PASS>") | sed -e s/\\$/\\$\\$/g
```
