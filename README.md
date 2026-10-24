<div align="center">

# 🚀 **ChPortScanner v1.0.0**

### *High-Speed Port Scanner for Restricted Lab Environments*

**Fast • Modular • Zero-Dependency • Cross-Platform • No Root Required**

<br/>

[![Version](https://img.shields.io/badge/version-1.0.0-blue?style=for-the-badge&logo=go)](https://github.com/Ch4120N/ChPortScanner)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)
[![Owner](https://img.shields.io/badge/Owner-Ch4120N-red?style=for-the-badge)](https://github.com/Ch4120N)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey?style=for-the-badge)](https://github.com/Ch4120N/ChPortScanner)

[![Stars](https://img.shields.io/github/stars/Ch4120N/ChPortScanner?style=social)](https://github.com/Ch4120N/ChPortScanner/stargazers)
[![Forks](https://img.shields.io/github/forks/Ch4120N/ChPortScanner?style=social)](https://github.com/Ch4120N/ChPortScanner/network/members)

</div>

---


## 📑 Table of Contents

- [🔥 What is ChPortScanner?](#-what-is-ChPortScanner)
- [✨ Key Features](#-key-features)
- [🏎️ Performance Highlights](#️-performance-highlights)
- [📦 Installation](#-installation)
- [🚀 Quick Start](#-quick-start)
- [📖 Usage](#-usage)
  - [Targeting](#-targeting)
  - [Port Selection](#-port-selection)
  - [Scanning Options](#-scanning-options)
  - [Probing](#-probing)
  - [Output Formats](#-output-formats)
  - [Utilities](#-utilities)
- [⚡ Performance Tuning](#-performance-tuning)
- [🎯 Use Cases](#-use-cases)
- [🏗️ Architecture](#️-architecture)
- [💡 Examples](#-examples)
- [🛠️ Building from Source](#️-building-from-source)
- [❓ FAQ](#-faq)
- [📜 License](#-license)
- [👤 Author](#-author)

---


## 🔥 What is ChPortScanner?

**ChPortScanner** is a **blazing-fast, modular TCP/UDP port scanner** built in pure **Go**, purpose-built for **restricted lab environments** where tools like `masscan` or `nmap` are blocked by policy, IDS signatures, or lack of root privileges.

It uses **TCP connect scanning** — meaning it **needs no root**, works in **unprivileged containers**, and looks like **normal traffic** to network sensors — while still hitting **thousands of ports per second** on local networks.

> 🎯 **Mission:** Give security researchers, penetration testers, and lab engineers a fast, safe, policy-compliant port scanner that runs anywhere — from a Raspberry Pi to a datacenter VM.

<div align="center">

### 🎁 **Written in Go — Zero External Dependencies**

*Single binary • Static compile • Cross-platform*

</div>

---


## ✨ Key Features

<table>
<tr>
<td width="50%">

### 🎯 **Powerful Targeting**
- ✅ **CIDR** notation — `192.168.1.0/24`
- ✅ **Ranges** — `10.0.0.1-10.0.0.255`
- ✅ **Single IPs** — `192.168.1.1`
- ✅ **IPv4 & IPv6** support
- ✅ **File input** (`-iL targets.txt`)
- ✅ **Exclude lists** (`--exclude`)
- ✅ **Sharding** (`--shard 1/4`) for distributed scans
- ✅ **Randomized** scan order (`--random`)

</td>
<td width="50%">

### 🔍 **Advanced Probing**
- ✅ **Banner grabbing** — grab service banners
- ✅ **HTTP title** — `--http-probe`
- ✅ **TLS certificate** — CN, issuer, expiry
- ✅ **Reverse DNS** — `--rdns`
- ✅ **Service name** lookup (100+ ports)
- ✅ **Retry** logic with jitter
- ✅ **UDP scan** mode (`--udp`)

</td>
</tr>
<tr>
<td>

### ⚡ **Speed & Efficiency**
- ✅ **Auto-tuned worker pool** (CPU-aware)
- ✅ **`SO_LINGER=0`** — no `TIME_WAIT` buildup
- ✅ **Zero-allocation hot path**
- ✅ **Batched token bucket** rate limiter
- ✅ **Buffered output** with async flush
- ✅ **Auto-raises `ulimit -n`** on Unix
- ✅ **`--lite` / `--turbo`** modes
- ✅ **Resume** interrupted scans

</td>
<td>

### 🎨 **Beautiful Output**
- ✅ **5 formats**: `text`, `json`, `jsonl`, `csv`, `grep`
- ✅ **Live progress bar** with ETA
- ✅ **Colored output** (disable with `--no-color`)
- ✅ **Quiet** / **Verbose** modes
- ✅ **Final statistics** (`--stats`)
- ✅ **State file** auto-save

</td>
</tr>
</table>

---

## 🏎️ Performance Highlights

<div align="center">

| Environment | Target | Ports | Time | Throughput |
|:------------|:-------|:------|:-----|:-----------|
| 🖥️ **Modern PC** (8-core, LAN) | `/24` (256 hosts) | `1-65535` | **~45 s** | **~370k pps** |
| 💻 **Laptop** (4-core, LAN) | `/24` (256 hosts) | `1-10000` | **~7 s** | **~340k pps** |
| 🍓 **Raspberry Pi 4** | `/24` (256 hosts) | `1-1000` | **~12 s** | **~21k pps** |
| 🐳 **Docker** (2 vCPU) | `/24` (256 hosts) | `Top 1000` | **~15 s** | **~17k pps** |

</div>

> 💡 **Note:** Actual throughput depends on **timeout**, **rate limit**, and **network latency**. Use `--turbo` and raise `ulimit` for maximum performance.

**Optimizations that make it fast:**

| Optimization | Impact |
|:-------------|:-------|
| 🔥 `SO_LINGER=0` | Prevents `TIME_WAIT` exhaustion (biggest win) |
| ⚙️ Auto worker sizing | 200 → 6,000 workers based on CPU count |
| 🔄 Zero-alloc address building | No `JoinHostPort` per dial |
| 📊 Batched rate limiter | 100× fewer timer fires than per-token |
| 📝 64 KB buffered output | No syscall per result |
| 🔧 `KeepAlive: -1` | Saves TCP keepalive syscalls |

---

## 📦 Installation

### 📥 **Pre-built Binaries**

Download the latest release for your platform from the [**Releases page**](https://github.com/Ch4120N/ChPortScanner/releases).

```bash
# Linux (amd64)
wget https://github.com/Ch4120N/ChPortScanner/releases/download/v1.0.0/ChPortScanner-linux-amd64
chmod +x ChPortScanner-linux-amd64
sudo mv ChPortScanner-linux-amd64 /usr/local/bin/ChPortScanner

# macOS (Apple Silicon)
curl -LO https://github.com/Ch4120N/ChPortScanner/releases/download/v1.0.0/ChPortScanner-darwin-arm64
chmod +x ChPortScanner-darwin-arm64
sudo mv ChPortScanner-darwin-arm64 /usr/local/bin/ChPortScanner
```

### 🐹 **Install via Go**

```bash
go install github.com/Ch4120N/ChPortScanner@latest
```

### 🛠️ **Build from Source**

```bash
git clone https://github.com/Ch4120N/ChPortScanner.git
cd ChPortScanner
go mod tidy

# Standard build
go build -ldflags="-s -w" -o ChPortScanner

# Static build (fully portable, no CGO)
CGO_ENABLED=0 go build -ldflags="-s -w" -o ChPortScanner

# Cross-compile
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ChPortScanner.exe
GOOS=linux   GOARCH=arm64 go build -ldflags="-s -w" -o ChPortScanner-arm64
```

### ✅ **Verify Installation**

```bash
ChPortScanner --version
# ChPortScanner v1.0.0 by Ch4120N
```

---

## 🚀 Quick Start

**Scan a subnet for common ports:**

```bash
ChPortScanner -t 192.168.1.0/24 -p 22,80,443
```

**Scan all 65535 ports on a single host — fast:**

```bash
ChPortScanner -t 192.168.1.10 -p 1-65535 --turbo --stats
```

**Full recon with probes:**

```bash
ChPortScanner -t 10.0.0.0/24 --top-ports 1000 \
  --http-probe --tls-cert --rdns -oJ -o results.jsonl
```

**Distributed scan across 4 machines:**

```bash
# On machine 1
ChPortScanner -t 192.168.0.0/16 -p 22,80 --shard 1/4 -oJ -o part1.jsonl
# On machine 2
ChPortScanner -t 192.168.0.0/16 -p 22,80 --shard 2/4 -oJ -o part2.jsonl
# ... machines 3 & 4 with --shard 3/4 and 4/4
```

---

## 📖 Usage

```
ChPortScanner [options] -t <IP/CIDR/range>
```

### 🎯 Targeting

| Flag | Description |
|:-----|:------------|
| `-t, --target <spec>` | Target IPs / CIDRs / ranges (comma-separated) |
| `-iL, --input <file>` | Input file with targets (one per line, `#` for comments) |
| `--exclude <file>` | File with IPs/CIDRs to exclude from scanning |
| `-sn` | Ping sweep only (host discovery, no port scan) |

**Examples:**

```bash
# Multiple target types at once
ChPortScanner -t "192.168.1.0/24,10.0.0.5,172.16.1.1-172.16.1.50" -p 80,443

# From file
ChPortScanner -iL targets.txt -p 22

# Exclude internal infrastructure
ChPortScanner -t 10.0.0.0/8 --exclude skip.txt -p 80
```

### 🔌 Port Selection

| Flag | Description |
|:-----|:------------|
| `-p, --ports <spec>` | Ports to scan (e.g. `80,443,1000-2000`) |
| `--top-ports <N>` | Scan top N most common ports |
| `--exclude-ports <spec>` | Ports to exclude from the scan |
| `--profile <name>` | Preset: `quick`, `default`, `full`, `stealth` |

### ⚙️ Scanning Options

| Flag | Default | Description |
|:-----|:--------|:------------|
| `--rate <N>` | `0` (unlimited) | Packets per second |
| `--workers <N>` | `0` (auto) | Concurrent worker count |
| `--timeout <N>` | `1` | Connection timeout (seconds) |
| `--retries <N>` | `0` | Retries for failed connections |
| `--retry-delay <N>` | `500` | Delay between retries (ms) |
| `-u, --udp` | `false` | UDP scan mode |
| `--random` | `false` | Randomize scan order |
| `--shard X/Y` | — | Scan shard (e.g., `1/4`) |
| `-S <IP>` | — | Source IP to bind |
| `--lite` | `false` | Low-resource mode (low-end PCs) |
| `--turbo` | `false` | Maximum speed mode |

### 🔍 Probing

| Flag | Description |
|:-----|:------------|
| `--banner` | Enable banner grabbing |
| `--banner-timeout <N>` | Banner read timeout in seconds (default: `2`) |
| `--http-probe` | Fetch HTTP page titles |
| `--tls-cert` | Extract TLS certificate information |
| `--rdns` | Reverse DNS lookups on open hosts |

### 📤 Output Formats

| Flag | Description |
|:-----|:------------|
| `-o, --output <file>` | Write results to file (default: stdout) |
| `-oJ` | JSON output (shortcut for `--format json`) |
| `--format <fmt>` | `text` \| `json` \| `jsonl` \| `csv` \| `grep` |
| `--no-color` | Disable colored output |
| `-q, --quiet` | Suppress banner and info messages |
| `-V, --verbose` | Verbose mode |
| `--stats` | Show final statistics summary |

### 🧰 Utilities

| Flag | Description |
|:-----|:------------|
| `--extract <file>` | Extract unique IPs from an output file |
| `--extract-ports` | Keep IP:port pairs during extraction |
| `--resume <state>` | Resume from a saved state file |
| `--state <file>` | State file path (default: `scan.state`) |
| `-c <file>` | Load configuration from JSON file |
| `-h, --help` | Show help |
| `-v, --version` | Show version |

---

## ⚡ Performance Tuning

### 🔧 **On Linux / macOS — Raise File Descriptor Limit**

The **#1 bottleneck** for fast scanning is **open file descriptors**:

```bash
# Check current limit
ulimit -n

# Raise it (session only)
ulimit -n 65535

# Permanent (add to ~/.bashrc or ~/.zshrc)
echo "ulimit -n 65535" >> ~/.zshrc
```

> 💡 ChPortScanner **auto-tries** to raise this on startup on Unix, but the OS may require you to do it manually.

### 🎛️ **Rate Limiting Strategy**

| Goal | Recommended |
|:-----|:------------|
| 🏃 **Maximum speed** | `--rate 0 --turbo` |
| ⚖️ **Balanced** | `--rate 5000` |
| 🥷 **Stealth** | `--profile stealth` |
| 🖥️ **Low-end PC** | `--lite --workers 200` |

### 🖥️ **Low-End PC Settings**

For **2-core / <4 GB RAM** systems:

```bash
ChPortScanner -t 192.168.1.0/24 --top-ports 1000 --lite
```

This auto-sets **200 workers**, **5k job buffer**, and **disables heavy probes**.

---

## 🎯 Use Cases

<table>
<tr>
<td width="33%" align="center">

### 🔬 **Security Research**

Discover exposed services on lab networks without privileged access.

</td>
<td width="33%" align="center">

### 🎓 **Training & Education**

Perfect for CTF labs, cybersecurity courses, and hands-on learning.

</td>
<td width="33%" align="center">

### 🏢 **Network Auditing**

Inventory services across internal networks, find forgotten hosts.

</td>
</tr>
<tr>
<td align="center">

### 🐳 **Container Environments**

Runs unprivileged in Docker, Kubernetes, and CI/CD pipelines.

</td>
<td align="center">

### 🧪 **Pentest Recon**

Fast initial discovery phase — then hand off to deeper tools.

</td>
<td align="center">

### 🏠 **Home Lab**

Scan your own infra without needing `sudo` or raw sockets.

</td>
</tr>
</table>

---

## 🏗️ Architecture

**ChPortScanner** is built with **clean modular design** — every package is independently testable and reusable.

```
ChPortScanner/
│
├── main.go                          # Entry point + signal handling
├── go.mod
├── go.sum
│
└── internal/
    │
    ├── config/                      # CLI flags, JSON config, profiles
    │   └── config.go
    │
    ├── target/                      # CIDR / range / IP parsing
    │   └── target.go
    │
    ├── port/                        # Port ranges + service names
    │   └── port.go
    │
    ├── netutil/                     # Socket options (cross-platform)
    │   ├── sockopt_unix.go          # SO_LINGER=0, ulimit
    │   └── sockopt_windows.go       # no-op fallbacks
    │
    ├── probe/                       # HTTP title, TLS, rDNS
    │   └── probe.go
    │
    ├── output/                      # Writers: text/json/jsonl/csv/grep
    │   └── output.go
    │
    ├── ui/                          # Banner, progress, help, colors
    │   └── ui.go
    │
    └── scanner/                     # Core scan engine
        ├── scanner.go               # Orchestration
        ├── producer.go              # Job generation (IP × port)
        ├── worker.go                # TCP/UDP dial + probes
        └── state.go                 # Resume state persistence
```

### 🔑 **Design Principles**

1. **🎯 Separation of concerns** — Config, parsing, scanning, output are decoupled
2. **⚡ Hot path = zero allocations** — Reusable buffers, pre-formatted strings
3. **🔒 Cross-platform** — Unix-specific syscalls behind build tags
4. **📊 Observable** — Progress bar, stats, state files
5. **🧪 Testable** — Each package has a clear, narrow API

---


## 💡 Examples

### 🎯 **1. Quick Recon of a Subnet**

```bash
ChPortScanner -t 192.168.1.0/24 --top-ports 100 --stats
```

**Output:**
```
╔══════════════════════════════════════════════════╗
║  ChPortScanner v1.0.0                            ║
║  Owner: Ch4120N                                  ║
╚══════════════════════════════════════════════════╝

[i] Starting scan: 1000 workers | rate=unlimited | timeout=1s | jobs=25600
[####################] 100.00% | done:25600 open:12 | 12800 pps

192.168.1.1:80 (http)
192.168.1.1:443 (https)
192.168.1.10:22 (ssh)
192.168.1.10:80 (http)
...

[+] Scan completed: 12 open port(s)
```

### 🔍 **2. Deep Scan with All Probes**

```bash
ChPortScanner \
  -t 10.0.0.0/24 \
  -p 1-65535 \
  --turbo \
  --banner \
  --http-probe \
  --tls-cert \
  --rdns \
  --format jsonl \
  -o recon.jsonl
```

**Sample output (`recon.jsonl`):**
```json
{"ip":"10.0.0.5","port":22,"proto":"tcp","state":"open","service":"ssh","banner":"SSH-2.0-OpenSSH_8.9p1"}
{"ip":"10.0.0.5","port":443,"proto":"tcp","state":"open","service":"https","tls":"CN=example.com issuer=R3 expiry=2025-03-15","hostname":"web01.lab.local"}
```

### 🥷 **3. Stealth Scan**

```bash
ChPortScanner -t 172.16.0.0/24 --profile stealth --random -q
```

### 🐳 **4. Docker / CI Pipeline**

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /ChPortScanner .

FROM alpine:latest
COPY --from=builder /ChPortScanner /usr/local/bin/ChPortScanner
ENTRYPOINT ["ChPortScanner"]
```

```bash
docker run --rm --network host ChPortScanner -t 10.0.0.0/24 -p 22,80,443
```

### 📂 **5. Extract Results**

```bash
# From a saved text output, extract unique IPs
ChPortScanner --extract scan.txt

# Keep port information
ChPortScanner --extract scan.txt --extract-ports > unique_targets.txt
```

### 🔄 **6. Resume an Interrupted Scan**

```bash
# Start a long scan (auto-saves state)
ChPortScanner -t 10.0.0.0/8 -p 1-65535 --state bigscan.state

# Press Ctrl+C anytime...

# Resume later
ChPortScanner --resume bigscan.state
```

### 🌐 **7. IPv6 Scanning**

```bash
ChPortScanner -t 2001:db8::/64 -p 22,80,443 --top-ports 100
```

---

## 🛠️ Building from Source

### 📋 **Requirements**

- **Go** 1.21 or later
- ~50 MB disk space
- No CGO required (pure Go)

### 🔨 **Build Commands**

```bash
# Development build
go build -o ChPortScanner .

# Optimized release build
go build -ldflags="-s -w" -trimpath -o ChPortScanner

# Static Linux binary
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags="-s -w" -trimpath -o ChPortScanner-linux-amd64

# All platforms
for os in linux darwin windows; do
  for arch in amd64 arm64; do
    CGO_ENABLED=0 GOOS=$os GOARCH=$arch \
      go build -ldflags="-s -w" -trimpath \
      -o dist/ChPortScanner-$os-$arch .
  done
done
```

### 🧪 **Run Tests**

```bash
go test ./... -v
go test ./... -race -coverprofile=coverage.out
```

---

<!-- ═══════════════════════════════════════════════════════════════ -->
<!--                        FAQ                                     -->
<!-- ═══════════════════════════════════════════════════════════════ -->

## ❓ FAQ

<details>
<summary><b>🔒 Do I need root / administrator privileges?</b></summary>

**No.** ChPortScanner uses **TCP connect scanning**, which works with standard user permissions. It runs in unprivileged containers, CI pipelines, and restricted lab environments.
</details>

<details>
<summary><b>⚡ How fast is it compared to masscan?</b></summary>

masscan can hit **millions** of packets per second using **raw sockets**. ChPortScanner targets **10k–400k pps** depending on hardware and network — fast enough for **most lab environments** while **remaining policy-compliant** and **needing no root**.
</details>

<details>
<summary><b>🛡️ Will it trigger IDS / firewall alerts?</b></summary>

Since ChPortScanner uses **standard TCP connections**, traffic looks like normal client behavior — far less likely to trip SYN-flood signatures than raw-packet scanners. Still, **only scan networks you own or have permission to test**.
</details>

<details>
<summary><b>🐌 Why is my scan slow?</b></summary>

Common causes:

1. **`ulimit -n` too low** — run `ulimit -n 65535`
2. **`--rate` set too low** — try `--rate 0`
3. **Timeout too high** — try `--timeout 1`
4. **Network latency** — LAN is fast; WAN is not
5. **Low-end CPU** — try `--workers 500`
</details>

<details>
<summary><b>🪟 Does it work on Windows?</b></summary>

**Yes.** Windows builds are fully supported. Some Unix-specific optimizations (`SO_LINGER=0`, `ulimit` auto-raise) are Unix-only, but scanning works correctly on all platforms.
</details>

<details>
<summary><b>📁 Where is the state file saved?</b></summary>

By default: `./scan.state`. Override with `--state <path>`. Use `--resume <file>` to continue an interrupted scan.
</details>

<details>
<summary><b>🔍 What's the difference between `--lite` and `--turbo`?</b></summary>

- **`--lite`** — Low-resource mode (200 workers, small buffer). Best for 2-core / <4 GB RAM systems.
- **`--turbo`** — Max speed (2× workers, all optimizations). Best for modern desktops / servers on LAN.
</details>

<details>
<summary><b>⚠️ Legal warning</b></summary>

**ChPortScanner is intended for authorized testing only.** Only scan networks and hosts you own or have explicit **written permission** to test. Unauthorized port scanning may be **illegal** in your jurisdiction.
</details>

---

---

## 🤝 Contributing

Contributions are **warmly welcome**! Whether you're fixing a typo or adding a feature:

1. 🍴 **Fork** the repo
2. 🌿 **Create** a branch (`git checkout -b feature/amazing-feature`)
3. ✏️ **Commit** your changes (`git commit -m 'Add amazing feature'`)
4. 📤 **Push** (`git push origin feature/amazing-feature`)
5. 🎉 **Open a Pull Request**

### 📜 **Contribution Guidelines**

- Follow [Go idioms](https://go.dev/doc/effective_go)
- Run `go test ./...` before submitting
- Add tests for new features
- Update documentation
- Keep commits clean and focused

---

## 📜 License

**MIT License** — see [LICENSE](LICENSE) for full text.

---

## 👤 Author

<div align="center">

### **Ch4120N**

*Security Researcher • Tool Builder • Open Source Enthusiast*

[![GitHub](https://img.shields.io/badge/GitHub-Ch4120N-181717?style=for-the-badge&logo=github)](https://github.com/Ch4120N)
[![Twitter](https://img.shields.io/badge/Twitter-@Ch4120N-1DA1F2?style=for-the-badge&logo=twitter)](https://twitter.com/Ch4120N)

</div>

---

<div align="center">

### ⭐ **If you find this project useful, please give it a star!** ⭐

*It helps others discover the tool and motivates continued development.*

<br/>

[![Stargazers](https://img.shields.io/github/stars/Ch4120N/ChPortScanner?style=for-the-badge)](https://github.com/Ch4120N/ChPortScanner/stargazers)
[![Forks](https://img.shields.io/github/forks/Ch4120N/ChPortScanner?style=for-the-badge)](https://github.com/Ch4120N/ChPortScanner/fork)
[![Watchers](https://img.shields.io/github/watchers/Ch4120N/ChPortScanner?style=for-the-badge)](https://github.com/Ch4120N/ChPortScanner/watchers)

<br/>

Made with ❤️  and ☕ by Ch4120N
© 2026 — ChPortScanner v1.0.0


**[⬆ Back to Top](#-ChPortScanner-v100)**

</div>

<!--
  Tags: port scanner, golang port scanner, fast port scanner, network scanner,
  tcp scanner, udp scanner, cidr scanner, masscan alternative, nmap alternative,
  no root port scanner, unprivileged port scanner, go security tools,
  penetration testing tools, network recon, port scanning tool,
  cross-platform port scanner, lightweight port scanner, modular port scanner,
  ChPortScanner, Ch4120N, cybersecurity tools, lab tools, CTF tools,
  containerized port scanner, docker port scanner, kubernetes scanner,
  security research tools, ethical hacking tools, recon tools
-->