This is a port-forwarding proxy.

So, if you have a server on port 10.10.10.10:4444 and you wanna examine traffic going to it from a bunch of client(s), you can:
* run this proxy at `-local_port=12345 -remote_ip=10.10.10.10 -remote_port=4444`
* tell your client to connect to this host at port 12345
* Now all traffic will be forwarded verbatim to 10.10.10.10:4444 through this proxy

The proxy supports multiple connections and by default will output traffic to stdout.
