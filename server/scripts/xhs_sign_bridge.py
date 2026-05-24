#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""小红书 API 签名桥接。依赖: pip install xhshow"""
import json
import sys


def _configure_stdio() -> None:
    """Windows 默认控制台编码非 UTF-8，避免 print 触发 surrogate 错误。"""
    try:
        sys.stdout.reconfigure(encoding="utf-8", errors="replace")
        sys.stderr.reconfigure(encoding="utf-8", errors="replace")
    except Exception:
        pass


def _strip_surrogates(text: str) -> str:
    if not isinstance(text, str):
        return text
    return "".join(ch for ch in text if not (0xD800 <= ord(ch) <= 0xDFFF))


def _sanitize_obj(obj):
    if isinstance(obj, str):
        return _strip_surrogates(obj)
    if isinstance(obj, dict):
        return {_sanitize_obj(k): _sanitize_obj(v) for k, v in obj.items()}
    if isinstance(obj, list):
        return [_sanitize_obj(x) for x in obj]
    return obj


def _emit(obj) -> None:
    """始终向 stdout 写入合法 UTF-8 字节，供 Go 解析。"""
    raw = json.dumps(obj, ensure_ascii=False)
    sys.stdout.buffer.write(raw.encode("utf-8", errors="replace"))
    sys.stdout.buffer.write(b"\n")
    sys.stdout.buffer.flush()


def cookie_str_to_dict(raw: str) -> dict:
    out = {}
    for part in (raw or "").split(";"):
        part = part.strip()
        if not part or "=" not in part:
            continue
        k, v = part.split("=", 1)
        out[k.strip()] = v.strip()
    return out


def main() -> None:
    _configure_stdio()
    try:
        stdin_bytes = sys.stdin.buffer.read()
        req = json.loads(stdin_bytes.decode("utf-8", errors="replace") or "{}")
    except json.JSONDecodeError as e:
        _emit({"ok": False, "error": "invalid stdin json: " + str(e)})
        return

    method = (req.get("method") or "GET").upper()
    uri = _strip_surrogates(req.get("uri") or "")
    cookie = _strip_surrogates(req.get("cookie") or "")
    params = _sanitize_obj(req.get("params"))
    payload = _sanitize_obj(req.get("payload"))
    cookies = cookie_str_to_dict(cookie)

    if not cookies.get("a1"):
        _emit({"ok": False, "error": "cookie 缺少 a1，请从 xiaohongshu.com 重新复制"})
        return

    try:
        from xhshow import Xhshow

        client = Xhshow()
        if method == "GET":
            headers = client.sign_headers_get(uri=uri, cookies=cookies, params=params or {})
        else:
            headers = client.sign_headers_post(uri=uri, cookies=cookies, payload=payload or {})
        headers = _sanitize_obj(headers or {})
        _emit({"ok": True, "headers": headers})
    except ImportError:
        _emit({"ok": False, "error": "请先安装签名库: pip install xhshow"})
    except Exception as e:
        _emit({"ok": False, "error": _strip_surrogates(str(e))})


if __name__ == "__main__":
    main()
