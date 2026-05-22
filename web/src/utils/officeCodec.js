import SparkMD5 from 'spark-md5'

export function md5Hex(text) {
  return SparkMD5.hash(text || '')
}

export function base64Encode(text) {
  const bytes = new TextEncoder().encode(text || '')
  let bin = ''
  bytes.forEach((b) => {
    bin += String.fromCharCode(b)
  })
  return btoa(bin)
}

export function base64Decode(b64) {
  const bin = atob(b64 || '')
  const bytes = Uint8Array.from(bin, (c) => c.charCodeAt(0))
  return new TextDecoder().decode(bytes)
}

export function unicodeToChinese(str) {
  return (str || '').replace(/\\u([0-9a-fA-F]{4})/g, (_, h) =>
    String.fromCharCode(parseInt(h, 16))
  )
}

export function chineseToUnicode(str) {
  return (str || '')
    .split('')
    .map((c) => {
      const code = c.charCodeAt(0)
      return code > 127 ? '\\u' + code.toString(16).padStart(4, '0') : c
    })
    .join('')
}

export function toCamel(str) {
  return (str || '')
    .replace(/[-_\s]+(.)?/g, (_, c) => (c ? c.toUpperCase() : ''))
    .replace(/^./, (m) => m.toLowerCase())
}

export function toSnake(str) {
  return (str || '')
    .replace(/([A-Z])/g, '_$1')
    .replace(/[-\s]+/g, '_')
    .toLowerCase()
    .replace(/^_/, '')
}

export function toKebab(str) {
  return toSnake(str).replace(/_/g, '-')
}

export function genUuid() {
  if (typeof crypto !== 'undefined' && crypto.randomUUID) {
    return crypto.randomUUID()
  }
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0
    const v = c === 'x' ? r : (r & 0x3) | 0x8
    return v.toString(16)
  })
}

export function genPassword(len = 16, opts = {}) {
  const lower = 'abcdefghijklmnopqrstuvwxyz'
  const upper = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'
  const num = '0123456789'
  const sym = '!@#$%^&*-_+='
  let pool = lower
  if (opts.upper !== false) pool += upper
  if (opts.num !== false) pool += num
  if (opts.symbol) pool += sym
  let out = ''
  const arr = new Uint32Array(len)
  crypto.getRandomValues(arr)
  for (let i = 0; i < len; i++) out += pool[arr[i] % pool.length]
  return out
}

/** 简易 JS 格式化（缩进） */
export function formatJS(code) {
  let out = ''
  let indent = 0
  const tab = '  '
  const s = (code || '').replace(/\s+/g, ' ').trim()
  for (let i = 0; i < s.length; i++) {
    const ch = s[i]
    if (ch === '{' || ch === '[') {
      out += ch + '\n'
      indent++
      out += tab.repeat(indent)
    } else if (ch === '}' || ch === ']') {
      out += '\n' + tab.repeat(--indent) + ch
    } else if (ch === ';') {
      out += ';\n' + tab.repeat(indent)
    } else if (ch === ',') {
      out += ',\n' + tab.repeat(indent)
    } else {
      out += ch
    }
  }
  return out
}

export function minifyJS(code) {
  return (code || '')
    .replace(/\/\*[\s\S]*?\*\//g, '')
    .replace(/\/\/.*$/gm, '')
    .replace(/\s+/g, ' ')
    .trim()
}

export function formatCSS(css) {
  return (css || '')
    .replace(/\s*{\s*/g, ' {\n  ')
    .replace(/\s*;\s*/g, ';\n  ')
    .replace(/\s*}\s*/g, '\n}\n')
    .replace(/\s*,\s*/g, ', ')
    .trim()
}

export function minifyCSS(css) {
  return (css || '').replace(/\s+/g, ' ').replace(/\s*([{}:;,])\s*/g, '$1').trim()
}
