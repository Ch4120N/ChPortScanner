/*
 * ChPortScanner - High-Speed Port Scanner for Restricted Lab Environments
 * Owner: Ch4120N
 * Version: 1.0.0
 *
 * Build with version metadata injection:
 *   go build -trimpath \
 *     -ldflags "-s -w \
 *       -X main.buildVersion=1.0.0 \
 *       -X main.buildTag=v1.0.0 \
 *       -X main.buildCommit=$(git rev-parse --short HEAD) \
 *       -X main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
 *     -o chportscanner .
 */

package main

import (
