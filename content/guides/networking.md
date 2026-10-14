# Networking from the ground up

**In one sentence:** networking is programs on different machines exchanging
bytes through a stack of layers, where each layer solves one problem:
finding the machine, reaching the right program, delivering reliably, and
agreeing what the bytes mean.

**Why it matters:** a large share of incidents start as "service A can't talk
to service B". If you know the layers, you can ask the right questions in
the right order (is it DNS? the route? the port? TLS? the app?) instead of
restarting things and hoping.

**You'll learn:**

- The four layers, and what each one adds.
- IP addresses, IPv4 vs IPv6, CIDR blocks and private ranges.
- Ports, sockets, and how TCP differs from UDP.
- How DNS turns names into addresses, and why its caches bite.
- What an HTTP request and response look like, and what TLS adds.
- Why every network call needs a timeout, and how to debug "can't connect".

## One request, start to finish

You type this on your laptop:

```bash
curl https://api.example.com/healthz
```

In a fraction of a second, four separate things happen:

1. **Find the address.** DNS turns `api.example.com` into an IP address such
   as `203.0.113.10`.
2. **Connect.** TCP opens a connection to port 443 on that address.
3. **Secure it.** TLS checks the server's certificate and encrypts the line.
4. **Ask.** HTTP sends `GET /healthz` and gets back `200 OK` and a body.

Each of the four can fail on its own, with its own error. The last section
of this guide turns this list into a debugging checklist.

## Layers: each one solves one problem

Networks are built in *layers*. Each layer uses the one below it without
caring how it works, just as you post a letter without knowing the truck's
route. (That's abstraction; see the *Abstraction* guide.)

| Layer | Its one job | Addresses by | Examples |
|---|---|---|---|
| Application | what the bytes mean | URL, path | HTTP, DNS, SSH |
| Transport | reach the right program | port | TCP, UDP |
| Internet | reach the right machine | IP address | IPv4, IPv6 |
| Link | reach the next hop on the wire | MAC address | Ethernet, Wi-Fi |

When you send data, each layer wraps the one above it in its own envelope.
This is called *encapsulation*.

```text
+---------------------------------------------------------+
| Ethernet header                                         |  link
|  +---------------------------------------------------+  |
|  | IP header: from 10.0.0.5 to 203.0.113.10          |  |  internet
|  |  +---------------------------------------------+  |  |
|  |  | TCP header: from port 51234 to port 443     |  |  |  transport
|  |  |  +---------------------------------------+  |  |  |
|  |  |  | GET /healthz HTTP/1.1 ...             |  |  |  |  application
|  |  |  +---------------------------------------+  |  |  |
|  |  +---------------------------------------------+  |  |
|  +---------------------------------------------------+  |
+---------------------------------------------------------+
```

Routers on the way only open the outer envelopes to pick the next hop. Only
the two ends read the TCP and HTTP parts. People also say "layer 4" for
transport and "layer 7" for application, after the older seven-layer OSI
model.

## IP addresses

An *IP address* identifies a machine's network interface, the way a street
address identifies a house.

### IPv4 and IPv6

- **IPv4**: 32 bits, written as four numbers from 0 to 255:
  `192.168.1.10`. That allows about 4.3 billion addresses, which the
  internet ran out of.
- **IPv6**: 128 bits, written in hex groups: `2606:4700:10::6814:179a`. The
  `::` means "a run of zero groups".

### CIDR: a block of addresses

You rarely deal with one address. Firewalls, VPC subnets and Kubernetes pod
networks all deal in **blocks**, written as a *CIDR block*: `10.0.0.0/24`.

The `/24` says how many leading bits are fixed. An IPv4 address has 32 bits,
so 32 − 24 = 8 bits are free, and 2⁸ = 256 addresses fit in the block:
`10.0.0.0` to `10.0.0.255`.

| CIDR | Free bits | Addresses | Typical use |
|---|---|---|---|
| `10.0.0.5/32` | 0 | 1 | exactly one host in a firewall rule |
| `10.0.0.0/24` | 8 | 256 | a subnet |
| `10.0.0.0/16` | 16 | 65,536 | an AWS VPC |
| `0.0.0.0/0` | 32 | all of them | "anywhere": the whole internet |

Each extra bit after the slash halves the block. In a security group,
`0.0.0.0/0` means *everyone*. In AWS, each subnet also keeps 5 addresses for
itself, so a `/24` gives you 251 usable ones.

### Private and special ranges

Some ranges are reserved for private networks. They never appear on the
public internet, so every company can reuse them.

| Range | What it is |
|---|---|
| `10.0.0.0/8` | private: common for VPCs and clusters |
| `172.16.0.0/12` | private: Docker's default networks live here |
| `192.168.0.0/16` | private: home and office routers |
| `127.0.0.0/8` | loopback: "this machine" (`127.0.0.1`, `localhost`) |
| `169.254.0.0/16` | link-local: AWS's instance metadata is `169.254.169.254` |

A private machine reaches the internet through *NAT* (network address
translation): a gateway swaps the private source address for its own public
one on the way out.

**Ops trap:** two networks that use overlapping ranges (say both
`10.0.0.0/16`) can't easily be peered or joined by VPN. Plan ranges early.

### In Go

Go's `net/netip` package does this maths for you, so you never split strings
on dots:

```go
package main

import (
	"fmt"
	"net/netip"
)

func main() {
	subnet := netip.MustParsePrefix("10.0.0.0/24")
	fmt.Println("subnet:", subnet, "prefix bits:", subnet.Bits())

	for _, s := range []string{"10.0.0.42", "10.0.1.5"} {
		ip := netip.MustParseAddr(s)
		fmt.Println(ip, "in", subnet, "?", subnet.Contains(ip))
	}

	// A rule written sloppily: host bits set. Masked() cleans it up.
	sloppy := netip.MustParsePrefix("10.0.0.77/24")
	fmt.Println(sloppy, "is really", sloppy.Masked())

	addrs := []string{"192.168.1.10", "172.20.0.5", "8.8.8.8", "fd00::1"}
	for _, s := range addrs {
		ip := netip.MustParseAddr(s)
		fmt.Printf("%-12s IPv4=%-5v private=%v\n", s, ip.Is4(), ip.IsPrivate())
	}
}
```

It prints:

```text
subnet: 10.0.0.0/24 prefix bits: 24
10.0.0.42 in 10.0.0.0/24 ? true
10.0.1.5 in 10.0.0.0/24 ? false
10.0.0.77/24 is really 10.0.0.0/24
192.168.1.10 IPv4=true  private=true
172.20.0.5   IPv4=true  private=true
8.8.8.8      IPv4=true  private=false
fd00::1      IPv4=false private=true
```

`10.0.1.5` is outside because the first 24 bits (the first three numbers)
must match. `fd00::1` is in IPv6's private range. You'll use this in
*IP addresses* and *CIDRs*.

## Ports and sockets

One machine runs many network programs: a web server, SSH, a database. The
IP address gets packets to the machine; a *port* (a number from 0 to 65535)
gets them to the right program.

A server *listens* on a port: it asks the operating system "give me whatever
arrives on port 443". Only one program can listen on a given address and
port, which is why you see `address already in use`.

| Port | Usually |
|---|---|
| 22 | SSH |
| 53 | DNS |
| 80 / 443 | HTTP / HTTPS |
| 5432 / 6443 | PostgreSQL / Kubernetes API server |

A *socket* is one end of a conversation: an IP address plus a port, owned by
a process. A TCP connection is identified by both ends. The client's side
uses an *ephemeral port*: a temporary port the OS picks from a high range
(32768–60999 on Linux), which is how one laptop holds many connections to
the same server.

```go
package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	// Become a server: ask the OS for any free port on this machine.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Println("listen:", err)
		return
	}
	addr := ln.Addr().String()
	fmt.Println("listening on", addr)

	// Be a client: connect to that port. The OS completes the handshake.
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		fmt.Println("dial:", err)
		return
	}
	fmt.Println("connected:", conn.LocalAddr(), "->", conn.RemoteAddr())
	conn.Close()

	// Stop listening, then try again: nobody is on that port now.
	ln.Close()
	_, err = net.DialTimeout("tcp", addr, 2*time.Second)
	fmt.Println("after close:", err)
}
```

One run printed (your port numbers will differ):

```text
listening on 127.0.0.1:45993
connected: 127.0.0.1:40928 -> 127.0.0.1:45993
after close: dial tcp 127.0.0.1:45993: connect: connection refused
```

Step by step:

1. Port `0` means "any free port"; the OS picked 45993.
2. The client connects from ephemeral port 40928 to 45993.
3. After the listener closes, the OS refuses new connections. **Connection
   refused** means "the machine is there, but nothing listens on that port".

**Ops trap: `127.0.0.1` vs `0.0.0.0`.** A server listening on
`127.0.0.1:8080` accepts connections from the same machine only. In a
container, the app answers `curl` inside the pod and nothing else. Listen on
`0.0.0.0:8080` (or `:8080` in Go) to accept on every interface.

## TCP and UDP

The transport layer offers two very different deals.

### TCP: a reliable conversation

*TCP* gives you a **connection**: an ordered, reliable stream of bytes. It
starts with a *three-way handshake*:

```text
client                                   server (listening on 443)
  |  ---- SYN: "can we talk?" ------------->  |
  |  <--- SYN-ACK: "yes, can you hear me?" -  |
  |  ---- ACK: "yes" ---------------------->  |
  |        connection open: send data         |
```

After that, TCP numbers every byte, re-sends anything lost, and delivers the
bytes in order. You pay one round trip to set up, and some delay when
packets get lost.

A connection attempt ends in one of three ways:

| Result | What happened | Usual cause |
|---|---|---|
| connected | handshake finished | the port is open |
| connection refused | the machine replied "nobody here" | service down, wrong port |
| timed out | no reply at all | firewall dropping packets, wrong IP |

### UDP: fire and forget

*UDP* sends single messages (datagrams). There's no connection, no
re-sending and no ordering. Each message arrives whole or not at all, and
the sender is never told which.

That's faster and simpler when a lost message doesn't matter, or when the
application handles loss itself. DNS lookups, StatsD metrics, video calls
and HTTP/3 (via QUIC) all use UDP.

## DNS: names to addresses

Nobody remembers `203.0.113.10`. *DNS* (the Domain Name System) is the
internet's phone book: it maps names like `api.example.com` to addresses.

```text
your program
   |  "api.example.com?"
   v
stub resolver on your machine (reads /etc/resolv.conf)
   |
   v
recursive resolver (your company's, AWS's, or 8.8.8.8)
   |  not cached? asks around:
   +--> root servers:   "ask the .com servers"
   +--> .com servers:   "ask example.com's name servers"
   +--> example.com's own server: "203.0.113.10, TTL 300"
   |
   v
answer cached, then returned to your program
```

The domain's own (authoritative) server is the source of truth: Route 53,
for example, if your team hosts the zone there. A *resolver* does the
legwork and caches answers.

| Record | Maps a name to | Example |
|---|---|---|
| A | an IPv4 address | `api.example.com → 203.0.113.10` |
| AAAA | an IPv6 address | `api.example.com → 2001:db8::10` |
| CNAME | another name | `www.example.com → example.com` |

### TTLs and caching

Every answer carries a *TTL* (time to live): how many seconds resolvers may
cache it. That's why "I changed the record, but half the users still hit
the old server" happens: their resolvers keep the old answer until its TTL
runs out. Before a migration, lower the TTL (say to 60 seconds) at least
one old TTL ahead: if it was 86400 (a day), lower it a day early. Then the
switch spreads within a minute.

### Check DNS early

"It's always DNS" is an ops joke. It isn't always, but DNS sits under almost
everything. In Kubernetes, every Service gets a name like
`payments.prod.svc.cluster.local`, answered by the cluster's DNS server
(CoreDNS). When that struggles, every service that calls others by name
breaks at once.

## HTTP: requests and responses

*HTTP* is the protocol of APIs, web UIs, Kubernetes, AWS and most webhooks.
It's text-shaped: a request, then a response. Here is a real exchange with
a small Go server, as HTTP/1.1 sends it:

```text
GET /healthz HTTP/1.1
Host: 127.0.0.1:18089
User-Agent: curl/8.22.0
Accept: */*

```

```text
HTTP/1.1 200 OK
Content-Type: application/json
Date: Tue, 13 Oct 2026 22:44:56 GMT
Content-Length: 15

{"status":"ok"}
```

| Part | In the example | Meaning |
|---|---|---|
| method | `GET` | the action: GET reads, POST creates, DELETE removes |
| path | `/healthz` | which resource on the server |
| headers | `Host:`, `Accept:` | extra facts, one `Name: value` per line |
| body | (none in the request) | data sent with POST or PUT, often JSON |

A blank line separates headers from body. The response has the same shape,
but its first line holds the *status code*.

### Status codes

The first digit tells you whose problem it is:

| Class | Meaning | Ones you'll see |
|---|---|---|
| 2xx | it worked | 200 OK, 201 Created, 204 No Content |
| 3xx | look elsewhere | 301 / 302 redirect, 304 Not Modified |
| 4xx | **the client** got it wrong | 400, 401 (who are you?), 403 (not allowed), 404, 429 (slow down) |
| 5xx | **the server** failed | 500, 502 Bad Gateway, 503 Service Unavailable, 504 Gateway Timeout |

Two rules to remember:

- **Retry 429 and 5xx with backoff, but only requests that are safe to
  repeat.** A GET is safe; a POST that creates something may have worked
  before the error, so retrying it can do it twice. Don't retry other 4xx:
  a 403 won't turn into a 200 by asking again.
- **502 and 504 usually come from a proxy** in front of the app: the app
  behind it failed or was too slow.

HTTP/2 and HTTP/3 change how the bytes travel, not what they mean: methods,
paths, headers, status codes and bodies are the same. You'll write both
sides in *HTTP clients* and *HTTP servers*.

## TLS: privacy and identity

Plain HTTP crosses the network as readable text. Anyone on the path can read
or change it. *TLS* (the S in HTTPS) adds two things:

1. **Encryption:** only the two ends can read the bytes.
2. **Identity:** the server proves it really is `api.example.com`.

Identity comes from a *certificate*: a document saying "this public key
belongs to `api.example.com`, valid until 25 December", signed by a
*certificate authority* (CA). Your machine ships with a list of CAs it
trusts. A certificate is trusted if its chain of signatures leads back to
one of them.

```text
trusted root CA (in your OS trust store)
    | signs
intermediate CA              <-- the server sends this one too
    | signs
api.example.com certificate  <-- the server sends this during the handshake
```

Your client checks three things: the chain leads to a trusted CA, the name
matches the host you asked for, and today is within the validity dates.

TLS failures you'll meet on call:

- **Expired certificate.** Certificates are short-lived, often 90 days or
  less. Forgetting to renew is a classic outage: automate renewal and alert
  before expiry.
- **Name mismatch.** The certificate says `api.example.com`, but you
  connected to `10.0.0.5` or `api-internal`.
- **Unknown CA.** An internal service uses a private CA your client doesn't
  trust. Add that CA to the client's trust; **don't** turn verification off.
  You'll practise both in *TLS*.

## Latency and bandwidth

A network can be "slow" in two different ways:

- *Latency* is how long one trip takes. There and back is the round-trip
  time (RTT).
- *Bandwidth* is how many bytes per second the path carries.

On a motorway, bandwidth is the lanes and latency is the drive time. More
lanes don't get one car there sooner.

Latency has a hard floor: light in fibre covers about 200 km per
millisecond. New York to London is about 5,600 km, so a round trip can never
beat about 56 ms.

| Path | Typical round trip |
|---|---|
| same machine (`127.0.0.1`) | microseconds |
| same data centre or AWS zone | under 1 ms |
| across a continent | 30–80 ms |
| across an ocean | 70–200 ms |

**Why this matters to your code:** a new HTTPS connection spends about one
round trip on TCP and one on TLS 1.3 (two on older TLS 1.2) before the
request even leaves. And 100 API calls made one after another, at 80 ms
each, take 8 seconds however fat the link is. The fixes are reusing
connections (Go's HTTP client does this for you, as long as you read and
close every response body), batching, and overlapping calls; see the
*Concurrency* guide.

## Timeouts everywhere

On a network, "no answer" is normal. Packets vanish, a firewall swallows
your connection attempt, a server accepts and then freezes. Without a
timeout, your program waits as long as that takes: possibly forever.

| Stage | What can hang | With no timeout set |
|---|---|---|
| TCP connect | a firewall dropping packets | about 2 minutes on Linux |
| TLS handshake | a stuck server | until the connection dies |
| reading the response | a server that never answers | **forever** |

A Go trap: `http.Get` uses the default client, which has **no overall
timeout**. Connecting and TLS do give up (after 30 and 10 seconds), but if
a dependency accepts the connection and never answers, the request waits
forever. Stuck requests pile up until your service runs out of memory or
connections. Build your own client:

```go
client := &http.Client{Timeout: 10 * time.Second}
resp, err := client.Get("https://api.example.com/healthz")
```

`Timeout` covers the whole exchange: connect, TLS, request and body. It turns
"hangs forever" into an error you can log, retry or alert on. The
*Concurrency* module adds contexts, which carry one deadline through a whole
chain of calls.

## Load balancers and proxies

You rarely talk to one server directly. A *load balancer* gives many
identical backends one address, spreads requests across them, and skips the
ones failing their health checks.

```text
                               +--> app-1 (healthy)
clients --> load balancer -----+--> app-2 (healthy)
            one address        +--> app-3 (failing health check: skipped)
```

| Kind | Looks at | Examples |
|---|---|---|
| layer 4 | IPs and ports; passes TCP through | AWS NLB, Kubernetes Service |
| layer 7 | HTTP: host, path, headers | AWS ALB, nginx, Kubernetes Ingress |

A *reverse proxy* sits in front of your app and forwards requests to it,
often ending TLS and adding `X-Forwarded-For` (the real client IP). So your
app sees the proxy's IP unless it reads that header. The proxy also has its
own timeouts: a 504 can mean "the app took 61 s and the proxy gave up at
60 s", even though the app finished.

## Debugging "can't connect", step by step

Walk up the layers. Each step rules out a whole class of problems.

```text
1. DNS        does the name resolve?          dig +short NAME
2. Route      does anything answer?           nc -vz HOST PORT
3. Port       is the service listening?       ss -ltnp   (on the server)
4. TLS        is the certificate OK?          curl -v https://...
5. HTTP       what status comes back?         curl -sS -o /dev/null -w ...
```

**1. DNS.** Does the name resolve, to the address you expect?

```bash
dig +short api.example.com
```

No output means no answer. When resolution fails, `curl` says:

```text
curl: (6) Could not resolve host: nosuchhost.invalid
```

Check the spelling, the record, and which resolver you use (in a pod, read
`/etc/resolv.conf`).

**2. Route, from the client.** Try a bare TCP connection with a 3-second
limit:

```bash
nc -vz -w 3 db.internal 5432
```

The three outcomes look like this (from three different targets):

```text
Connection to 127.0.0.1 15432 port [tcp/*] succeeded!
nc: connect to 127.0.0.1 port 9 (tcp) failed: Connection refused
nc: connect to 10.255.255.1 port 5432 (tcp) failed: Connection timed out
```

- **Succeeded:** the network is fine; look higher up.
- **Refused:** you reached the machine, but nothing listens on that port.
- **Timed out:** nothing answered. Suspect a security group, firewall,
  network policy, or a wrong IP.

**3. Listening, on the server.** Run `ss -ltnp` and find your port.
`127.0.0.1:8080` means local only; you want `0.0.0.0:8080` or `*:8080`.

**4. TLS.** `curl -v` prints the certificate's subject, issuer and dates. A
failed check looks like this:

```text
curl: (60) SSL certificate OpenSSL verify result: self-signed certificate (18)
```

Check the expiry, the name, and whether your client trusts the CA.

**5. HTTP.** You're connected; what does the app say? This prints only the
status code, such as `404`:

```bash
curl -sS -o /dev/null -w '%{http_code}\n' https://example.com/nope
```

Now it's an application question: 4xx means check your request (URL, auth,
body); 5xx means read the server's logs.

Write down what each step showed. "DNS ok, TCP connect times out from the
pod but works from the node" is something the network team can act on. "It's
broken" isn't.

## Common mistakes

- **No timeout on a network call.** `http.Get` and bare connections can hang
  forever. Set `http.Client.Timeout` or a deadline every time.
- **Turning off certificate checks** (`InsecureSkipVerify: true`) to "fix" a
  TLS error. That throws away the identity half of TLS. Trust the right CA
  instead.
- **Listening on `127.0.0.1` in a container,** then wondering why nothing
  outside can connect.
- **Retrying every error.** Retrying a 400 or 403 repeats the mistake,
  faster. Retry timeouts, 5xx and 429 with backoff, and only for requests
  that are safe to repeat.
- **Opening `0.0.0.0/0`** in a security group "just to test", and leaving it.

## Check yourself

1. How many addresses are in `10.20.0.0/22`, and is `10.20.3.200` inside it?
2. `nc -vz 10.0.3.7 5432` hangs, then says "Connection timed out". What do
   you suspect first, and what would "Connection refused" have meant?
3. You moved a service to a new IP an hour ago. Some clients still reach
   the old one. What's the likely cause, and how could you have avoided it?
4. Your Go service calls a dependency with `http.Get`. The dependency starts
   accepting connections but never replying. What happens to your service?
5. A web app in a pod answers `curl localhost:8080` from inside the pod but
   not from another pod. What's the first thing to check?
6. A Go client calling an internal service fails with `x509: certificate
   signed by unknown authority`. A teammate suggests
   `InsecureSkipVerify: true`. Why is that wrong, and what's the right fix?
7. Your client sends a POST that starts a deploy and gets back a 502. Should
   it retry automatically?

## Answers

1. 1,024 addresses: 32 − 22 = 10 free bits, and 2¹⁰ = 1,024. Ten free bits
   are the 8 bits of the last number plus 2 bits of the third, so the third
   number runs 0–3: the block is `10.20.0.0` to `10.20.3.255`, and
   `10.20.3.200` is inside.
2. A timeout means nothing answered, so suspect something dropping packets:
   a security group, firewall, network policy, or a wrong IP. "Refused"
   would mean the machine answered but nothing listens on 5432.
3. DNS caching: resolvers keep the old record until its TTL expires. Lower
   the TTL at least one old TTL before a migration so the change spreads
   within minutes.
4. With no timeout, each call waits forever. Stuck requests pile up until
   the service runs out of connections or memory, so one slow dependency
   takes you down too. Use an `http.Client` with a `Timeout`.
5. The address the app listens on. On `127.0.0.1:8080` only the pod itself
   can connect; it should listen on `0.0.0.0:8080` (or `:8080`).
6. Skipping verification keeps the encryption but drops the identity check,
   so anyone in the middle can pretend to be the service. The service uses
   a private CA your client doesn't trust: add that CA to the client's
   trusted list instead.
7. Not blindly. A 502 comes from a proxy that didn't get a good answer from
   the app, but the app may already have started the deploy, so a retry
   could start a second one. Retry only requests that are safe to repeat,
   or check the deploy's state first.

## Key terms

| Term | Meaning |
|---|---|
| layer | one level of the network stack, solving one problem |
| encapsulation | each layer wrapping the data from the layer above |
| IP address | identifies a machine's network interface |
| CIDR block | a range of addresses written as address/prefix bits |
| NAT | swapping private addresses for a public one at a gateway |
| port | a number (0–65535) that picks the program on a machine |
| socket | one end of a connection: IP address plus port |
| ephemeral port | the temporary client-side port the OS picks for a connection |
| TCP | reliable, ordered byte stream over a connection |
| three-way handshake | SYN, SYN-ACK, ACK: how TCP opens a connection |
| UDP | single messages, no connection, no delivery guarantee |
| DNS | the system that maps names to addresses |
| resolver | the server that looks names up and caches answers |
| TTL | how many seconds a DNS answer may be cached |
| HTTP | the request/response protocol of APIs and the web |
| status code | the 3-digit result of an HTTP request |
| TLS | encryption plus server identity (the S in HTTPS) |
| certificate | a signed statement that a key belongs to a name |
| certificate authority | an organisation whose signature clients trust |
| latency / bandwidth | time for one trip / bytes per second |
| load balancer | spreads requests from one address across backends |
| reverse proxy | a server that forwards requests to your app |

## Watch

- [How the Internet Works in 5 Minutes](https://www.youtube.com/watch?v=7_LPdttKXPc) · Aaron · 5 min: packets, IP addresses, routers and DNS in one quick pass, a good first picture of the whole stack.
- [The Internet: IP Addresses & DNS](https://www.youtube.com/watch?v=5o8CwafCxnU) · CodeAI · 7 min: Vint Cerf, a co-inventor of the internet, explains addresses and how DNS turns names into them.
- [SSL, TLS, HTTPS Explained](https://www.youtube.com/watch?v=j9QmMEWmcfo) · ByteByteGo · 6 min: an animated TLS handshake showing how certificates prove identity and keys make the connection private.
- [CS50x - Lecture 8 - HTML, CSS, JavaScript](https://www.youtube.com/watch?v=yYst7puZXjw) · CS50 · 142 min: watch only the first 42 minutes, Harvard's lively tour of TCP/IP, ports, DNS, DHCP, HTTP and status codes; the rest is web pages.

## Go deeper

- [An overview of HTTP (MDN)](https://developer.mozilla.org/en-US/docs/Web/HTTP/Guides/Overview):
  a clear, beginner-friendly tour of requests, responses and headers.
- [What is DNS? (Cloudflare)](https://www.cloudflare.com/learning/dns/what-is-dns/):
  the lookup chain from this guide, with diagrams.
- [The complete guide to Go net/http timeouts](https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/):
  exactly which timeout covers which stage, for clients and servers.
- [High Performance Browser Networking](https://hpbn.co/): a free book; its
  TCP, UDP and TLS chapters explain latency and handshakes in depth.
- [Beej's Guide to Network Programming](https://beej.us/guide/bgnet/): the
  classic friendly guide to sockets, ports and connections.
