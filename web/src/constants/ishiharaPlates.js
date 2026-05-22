/** 石原氏色盲检测图 — 图源 Wikimedia Commons；答案按「正常色觉」所见数字标注 */
export const ISHIHARA_WIKI = 'https://commons.wikimedia.org/wiki/Special:FilePath'

export const ishiharaPlates = [
  {
    id: 1,
    file: 'Ishihara_1.PNG',
    local: '/portal/ishihara/01.png',
    title: '第 1 题（对照图）',
    answer: '12',
    options: ['12', '21', '无法辨认', '看不出']
  },
  {
    id: 2,
    file: '47-rg12.jpg',
    local: '/portal/ishihara/02.png',
    title: '第 2 题',
    answer: '17',
    options: ['17', '12', '71', '无法辨认']
  },
  {
    id: 3,
    file: 'Ishihara_3.jpg',
    local: '/portal/ishihara/03.jpg',
    title: '第 3 题',
    answer: '6',
    options: ['6', '29', '5', '无法辨认']
  },
  {
    id: 4,
    file: 'Ishihara_9.png',
    local: '/portal/ishihara/04.png',
    title: '第 4 题',
    answer: '71',
    options: ['71', '42', '17', '无法辨认']
  },
  {
    id: 5,
    file: 'Ishihara_11.PNG',
    local: '/portal/ishihara/05.png',
    title: '第 5 题',
    answer: '6',
    options: ['6', '74', '5', '无法辨认']
  },
  {
    id: 6,
    file: 'Ishihara_19.PNG',
    local: '/portal/ishihara/06.png',
    title: '第 6 题',
    answer: '8',
    options: ['8', '3', '6', '无法辨认']
  },
  {
    id: 7,
    file: 'Ishihara_23.PNG',
    local: '/portal/ishihara/07.png',
    title: '第 7 题',
    answer: '42',
    options: ['42', '2', '45', '无法辨认']
  },
  {
    id: 8,
    file: 'Ishihara_2.svg',
    local: '/portal/ishihara/08.png',
    title: '第 8 题',
    answer: '2',
    options: ['2', '8', '5', '无法辨认']
  }
]

export function plateRemoteUrl(file, width = 400) {
  const enc = encodeURIComponent(file)
  return `${ISHIHARA_WIKI}/${enc}?width=${width}`
}
