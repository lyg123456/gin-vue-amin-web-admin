/** 顶部导航「视频抓取」下拉项（默认抖音） */
export const VIDEO_CRAWL_NAV_ITEMS = [
  { id: 'douyin', label: '抖音垂类抓取', tool: 'douyin' },
  { id: 'xhs', label: '小红书爆款抓取', tool: 'xhs' }
]

export const VIDEO_CRAWL_TOOL_IDS = new Set(VIDEO_CRAWL_NAV_ITEMS.map((i) => i.tool))

export const DEFAULT_VIDEO_CRAWL_TOOL = 'douyin'
