#!/usr/bin/env bash
# Emit mechanical facts about a repository README. Judgement stays with the caller.
# Usage: readme-audit.sh [path-to-readme-or-repo-root]

# Requires: bash 3.2+ (macOS default) or bash 4/5 (Linux), plus grep, sed, awk,
# wc, sort, uniq, tr, head, cut, dirname.
# Portable across macOS (BSD userland) and Linux (GNU coreutils): uses POSIX
# regex classes and no GNU-only flags.

set -uo pipefail

target="${1:-.}"
if [ -d "$target" ]; then
  root="$target"
  file=""
  for candidate in "$root/.github/README.md" "$root/README.md" "$root/docs/README.md"; do
    [ -f "$candidate" ] || continue
    if [ -z "$file" ]; then file="$candidate"; else echo "SHADOWED: $candidate is never rendered while $file exists"; fi
  done
  if [ -z "$file" ]; then echo "NO README FOUND under $root (.github/, root, docs/)"; exit 1; fi
else
  file="$target"
  root="$(dirname "$file")"
  [ -f "$file" ] || { echo "NOT A FILE: $file"; exit 1; }
fi

echo "FILE: $file"

bytes=$(wc -c < "$file" | tr -d ' ')
lines=$(wc -l < "$file" | tr -d ' ')
words=$(wc -w < "$file" | tr -d ' ')
echo "SIZE: ${bytes} bytes / ${lines} lines / ${words} words"
[ "$bytes" -gt 512000 ] && echo "WARNING: over 500 KiB, GitHub truncates the rendered view"
[ "$bytes" -lt 2048 ] && echo "WARNING: under 2 KiB — the size at which research samples treat a README as effectively empty"

# Strip fenced code blocks so headings inside examples are not counted as sections.
prose=$(awk '/^[[:space:]]*(```|~~~)/{f=!f; next} !f' "$file")

echo
echo "HEADINGS:"
printf '%s\n' "$prose" | grep -nE '^#{1,6} ' | sed 's/^/  /' || echo "  (none)"
section_count=$(printf '%s\n' "$prose" | grep -cE '^#{2,6} ')
echo "SECTION COUNT (h2-h6): $section_count   [reference sample: median 7, middle 50% between 5 and 12]"

echo
first_fence=$(grep -nE '^[[:space:]]*(```|~~~)' "$file" | head -1 | cut -d: -f1)
if [ -n "${first_fence:-}" ]; then
  pre_words=$(head -n "$((first_fence - 1))" "$file" | wc -w | tr -d ' ')
  echo "FIRST CODE BLOCK: line $first_fence, after $pre_words words of prose"
else
  echo "FIRST CODE BLOCK: none — the reader never gets a copyable command"
fi

echo
badges=$(head -40 "$file" | grep -oE '\[!\[[^]]*\]\([^)]*\)\]\([^)]*\)|!\[[^]]*\]\([^)]*(shields\.io|badge|badgen|codecov|coveralls|circleci|travis-ci|appveyor)[^)]*\)' | wc -l | tr -d ' ')
echo "BADGES IN FIRST 40 LINES: $badges"

noalt=$(grep -cE '!\[[[:space:]]*\]\(' "$file")
echo "IMAGES WITH EMPTY ALT TEXT: $noalt"

# POSIX ERE has no \b word-boundary token (a GNU-only extension), so the tag
# match instead captures one trailing non-word character (or end of line) to
# rule out a real word merely starting with a tag name (e.g. "<divider>" must
# not count as a "<div" hit). That trailing character is then stripped before
# counting, so grouping by tag name for "sort | uniq -c" is unaffected.
html_tags=$(grep -oE '<(picture|source|details|summary|img|div|table|br|h[1-6])([^[:alnum:]_]|$)' "$file" \
  | sed -E 's/[^[:alnum:]_]$//' \
  | sort | uniq -c | tr '\n' ' ')
echo "RAW HTML TAGS: ${html_tags:-none}   [registries and terminal readers may strip these]"

alerts=$(grep -cE '^>[[:space:]]*\[!(NOTE|TIP|IMPORTANT|WARNING|CAUTION)\]' "$file")
echo "GITHUB ALERTS: $alerts   [GitHub advises one or two per document]"

echo
echo "BROKEN RELATIVE LINKS:"
broken=0
while IFS= read -r link; do
  case "$link" in http*|\#*|mailto:*|"") continue ;; esac
  path="${link%%#*}"
  [ -z "$path" ] && continue
  case "$path" in /*) resolved="$root$path" ;; *) resolved="$root/$path" ;; esac
  if [ ! -e "$resolved" ]; then echo "  MISSING: $link"; broken=$((broken + 1)); fi
done < <(grep -oE '\]\([^)]+\)' "$file" | sed -E 's/^\]\(//; s/\)$//' | sed -E 's/[[:space:]]+".*"$//')
[ "$broken" -eq 0 ] && echo "  (none)"

echo
echo "CONTENT CATEGORIES PRESENT (heading heuristic):"
check() {
  if printf '%s\n' "$prose" | grep -qiE "^#{1,6} .*($2)"; then echo "  $1: yes"; else echo "  $1: MISSING"; fi
}
check "What (intro/overview)"        "intro|about|overview|what is|description|synopsis"
check "Why (differentiation)"        "why|motivation|compar|alternativ|advantage|when to use|trade-?off"
check "How (install/usage)"          "install|usage|getting started|quick ?start|setup|configur|requirement|example|run |build|api|cli|command"
check "When (status/roadmap)"        "status|roadmap|version|stability|maturity|release|caveat|limitation|known issue"
check "Who (licence/maintainers)"    "licen[cs]e|maintainer|author|credit|acknowledg|team|contact|code of conduct|sponsor"
check "References (docs/support)"    "document|docs|support|faq|resource|related|see also|link|learn more"
check "Contribution"                 "contribut|development setup|hacking|dev environment"
echo "  [in the reference sample: What 97%, How 88.5%, References 61%, Who 53%, Contribution 28%, Why 26%, When 21%]"

echo
placeholders=$(grep -niE 'TODO|FIXME|coming soon|lorem ipsum|<your |xxx+|placeholder|\[insert ' "$file" | head -10)
if [ -n "$placeholders" ]; then echo "PLACEHOLDERS LEFT IN FILE:"; printf '%s\n' "$placeholders" | sed 's/^/  /'; else echo "PLACEHOLDERS: none"; fi

echo
toc=$(printf '%s\n' "$prose" | grep -inE '^#{1,6} .*(table of contents|contents)$' | head -1)
[ -n "$toc" ] && echo "HAND-WRITTEN TOC at $toc — GitHub already generates an outline from headings; a manual one goes stale"

exit 0
