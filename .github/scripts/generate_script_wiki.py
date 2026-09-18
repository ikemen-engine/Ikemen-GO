#!/usr/bin/env python3
"""
Update the Lua wiki page from Lua API docblocks in src/script.go.

Patches the existing wiki page in place:
- Embedded functions from systemScriptInit()
- Trigger Redirection names from triggerRedirection()
- Trigger Functions names from triggerFunctions()

Function signatures are synthesised from the `@function` name and the ordered
`@tparam` tags.

Example:
    python3 .github/scripts/generate_script_wiki.py \
      --source src/script.go \
      --wiki-page wiki/Lua.md
"""

from __future__ import annotations

import argparse
import pathlib
import re
import sys
from dataclasses import dataclass, field
from typing import List, Optional, Tuple


EMBEDDED_START = "<!-- AUTOGEN:START embedded-functions -->"
EMBEDDED_END = "<!-- AUTOGEN:END embedded-functions -->"
TRIGGER_REDIRECTION_START = "<!-- AUTOGEN:START trigger-redirection -->"
TRIGGER_REDIRECTION_END = "<!-- AUTOGEN:END trigger-redirection -->"
TRIGGER_FUNCTIONS_START = "<!-- AUTOGEN:END trigger-functions -->".replace("END", "START")
TRIGGER_FUNCTIONS_END = "<!-- AUTOGEN:END trigger-functions -->"


@dataclass
class Tag:
    tag: str
    modifier: Optional[str]
    raw: str


@dataclass
class ParamDoc:
    type_name: str
    name: str
    description: str
    modifier: Optional[str] = None


@dataclass
class ReturnDoc:
    type_name: str
    name: str
    description: str


@dataclass
class RegisteredFunction:
    name: str
    docblock: Optional[str] = None
    summary: str = ""
    signature: Optional[str] = None
    params: List[ParamDoc] = field(default_factory=list)
    returns: List[ReturnDoc] = field(default_factory=list)


class DelimiterError(ValueError):
    def __init__(self, start_index: int, open_char: str, close_char: str, depth: int, state: str) -> None:
        self.start_index = start_index
        self.open_char = open_char
        self.close_char = close_char
        self.depth = depth
        self.state = state
        super().__init__(f"Unbalanced delimiters starting at index {start_index}")


REGISTER_NAME_RE = re.compile(r'luaRegister\s*\(\s*l\s*,\s*"([^"]+)"\s*,', re.S)
TAG_RE = re.compile(r'^@(?P<tag>[A-Za-z_][A-Za-z0-9_]*)(?:\[(?P<modifier>[^\]]+)\])?\s+(?P<rest>.*)$')


def main() -> int:
    parser = argparse.ArgumentParser(description="Update the Lua wiki page from src/script.go Lua docs.")
    parser.add_argument("--source", required=True, help="Path to src/script.go")
    parser.add_argument("--wiki-page", required=True, help="Path to the existing wiki page markdown file")
    args = parser.parse_args()

    source_path = pathlib.Path(args.source)
    wiki_page_path = pathlib.Path(args.wiki_page)

    source_text = source_path.read_text(encoding="utf-8")
    wiki_text = wiki_page_path.read_text(encoding="utf-8")

    system_body, system_body_offset = extract_function_body(source_text, "systemScriptInit")
    trigger_redirection_body, trigger_redirection_body_offset = extract_function_body(source_text, "triggerRedirection")
    trigger_functions_body, trigger_functions_body_offset = extract_function_body(source_text, "triggerFunctions")

    system_entries = [
        parse_registered_function(call)
        for call in extract_lua_register_calls(
            system_body,
            scope_name="systemScriptInit",
            source_text=source_text,
            base_offset=system_body_offset,
        )
    ]
    documented_entries = [entry for entry in system_entries if entry.docblock]
    undocumented_system_entries = [entry.name for entry in system_entries if not entry.docblock]

    trigger_redirection_entries = [
        parse_registered_function(call).name
        for call in extract_lua_register_calls(
            trigger_redirection_body,
            scope_name="triggerRedirection",
            source_text=source_text,
            base_offset=trigger_redirection_body_offset,
        )
    ]
    trigger_function_entries = [
        parse_registered_function(call).name
        for call in extract_lua_register_calls(
            trigger_functions_body,
            scope_name="triggerFunctions",
            source_text=source_text,
            base_offset=trigger_functions_body_offset,
        )
    ]

    embedded_fragment = render_embedded_fragment(documented_entries, undocumented_system_entries)
    trigger_redirection_fragment = render_name_list(trigger_redirection_entries)
    trigger_functions_fragment = render_name_list(trigger_function_entries)

    updated_wiki_text = patch_wiki_page(
        wiki_text,
        embedded_fragment=embedded_fragment,
        trigger_redirection_fragment=trigger_redirection_fragment,
        trigger_functions_fragment=trigger_functions_fragment,
    )

    wiki_page_path.write_text(updated_wiki_text, encoding="utf-8")
    return 0


def extract_function_body(text: str, func_name: str) -> Tuple[str, int]:
    match = re.search(rf'func\s+{re.escape(func_name)}\s*\([^)]*\)\s*\{{', text)
    if not match:
        raise ValueError(f"Could not find function {func_name}")

    brace_start = match.end() - 1
    try:
        brace_end = find_matching_delimiter(text, brace_start, "{", "}")
    except DelimiterError as exc:
        line, col = index_to_line_col(text, brace_start)
        excerpt = make_line_excerpt(text, brace_start)
        raise ValueError(
            f'Could not find closing "}}" for function `{func_name}` '
            f'(opened at line {line}, column {col}). '
            f"Scanner ended in state={exc.state!r}, depth={exc.depth}.\n\n{excerpt}"
        ) from exc

    return text[brace_start + 1 : brace_end], brace_start + 1


def extract_lua_register_calls(
    body: str,
    *,
    scope_name: str,
    source_text: Optional[str] = None,
    base_offset: int = 0,
) -> List[str]:
    calls: List[str] = []
    idx = 0
    source_text = source_text if source_text is not None else body

    while True:
        start = find_token_in_code(body, "luaRegister(", idx)
        if start == -1:
            break

        open_paren = body.find("(", start)
        try:
            end = find_matching_delimiter(body, open_paren, "(", ")")
        except DelimiterError as exc:
            global_start = base_offset + start
            global_open = base_offset + open_paren
            call_line, call_col = index_to_line_col(source_text, global_start)
            open_line, open_col = index_to_line_col(source_text, global_open)
            name = detect_partial_register_name(body[start : start + 200])
            excerpt = make_line_excerpt(source_text, global_start)

            raise ValueError(
                f'While parsing `{scope_name}()`, could not find closing `)` for '
                f'`luaRegister("{name}")` starting at line {call_line}, column {call_col} '
                f'(opening `(` at line {open_line}, column {open_col}). '
                f"Scanner ended in state={exc.state!r}, depth={exc.depth}.\n\n{excerpt}"
            ) from exc

        calls.append(body[start : end + 1])
        idx = end + 1

    return calls


def find_token_in_code(text: str, token: str, start_index: int = 0) -> int:
    i = start_index
    state = "code"

    while i < len(text):
        ch = text[i]
        nxt = text[i + 1] if i + 1 < len(text) else ""

        if state == "code":
            if text.startswith(token, i):
                return i
            if ch == '"':
                state = "double"
            elif ch == "'":
                state = "single"
            elif ch == "`":
                state = "raw"
            elif ch == "/" and nxt == "/":
                state = "line_comment"
                i += 1
            elif ch == "/" and nxt == "*":
                state = "block_comment"
                i += 1

        elif state == "double":
            if ch == "\\":
                i += 1
            elif ch == '"':
                state = "code"

        elif state == "single":
            if ch == "\\":
                i += 1
            elif ch == "'":
                state = "code"

        elif state == "raw":
            if ch == "`":
                state = "code"

        elif state == "line_comment":
            if ch == "\n":
                state = "code"

        elif state == "block_comment":
            if ch == "*" and nxt == "/":
                state = "code"
                i += 1

        i += 1

    return -1


def find_matching_delimiter(text: str, start_index: int, open_char: str, close_char: str) -> int:
    depth = 0
    i = start_index
    state = "code"

    while i < len(text):
        ch = text[i]
        nxt = text[i + 1] if i + 1 < len(text) else ""

        if state == "code":
            if ch == '"':
                state = "double"
            elif ch == "'":
                state = "single"
            elif ch == "`":
                state = "raw"
            elif ch == "/" and nxt == "/":
                state = "line_comment"
                i += 1
            elif ch == "/" and nxt == "*":
                state = "block_comment"
                i += 1
            elif ch == open_char:
                depth += 1
            elif ch == close_char:
                depth -= 1
                if depth == 0:
                    return i

        elif state == "double":
            if ch == "\\":
                i += 1
            elif ch == '"':
                state = "code"

        elif state == "single":
            if ch == "\\":
                i += 1
            elif ch == "'":
                state = "code"

        elif state == "raw":
            if ch == "`":
                state = "code"

        elif state == "line_comment":
            if ch == "\n":
                state = "code"

        elif state == "block_comment":
            if ch == "*" and nxt == "/":
                state = "code"
                i += 1

        i += 1

    raise DelimiterError(
        start_index=start_index,
        open_char=open_char,
        close_char=close_char,
        depth=depth,
        state=state,
    )


def index_to_line_col(text: str, index: int) -> Tuple[int, int]:
    line = text.count("\n", 0, index) + 1
    line_start = text.rfind("\n", 0, index)
    col = index + 1 if line_start == -1 else index - line_start
    return line, col


def make_line_excerpt(text: str, index: int, context_lines: int = 3) -> str:
    line_no, _ = index_to_line_col(text, index)
    lines = text.splitlines()
    start = max(1, line_no - context_lines)
    end = min(len(lines), line_no + context_lines)

    out = []
    for n in range(start, end + 1):
        prefix = ">" if n == line_no else " "
        out.append(f"{prefix}{n:6d}: {lines[n - 1]}")
    return "\n".join(out)


def detect_partial_register_name(text: str) -> str:
    match = re.search(r'luaRegister\s*\(\s*l\s*,\s*"([^"\n]*)', text, re.S)
    if match:
        return match.group(1)
    return "<unknown>"


def parse_registered_function(call_text: str) -> RegisteredFunction:
    name_match = REGISTER_NAME_RE.search(call_text)
    if not name_match:
        raise ValueError(f"Could not parse luaRegister name from:\n{call_text[:200]}")
    name = name_match.group(1)

    docblock_match = re.search(r'/\*(.*?)\*/', call_text, re.S)
    entry = RegisteredFunction(name=name)
    if not docblock_match:
        return entry

    raw_docblock = docblock_match.group(1)
    entry.docblock = raw_docblock
    summary, func_name, params, returns = parse_docblock(raw_docblock)
    entry.summary = summary
    entry.params = params
    entry.returns = returns
    entry.signature = build_signature(func_name or name, params)
    return entry


def parse_docblock(docblock: str) -> Tuple[str, Optional[str], List[ParamDoc], List[ReturnDoc]]:
    lines = [line.rstrip() for line in docblock.splitlines()]
    cleaned_lines = [line.strip() for line in lines]

    summary_lines: List[str] = []
    tag_lines: List[str] = []
    func_name: Optional[str] = None
    current_tag: Optional[str] = None

    for raw_line, line in zip(lines, cleaned_lines):
        _ = raw_line
        if not line:
            if current_tag is None:
                if summary_lines and summary_lines[-1] != "":
                    summary_lines.append("")
            else:
                current_tag += "\n"
            continue

        if line.startswith("@"):
            if current_tag is not None:
                tag_lines.append(current_tag)
            current_tag = line
            continue

        if current_tag is not None:
            current_tag += "\n" + line
        else:
            summary_lines.append(line)

    if current_tag is not None:
        tag_lines.append(current_tag)

    summary = normalize_doc_text("\n".join(summary_lines).strip())

    parsed_tags: List[Tag] = []
    for tag_line in tag_lines:
        first, *continuation = tag_line.splitlines()
        tag_match = TAG_RE.match(first)
        if not tag_match:
            continue
        raw = tag_match.group("rest")
        if continuation:
            raw += "\n" + "\n".join(continuation)
        parsed_tags.append(
            Tag(
                tag=tag_match.group("tag"),
                modifier=tag_match.group("modifier"),
                raw=normalize_doc_text(raw),
            )
        )

    params: List[ParamDoc] = []
    returns: List[ReturnDoc] = []

    for tag in parsed_tags:
        if tag.tag == "tparam":
            type_name, name, description = split_param_doc(tag.raw)
            params.append(
                ParamDoc(
                    type_name=type_name,
                    name=name,
                    description=description,
                    modifier=tag.modifier,
                )
            )
        elif tag.tag == "treturn":
            type_name, name, description = split_return_doc(tag.raw)
            returns.append(
                ReturnDoc(
                    type_name=type_name,
                    name=name,
                    description=description,
                )
            )
        elif tag.tag == "function" and func_name is None:
            func_name = normalize_inline_text(tag.raw)

    return summary, func_name, params, returns


def build_signature(func_name: Optional[str], params: List[ParamDoc]) -> Optional[str]:
    """Synthesise a Lua usage line from the @function name and @tparam order."""
    if not func_name:
        return None
    names = [param.name for param in params if param.name]
    return f"{func_name}({', '.join(names)})"


def split_param_doc(raw: str) -> Tuple[str, str, str]:
    first_line, rest = split_first_line(raw)
    parts = first_line.split(None, 2)
    if len(parts) == 1:
        type_name, name, description = parts[0], "", ""
    elif len(parts) == 2:
        type_name, name, description = parts[0], parts[1], ""
    else:
        type_name, name, description = parts[0], parts[1], parts[2]

    if rest:
        description = join_description(description, rest)
    return type_name, name, description


def split_return_doc(raw: str) -> Tuple[str, str, str]:
    first_line, rest = split_first_line(raw)
    parts = first_line.split(None, 2)

    if len(parts) == 1:
        type_name, name, description = parts[0], "", ""
    elif len(parts) == 2:
        type_name, name, description = parts[0], "", parts[1]
    else:
        type_name, second, tail = parts
        if re.match(r"^[a-z_][A-Za-z0-9_]*$", second):
            name, description = second, tail
        else:
            name, description = "", f"{second} {tail}"

    if rest:
        description = join_description(description, rest)
    return type_name, name, description


def split_first_line(text: str) -> Tuple[str, str]:
    lines = text.splitlines()
    if not lines:
        return "", ""
    first = lines[0].strip()
    rest = "\n".join(line.rstrip() for line in lines[1:]).strip()
    return first, rest


def join_description(first: str, rest: str) -> str:
    first = first.strip()
    rest = rest.strip()
    if first and rest:
        return f"{first}\n{rest}"
    return first or rest


def normalize_inline_text(text: str) -> str:
    return re.sub(r"\s+", " ", text).strip()


def normalize_doc_text(text: str) -> str:
    lines = [line.strip() for line in text.splitlines()]
    out: List[str] = []
    paragraph_parts: List[str] = []

    def flush_paragraph() -> None:
        nonlocal paragraph_parts
        if paragraph_parts:
            out.append(" ".join(paragraph_parts).strip())
            paragraph_parts = []

    for line in lines:
        if not line:
            flush_paragraph()
            if out and out[-1] != "":
                out.append("")
            continue

        if is_list_line(line):
            flush_paragraph()
            out.append(line)
            continue

        paragraph_parts.append(line)

    flush_paragraph()

    while out and out[0] == "":
        out.pop(0)
    while out and out[-1] == "":
        out.pop()

    return "\n".join(out)


def is_list_line(line: str) -> bool:
    stripped = line.lstrip()
    return stripped.startswith("- ") or stripped.startswith("* ")


def md_escape(value: str) -> str:
    return value.replace("|", r"\|")


def format_doc_text(value: str) -> str:
    escaped_lines = [md_escape(line) for line in value.splitlines()]
    return "<br>".join(escaped_lines)


def format_param_name(param: ParamDoc) -> str:
    name = f"`{md_escape(param.name)}`" if param.name else ""
    if param.modifier and param.modifier.startswith("opt"):
        name += "*"
    return name


def format_param_description(param: ParamDoc) -> str:
    desc = param.description.strip()
    if not param.modifier:
        return desc

    if param.modifier == "opt":
        prefix = "Optional."
    elif param.modifier.startswith("opt="):
        default_value = param.modifier.split("=", 1)[1]
        prefix = f"Optional. Default: `{md_escape(default_value)}`."
    else:
        prefix = f"`{md_escape(param.modifier)}`."

    if not desc:
        return prefix
    return f"{prefix}\n{desc}" if "\n" in desc else f"{prefix} {desc}"


def format_return_name(ret: ReturnDoc) -> str:
    return f"`{md_escape(ret.name)}`" if ret.name else ""


def render_embedded_fragment(
    documented_entries: List[RegisteredFunction],
    undocumented_system_entries: List[str],
) -> str:
    out: List[str] = []

    if not documented_entries:
        out.append("No machine-parsable docblocks were found.")
    else:
        out.append("Below are the documented embedded Lua functions exposed by the engine.")
        out.append("")
        out.append("`*` marks an optional parameter.")
        out.append("")

        for entry in documented_entries:
            out.append(f"### `{entry.name}`")
            out.append("")
            if entry.signature:
                out.append("```lua")
                out.append(entry.signature)
                out.append("```")
                out.append("")
            if entry.summary:
                out.append(entry.summary)
                out.append("")

            if entry.params:
                out.append("**Parameters**")
                out.append("")
                out.append("| # | Name | Type | Description |")
                out.append("| --- | --- | --- | --- |")
                for idx, param in enumerate(entry.params, start=1):
                    out.append(
                        f"| {idx} | {format_param_name(param)} | `{md_escape(param.type_name)}` | {format_doc_text(format_param_description(param))} |"
                    )
                out.append("")

            if entry.returns:
                out.append("**Returns**")
                out.append("")
                out.append("| Name | Type | Description |")
                out.append("| --- | --- | --- |")
                for ret in entry.returns:
                    out.append(
                        f"| {format_return_name(ret)} | `{md_escape(ret.type_name)}` | {format_doc_text(ret.description)} |"
                    )
                out.append("")

    if undocumented_system_entries:
        out.append("### Undocumented embedded functions")
        out.append("")
        for name in undocumented_system_entries:
            out.append(f"- `{name}`")
        out.append("")

    return "\n".join(out).rstrip()


def render_name_list(names: List[str]) -> str:
    if not names:
        return "None."
    return "\n".join(f"- `{name}`" for name in names)


def patch_wiki_page(
    wiki_text: str,
    *,
    embedded_fragment: str,
    trigger_redirection_fragment: str,
    trigger_functions_fragment: str,
) -> str:
    wiki_text = patch_embedded_section(wiki_text, embedded_fragment)
    wiki_text = patch_trigger_redirection_section(wiki_text, trigger_redirection_fragment)
    wiki_text = patch_trigger_functions_section(wiki_text, trigger_functions_fragment)
    return wiki_text


def patch_embedded_section(text: str, fragment: str) -> str:
    marker_block = render_marker_block(EMBEDDED_START, EMBEDDED_END, fragment)
    if EMBEDDED_START in text and EMBEDDED_END in text:
        return replace_marker_block(text, EMBEDDED_START, EMBEDDED_END, fragment)

    section = "\n".join(
        [
            '## <a name="functions_embedded">Embedded functions</a>',
            "",
            marker_block,
            "",
        ]
    )

    pattern = re.compile(
        r'## <a name="functions_embedded">Embedded functions</a>\r?\n.*?(?=\r?\n## <a name="functions_triggers">Lua "triggers"</a>\r?\n)',
        re.S,
    )
    updated, count = pattern.subn(section.rstrip(), text, count=1)
    if count != 1:
        raise ValueError("Could not replace the `Embedded functions` section in the wiki page.")
    return updated


def patch_trigger_redirection_section(text: str, fragment: str) -> str:
    marker_block = render_marker_block(TRIGGER_REDIRECTION_START, TRIGGER_REDIRECTION_END, fragment)
    if TRIGGER_REDIRECTION_START in text and TRIGGER_REDIRECTION_END in text:
        return replace_marker_block(text, TRIGGER_REDIRECTION_START, TRIGGER_REDIRECTION_END, fragment)

    section = "\n".join(
        [
            "**Trigger Redirection**",
            "",
            "Redirection returns *true* if it successfully finds *n-th* player, or *false* otherwise. "
            'Lua "trigger" functions code used after redirection will be executed via matched player/helper, '
            "as long as new redirection is not used.",
            "",
            marker_block,
            "",
        ]
    )

    pattern = re.compile(
        r"\*\*Trigger Redirection\*\*\r?\n.*?(?=\r?\n\*\*Trigger Functions\*\*\r?\n)",
        re.S,
    )
    updated, count = pattern.subn(section.rstrip(), text, count=1)
    if count != 1:
        raise ValueError("Could not replace the `Trigger Redirection` section in the wiki page.")
    return updated


def patch_trigger_functions_section(text: str, fragment: str) -> str:
    marker_block = render_marker_block(TRIGGER_FUNCTIONS_START, TRIGGER_FUNCTIONS_END, fragment)
    if TRIGGER_FUNCTIONS_START in text and TRIGGER_FUNCTIONS_END in text:
        return replace_marker_block(text, TRIGGER_FUNCTIONS_START, TRIGGER_FUNCTIONS_END, fragment)

    section = "\n".join(
        [
            "**Trigger Functions**",
            "",
            marker_block,
            "",
        ]
    )

    pattern = re.compile(
        r"\*\*Trigger Functions\*\*\r?\n.*?(?=\r?\n# <a name=\"lua_troubleshooting\">Troubleshooting scripts</a>\r?\n)",
        re.S,
    )
    updated, count = pattern.subn(section.rstrip(), text, count=1)
    if count != 1:
        raise ValueError("Could not replace the `Trigger Functions` section in the wiki page.")
    return updated


def render_marker_block(start_marker: str, end_marker: str, fragment: str) -> str:
    return f"{start_marker}\n{fragment.rstrip()}\n{end_marker}"


def replace_marker_block(text: str, start_marker: str, end_marker: str, fragment: str) -> str:
    start_index = text.find(start_marker)
    if start_index == -1:
        raise ValueError(f"Missing marker: {start_marker}")
    end_index = text.find(end_marker, start_index)
    if end_index == -1:
        raise ValueError(f"Missing marker: {end_marker}")
    end_index += len(end_marker)
    return text[:start_index] + render_marker_block(start_marker, end_marker, fragment) + text[end_index:]


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except Exception as exc:  # pragma: no cover - CLI error path
        print(f"ERROR: {exc}", file=sys.stderr)
        raise
