#!/usr/bin/env python3
"""Reject broken local links in the rendered GitHub Pages artifact."""

from __future__ import annotations

import argparse
from html.parser import HTMLParser
from pathlib import Path
from urllib.parse import unquote, urlparse


class LinkParser(HTMLParser):
    def __init__(self) -> None:
        super().__init__()
        self.links: list[str] = []

    def handle_starttag(self, tag: str, attrs: list[tuple[str, str | None]]) -> None:
        if tag not in {"a", "link", "script", "img"}:
            return
        attribute = "href" if tag in {"a", "link"} else "src"
        for name, value in attrs:
            if name == attribute and value:
                self.links.append(value)


def check(root: Path, baseurl: str) -> None:
    root = root.resolve()
    failures: list[str] = []
    for page in root.rglob("*.html"):
        parser = LinkParser()
        parser.feed(page.read_text(encoding="utf-8"))
        for link in parser.links:
            parsed = urlparse(link)
            if parsed.scheme or parsed.netloc or link.startswith(("#", "mailto:", "data:")):
                continue
            path = unquote(parsed.path)
            if path.startswith(baseurl + "/"):
                target = root / path[len(baseurl) + 1 :]
            elif path.startswith("/"):
                failures.append(f"{page.relative_to(root)}: unexpected absolute path {link}")
                continue
            else:
                target = page.parent / path
            if path.endswith("/"):
                target = target / "index.html"
            try:
                target.resolve().relative_to(root)
            except ValueError:
                failures.append(f"{page.relative_to(root)}: link escapes site {link}")
                continue
            if not target.exists():
                failures.append(f"{page.relative_to(root)}: missing {link}")
    if failures:
        raise ValueError("broken rendered Pages links:\n" + "\n".join(failures))


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--root", type=Path, required=True)
    parser.add_argument("--baseurl", default="/MalZone")
    args = parser.parse_args()
    check(args.root, args.baseurl)
    print("rendered Pages links passed")


if __name__ == "__main__":
    main()
