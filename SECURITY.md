# Security Policy

## Scope

hanko is a local, read-only CLI tool. It reads plugin manifest files (`plugin.json`, `marketplace.json`) from paths you hand it, validates them against an embedded JSON Schema, and prints a report. It does not open network sockets, does not phone home, does not collect telemetry, does not require elevated privileges, and does not modify any file on your disk.

A security issue is anything that violates the above: network egress, privilege escalation, arbitrary code execution during validation, or anything that leaks data off the machine.

## Reporting a vulnerability

**Do not file a public issue.**

Send the details to **security@roninforge.org** with:

- A description of the issue and its impact
- Steps to reproduce (ideally a minimal manifest that triggers it)
- Affected versions
- Your name and whether you want credit in the advisory

You will get an acknowledgement within 72 hours. We aim to have a patched release available within 14 days for high-severity issues and 30 days for lower-severity ones. The embargo window is 90 days.

## Supported versions

Only the latest minor release on the `main` branch receives security fixes.

## No bug bounty

hanko is a small OSS project. We cannot pay bounties. We will credit responsible disclosures in release notes and the advisory.
