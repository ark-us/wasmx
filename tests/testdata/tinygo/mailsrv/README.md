# mail server

```
sudo GOWORK=off go run .
```

## DNS zone (public)

dmail 6 Hours A <IP>
dmail MX 10 dmail.provable.dev.
dmail TXT "v=spf1 ip4:<IP> -all"
_dmarc.dmail TXT "v=DMARC1;p=reject;rua=mailto:dmarc-reports@dmail.provable.dev!10m"


dmail TXT "v=spf1 ip4:82.64.92.77 -all"





; --- A / AAAA -------------------------------------------------
mail           IN  A     203.0.113.10          ; IPv4 of your VM / server
mail           IN  AAAA  2001:db8::10          ; (optional) IPv6

; --- MX (“where to deliver mail for the domain”) --------------
example.com.   IN  MX 10 mail.example.com.     ; priority 10 :contentReference[oaicite:0]{index=0}

; --- SPF (minimal) --------------------------------------------
@              IN  TXT "v=spf1 mx -all"        ; allow only the MX host to send :contentReference[oaicite:1]{index=1}

; --- (Recommended) DMARC --------------------------------------
_dmarc         IN  TXT "v=DMARC1; p=none; rua=mailto:[email protected]"

; --- PTR (record is set at your ISP) --------------------------
10.113.0.203.in-addr.arpa.  IN  PTR  mail.example.com.


##

```sh
sudo adduser --system --group mail

# Install Go ≥ 1.22
sudo apt install golang

# Open port 25 in firewall
sudo ufw allow 25/tcp

# Get TLS cert
sudo certbot certonly --standalone -d dmail.provable.dev

sudo certbot certonly --manual --preferred-challenges=dns
# dmail.provable.dev

# /var/log/letsencrypt/letsencrypt.log
# Successfully received certificate.
# Certificate is saved at: /etc/letsencrypt/live/dmail.provable.dev/fullchain.pem
# Key is saved at:         /etc/letsencrypt/live/dmail.provable.dev/privkey.pem

```

##

build & install

```
go mod tidy
go build -o mailsrv
sudo mv mailsrv /usr/local/bin/
```


* `/etc/systemd/system/mailsrv.service`
```
[Unit]
Description=Minimal Go SMTP receiver
After=network.target

[Service]
ExecStart=/usr/local/bin/mailsrv
User=mail
Restart=on-failure
AmbientCapabilities=CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target

```

```sh
sudo systemctl daemon-reload
sudo systemctl enable --now mailsrv
journalctl -u mailsrv -f   # tail the log, you should see messages arrive

```

## Test

```
dig -x 82.64.92.77
nc -vz 82.64.92.77 25
telnet dmail.provable.dev 25
```


* smoke-test from another host
```
swaks --to you@example.com --server mail.example.com --ehlo test --tls
```
