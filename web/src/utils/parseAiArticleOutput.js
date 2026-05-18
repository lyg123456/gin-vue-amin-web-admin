/**
 * 从 AI 返回的全文里拆分「标题 / 摘要 / SEO / 正文」等区块（常见「标签：内容」格式）
 */

const SECTION_HEADER_RE =
  /^(?:#{1,3}\s*)?(?:【\s*(标题|摘要|SEO\s*标题|SEO\s*描述|SEO\s*关键词|关键词|正文)\s*】|(标题|摘要|SEO\s*标题|SEO\s*描述|SEO\s*关键词|关键词|正文))\s*[：:]?\s*(.*)$/i

function mapSectionKey(label) {
  const s = String(label || '').replace(/\s+/g, '')
  if (s === '标题') return 'title'
  if (s === '摘要') return 'summary'
  if (s === 'SEO标题') return 'seoTitle'
  if (s === 'SEO描述') return 'seoDescription'
  if (s === 'SEO关键词') return 'seoKeywords'
  if (s === '关键词') return 'keywords'
  if (s === '正文') return 'body'
  return null
}

/**
 * @param {string} raw
 * @returns {{ title: string, summary: string, seoTitle: string, seoDescription: string, seoKeywords: string, content: string, hasStructuredSections: boolean }}
 */
export function parseAiArticleOutput(raw) {
  const text = String(raw || '')
    .replace(/\r\n/g, '\n')
    .trim()

  const blocks = {
    title: [],
    summary: [],
    seoTitle: [],
    seoDescription: [],
    seoKeywords: [],
    keywords: [],
    body: [],
    orphan: []
  }

  let current = null
  let buf = []

  const flush = () => {
    if (!current || !buf.length) {
      buf = []
      return
    }
    blocks[current].push(...buf)
    buf = []
  }

  const lines = text.split('\n')
  for (const line of lines) {
    const m = line.match(SECTION_HEADER_RE)
    if (m) {
      flush()
      const label = m[1] || m[2]
      current = mapSectionKey(label)
      const inline = (m[3] || '').trim()
      if (current) {
        buf = inline ? [inline] : []
      } else {
        current = null
        blocks.orphan.push(line)
      }
    } else if (current) {
      buf.push(line)
    } else {
      blocks.orphan.push(line)
    }
  }
  flush()

  const join = (key) => blocks[key].join('\n').trim()

  const hasStructuredSections =
    join('summary') !== '' ||
    join('seoDescription') !== '' ||
    join('seoKeywords') !== '' ||
    join('keywords') !== '' ||
    join('body') !== '' ||
    join('title') !== ''

  let content = join('body')
  if (!content) {
    content = blocks.orphan.join('\n').trim()
  }
  content = stripLeadingArticleTitle(content)

  const seoKeywords = join('seoKeywords') || join('keywords')

  return {
    title: join('title'),
    summary: join('summary'),
    seoTitle: join('seoTitle'),
    seoDescription: join('seoDescription'),
    seoKeywords,
    content,
    hasStructuredSections
  }
}

/** 去掉正文开头与文章标题重复的 Markdown 一级标题行 */
function stripLeadingArticleTitle(content) {
  let s = String(content || '').trim()
  if (!s) return ''
  const lines = s.split('\n')
  while (lines.length) {
    const first = lines[0].trim()
    if (!first) {
      lines.shift()
      continue
    }
    if (/^#\s+/.test(first)) {
      lines.shift()
      continue
    }
    break
  }
  return lines.join('\n').trim()
}

/** 无结构化标签时，从正文提炼短摘要 */
export function buildSummaryFromMarkdown(md) {
  const t = String(md || '')
    .replace(/^#+\s+/gm, '')
    .replace(/[*_`]/g, '')
    .trim()
  if (!t) return ''
  return t.length > 220 ? `${t.slice(0, 220)}…` : t
}
