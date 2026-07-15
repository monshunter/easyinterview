#!/usr/bin/env python3
"""Redact the scenario's synthetic email from a text stream."""

from __future__ import annotations

import sys
from urllib.parse import quote


REDACTION = "<redacted-synthetic-email>"


def main() -> None:
    if len(sys.argv) != 2 or not sys.argv[1]:
        raise SystemExit("usage: redact_stream.py EMAIL")

    email = sys.argv[1]
    sensitive_values = (email, quote(email, safe=""))
    for line in sys.stdin:
        for value in sensitive_values:
            line = line.replace(value, REDACTION)
        sys.stdout.write(line)


if __name__ == "__main__":
    main()
