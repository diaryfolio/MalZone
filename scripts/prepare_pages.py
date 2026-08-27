#!/usr/bin/env python3
"""Prepare an allow-listed, static Jekyll source tree for GitHub Pages."""

from __future__ import annotations

import argparse
import re
import shutil
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[1]
ALLOWED_ROOT_FILES = ("README.md", "CLAUDE.md", "AGENTS.md", "CONTRIBUTING.md")
ALLOWED_TREES = (
    Path("docs/design"),
    Path("docs/development"),
    Path("docs/prompts/governance"),
    Path("contracts"),
    Path("examples/windows"),
)
ALLOWED_SUFFIXES = {".md", ".json"}
SECRET_PATTERNS = (
    re.compile(r"-----BEGIN (?:RSA |EC |OPENSSH )?PRIVATE KEY-----"),
    re.compile(r"\bAKIA[0-9A-Z]{16}\b"),
    re.compile(r"\bgh[pousr]_[A-Za-z0-9]{30,}\b"),
    re.compile(r"\bxox[baprs]-[A-Za-z0-9-]{20,}\b"),
)
MARKDOWN_LINK = re.compile(r"(\[[^]]+\]\()([^)#?]+)\.md((?:#[^)]*)?\))")
LOCAL_LINK = re.compile(r"(\[[^]]+\]\()([^)]+)(\))")
GITHUB_BLOB = "https://github.com/diaryfolio/MalZone/blob/main/"


def page_front_matter(text: str, source: Path) -> str:
    title_match = re.search(r"^#\s+(.+)$", text, re.MULTILINE)
    title = title_match.group(1).strip() if title_match else source.stem.replace("-", " ").title()
    title = title.replace("\\", "\\\\").replace('"', '\\"')
    if text.startswith("---\n"):
        closing = text.find("\n---\n", 4)
        if closing < 0:
            raise ValueError(f"unterminated front matter: {source}")
        header = text[4:closing]
        additions = []
        if not re.search(r"^layout:\s*", header, re.MULTILINE):
            additions.append("layout: default")
        if not re.search(r"^title:\s*", header, re.MULTILINE):
            additions.append(f'title: "{title}"')
        if additions:
            header = header.rstrip() + "\n" + "\n".join(additions)
        return f"---\n{header}\n---\n{text[closing + 5:]}"
    return f'---\nlayout: default\ntitle: "{title}"\n---\n\n{text}'


def safe_text(source: Path) -> str:
    text = source.read_text(encoding="utf-8")
    for pattern in SECRET_PATTERNS:
        if pattern.search(text):
            raise ValueError(f"possible secret in publishable file: {source}")
    return text


def is_publishable(relative: Path) -> bool:
    if relative.as_posix() in ALLOWED_ROOT_FILES or relative == Path("docs/index.md"):
        return True
    return any(
        relative.is_relative_to(tree) and relative.suffix in ALLOWED_SUFFIXES
        for tree in ALLOWED_TREES
    )


def externalize_nonpublished_links(text: str, source: Path) -> str:
    def replace(match: re.Match[str]) -> str:
        target = match.group(2)
        if "://" in target or target.startswith(("#", "mailto:")):
            return match.group(0)
        path, separator, fragment = target.partition("#")
        resolved = (source.parent / path).resolve()
        try:
            relative = resolved.relative_to(REPO_ROOT)
        except ValueError:
            return match.group(0)
        if is_publishable(relative):
            return match.group(0)
        suffix = f"#{fragment}" if separator else ""
        return f"{match.group(1)}{GITHUB_BLOB}{relative.as_posix()}{suffix}{match.group(3)}"

    return LOCAL_LINK.sub(replace, text)


def copy_publishable(source: Path, destination: Path, *, relocate_docs_index: bool = False) -> None:
    if source.suffix not in ALLOWED_SUFFIXES:
        raise ValueError(f"unsupported publishable suffix: {source}")
    text = safe_text(source)
    text = externalize_nonpublished_links(text, source)
    if relocate_docs_index:
        replacements = {
            "](../README.md)": "](README.md)",
            "](design/": "](docs/design/",
            "](../contracts/": "](contracts/",
            "](development/": "](docs/development/",
            "](../CLAUDE.md)": "](CLAUDE.md)",
            "](../AGENTS.md)": "](AGENTS.md)",
            "](../CONTRIBUTING.md)": "](CONTRIBUTING.md)",
            "](prompts/": "](docs/prompts/",
        }
        for old, new in replacements.items():
            text = text.replace(old, new)
    destination.parent.mkdir(parents=True, exist_ok=True)
    if source.suffix == ".md":
        text = MARKDOWN_LINK.sub(r"\1\2.html\3", page_front_matter(text, source))
    destination.write_text(text, encoding="utf-8")


def prepare(output: Path) -> None:
    resolved = output.resolve()
    if resolved == REPO_ROOT or REPO_ROOT in resolved.parents and resolved.name == ".git":
        raise ValueError("refusing unsafe Pages output path")
    if output.exists():
        shutil.rmtree(output)
    output.mkdir(parents=True)

    for relative in ALLOWED_ROOT_FILES:
        copy_publishable(REPO_ROOT / relative, output / relative)
    for tree in ALLOWED_TREES:
        for source in sorted((REPO_ROOT / tree).rglob("*")):
            if source.is_file() and source.suffix in ALLOWED_SUFFIXES:
                copy_publishable(source, output / source.relative_to(REPO_ROOT))

    copy_publishable(REPO_ROOT / "docs/index.md", output / "docs/index.md")
    copy_publishable(
        REPO_ROOT / "docs/index.md", output / "index.md", relocate_docs_index=True
    )
    shutil.copy2(REPO_ROOT / "docs/_config.yml", output / "_config.yml")
    (output / "_layouts").mkdir()
    (output / "assets").mkdir()
    shutil.copy2(REPO_ROOT / "docs/site/default.html", output / "_layouts/default.html")
    shutil.copy2(REPO_ROOT / "docs/site/site.css", output / "assets/site.css")
    shutil.copy2(REPO_ROOT / "docs/site/mermaid.js", output / "assets/mermaid.js")


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--output", type=Path, default=REPO_ROOT / "build/pages-src")
    args = parser.parse_args()
    prepare(args.output)


if __name__ == "__main__":
    main()
