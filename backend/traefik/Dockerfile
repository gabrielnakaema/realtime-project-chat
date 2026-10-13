FROM traefik:v3.6.17

COPY dynamic-prod.yml /etc/traefik/dynamic.yml

CMD ["--entrypoints.web.address=:80", "--providers.file.filename=/etc/traefik/dynamic.yml"]
